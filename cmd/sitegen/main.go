package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zishan044/education-board-result-publishing-system/internal/render"
	"github.com/zishan044/education-board-result-publishing-system/internal/result"
	"github.com/zishan044/education-board-result-publishing-system/internal/store"
)

func main() {
	var (
		dbURL     = flag.String("db", os.Getenv("DATABASE_URL"), "Postgres connection string")
		out       = flag.String("out", "./site", "output directory for the static site")
		exam      = flag.String("exam", "SSC", "exam name")
		year      = flag.Int("year", 2026, "exam year")
		workers   = flag.Int("workers", runtime.NumCPU()*4, "concurrent render workers")
		boardsCSV = flag.String("boards", "Dhaka,Chattogram,Rajshahi,Jashore,Barishal,Sylhet,Dinajpur,Mymensingh", "boards to list in the search form")
	)
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("set -db or DATABASE_URL")
	}

	ctx := context.Background()
	st, err := store.New(ctx, *dbURL, int32(*workers))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer st.Close()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("mkdir out: %v", err)
	}

	start := time.Now()
	rowsCh := make(chan result.Result, 1000)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var produced int64

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range rowsCh {
				if err := renderOne(*out, r); err != nil {
					log.Printf("render %s-%d: %v", r.Board, r.Roll, err)
					continue
				}
				mu.Lock()
				produced++
				n := produced
				mu.Unlock()
				if n%100_000 == 0 {
					log.Printf("rendered %d pages", n)
				}
			}
		}()
	}

	listErr := st.ListAll(ctx, *exam, int16(*year), func(r result.Result) error {
		rowsCh <- r
		return nil
	})
	close(rowsCh)
	wg.Wait()
	if listErr != nil {
		log.Fatalf("list all: %v", listErr)
	}
	log.Printf("rendered %d result pages in %s", produced, time.Since(start).Round(time.Second))

	stats, err := st.Stats(ctx, *exam, int16(*year))
	if err != nil {
		log.Fatalf("stats: %v", err)
	}
	if err := renderStats(*out, stats); err != nil {
		log.Fatalf("render stats: %v", err)
	}
	log.Print("rendered stats.html")

	boards := strings.Split(*boardsCSV, ",")
	if err := renderIndex(*out, boards, []int16{int16(*year)}); err != nil {
		log.Fatalf("render index: %v", err)
	}
	log.Print("rendered index.html")
}

func renderOne(outRoot string, r result.Result) error {
	full := filepath.Join(outRoot, "results", r.Key().ResultPath())
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	tmp := full + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := render.Result(f, &r); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, full)
}

func renderStats(outRoot string, s *result.Stats) error {
	f, err := os.Create(filepath.Join(outRoot, "stats.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return render.Stats(f, s)
}

func renderIndex(outRoot string, boards []string, years []int16) error {
	f, err := os.Create(filepath.Join(outRoot, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return render.Index(f, render.IndexData{Boards: boards, Years: years})
}