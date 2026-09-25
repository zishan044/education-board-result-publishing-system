package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zishan044/education-board-result-publishing-system/internal/result"
)

var ErrNotFound = errors.New("result not found")

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string, maxConns int32) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close()                         { s.pool.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

const getResultSQL = `
SELECT exam, exam_year, board, roll, registration, name,
       division, district, institution_code, gpa, passed, subjects
FROM results
WHERE exam = $1 AND exam_year = $2 AND board = $3
  AND roll = $4 AND registration = $5`

func (s *Store) Get(ctx context.Context, k result.Key) (*result.Result, error) {
	var r result.Result
	err := s.pool.QueryRow(ctx, getResultSQL,
		k.Exam, k.ExamYear, k.Board, k.Roll, k.Registration,
	).Scan(
		&r.Exam, &r.ExamYear, &r.Board, &r.Roll, &r.Registration, &r.Name,
		&r.Division, &r.District, &r.InstitutionCode, &r.GPA, &r.Passed, &r.Subjects,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get result: %w", err)
	}
	return &r, nil
}

func (s *Store) ListAll(ctx context.Context, exam string, year int16, fn func(result.Result) error) error {
	rows, err := s.pool.Query(ctx, `
		SELECT exam, exam_year, board, roll, registration, name,
		       division, district, institution_code, gpa, passed, subjects
		FROM results
		WHERE exam = $1 AND exam_year = $2
		ORDER BY board, roll`, exam, year)
	if err != nil {
		return fmt.Errorf("list all: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r result.Result
		if err := rows.Scan(
			&r.Exam, &r.ExamYear, &r.Board, &r.Roll, &r.Registration, &r.Name,
			&r.Division, &r.District, &r.InstitutionCode, &r.GPA, &r.Passed, &r.Subjects,
		); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		if err := fn(r); err != nil {
			return err
		}
	}
	return rows.Err()
}

const statsSQL = `
SELECT board,
       count(*) AS total,
       count(*) FILTER (WHERE passed) AS passed,
       coalesce(avg(gpa) FILTER (WHERE passed), 0) AS avg_gpa
FROM results
WHERE exam = $1 AND exam_year = $2
GROUP BY board
ORDER BY board`

func (s *Store) Stats(ctx context.Context, exam string, year int16) (*result.Stats, error) {
	rows, err := s.pool.Query(ctx, statsSQL, exam, year)
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	defer rows.Close()

	st := &result.Stats{Exam: exam, ExamYear: year, GeneratedAt: time.Now().UTC()}
	for rows.Next() {
		var b result.BoardStat
		if err := rows.Scan(&b.Board, &b.Total, &b.Passed, &b.AvgGPA); err != nil {
			return nil, fmt.Errorf("scan stats row: %w", err)
		}
		b.Failed = b.Total - b.Passed
		if b.Total > 0 {
			b.PassRate = math.Round(float64(b.Passed)/float64(b.Total)*10000) / 100
		}
		b.AvgGPA = math.Round(b.AvgGPA*100) / 100
		st.ByBoard = append(st.ByBoard, b)
		st.Total += b.Total
		st.Passed += b.Passed
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	st.Failed = st.Total - st.Passed
	if st.Total > 0 {
		st.PassRate = math.Round(float64(st.Passed)/float64(st.Total)*10000) / 100
	}
	var weighted float64
	for _, b := range st.ByBoard {
		weighted += b.AvgGPA * float64(b.Total)
	}
	if st.Total > 0 {
		st.AvgGPA = math.Round(weighted/float64(st.Total)*100) / 100
	}
	return st, nil
}