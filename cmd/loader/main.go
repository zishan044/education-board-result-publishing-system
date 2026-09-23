package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/zishan044/education-board-result-publishing-system/internal/result"
)

func main() {
	var (
		dbURL       = flag.String("db", os.Getenv("DATABASE_URL"), "Postgres connection string (direct to :5432, not PgBouncer)")
		count       = flag.Int64("count", 100_000, "number of rows to generate")
		seed        = flag.Uint64("seed", 1, "RNG seed (same seed = same dataset)")
		exam        = flag.String("exam", "SSC", "exam name")
		year        = flag.Int("year", 2026, "exam year")
		sampleEvery = flag.Int64("sample-every", 0, "record every Nth key for load tests (0 = auto, ~100K keys)")
		sampleOut   = flag.String("sample-out", "loadtest/keys.json", "where to write the sampled keys")
	)
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("set -db or DATABASE_URL")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	every := *sampleEvery
	if every == 0 {
		every = max(*count/100_000, 1)
	}

	gen := NewGenerator(Config{
		Exam:        *exam,
		Year:        int16(*year),
		Count:       *count,
		Seed:        *seed,
		SampleEvery: every,
	})

	start := time.Now()
	n, err := conn.CopyFrom(ctx, pgx.Identifier{"results"}, resultColumns, gen)
	if err != nil {
		log.Fatalf("copy: %v", err)
	}
	elapsed := time.Since(start)
	log.Printf("loaded %d rows in %s (%.0f rows/s)", n, elapsed.Round(time.Millisecond), float64(n)/elapsed.Seconds())

	if err := writeSample(*sampleOut, gen.Sample()); err != nil {
		log.Fatalf("write sample: %v", err)
	}
	log.Printf("wrote %d sample keys to %s", len(gen.Sample()), *sampleOut)
}

type sampleKey struct {
	Exam         string `json:"exam"`
	Year         int16  `json:"year"`
	Board        string `json:"board"`
	Roll         int32  `json:"roll"`
	Registration int64  `json:"registration"`
}

func writeSample(path string, keys []result.Key) error {
	out := make([]sampleKey, len(keys))
	for i, k := range keys {
		out[i] = sampleKey{k.Exam, k.ExamYear, k.Board, k.Roll, k.Registration}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}