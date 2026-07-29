package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rainforestpay/pgxscan"
)

func NewScannerRows(_ context.Context, rows pgx.Rows) pgxscan.Scanner {
	return newScanner(rows)
}

func NewScannerRow(_ context.Context, rows pgx.Rows) pgxscan.Scanner {
	return pgxscan.NewScanner(rows)
}

func newScanner(rows pgx.Rows) pgxscan.Scanner {
	return &scanner{inner: pgxscan.NewScanner(rows)}
}

type scanner struct {
	inner pgxscan.Scanner
}

func (s *scanner) Scan(v ...any) error {
	if err := s.inner.Scan(v...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}
