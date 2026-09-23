package store

import (
	"context"
	"errors"
	"fmt"

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