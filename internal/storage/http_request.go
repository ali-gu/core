package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type HTTPRequest struct {
	ID           ksuid.KSUID  `db:"id" json:"id"`
	PracticeID   *ksuid.KSUID `db:"practice_id" json:"practice_id"`
	UserID       *ksuid.KSUID `db:"user_id" json:"user_id"`
	Method       string       `db:"method" json:"method"`
	Path         string       `db:"path" json:"path"`
	QueryParams  string       `db:"query_params" json:"query_params"`
	Headers      string       `db:"headers" json:"headers"`
	IPAddress    string       `db:"ip_address" json:"ip_address"`
	RequestBody  string       `db:"request_body" json:"request_body"`
	ResponseBody string       `db:"response_body" json:"response_body"`
	StatusCode   int          `db:"status_code" json:"status_code"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}

type HTTPRequestFilters struct {
	PracticeID *ksuid.KSUID
}

type IHTTPRequestStorage interface {
	Create(ctx context.Context, db DB, req HTTPRequest) error
	Get(ctx context.Context, db DB, filters HTTPRequestFilters) ([]HTTPRequest, error)
}

type HTTPRequestStorage struct{}

var _ IHTTPRequestStorage = (*HTTPRequestStorage)(nil)

func (s *HTTPRequestStorage) Create(ctx context.Context, db DB, req HTTPRequest) error {
	sql, args, err := StatementBuilder.Insert("http_requests").SetMap(map[string]any{
		"id":            req.ID,
		"practice_id":   req.PracticeID,
		"user_id":       req.UserID,
		"method":        req.Method,
		"path":          req.Path,
		"query_params":  req.QueryParams,
		"headers":       req.Headers,
		"ip_address":    req.IPAddress,
		"request_body":  req.RequestBody,
		"response_body": req.ResponseBody,
		"status_code":   req.StatusCode,
		"created_at":    req.CreatedAt,
	}).ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *HTTPRequestStorage) Get(ctx context.Context, db DB, filters HTTPRequestFilters) ([]HTTPRequest, error) {
	builder := StatementBuilder.Select("*").From("http_requests").OrderBy("created_at DESC")

	if filters.PracticeID != nil {
		builder = builder.Where(squirrel.Eq{"practice_id": *filters.PracticeID})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	reqs, err := pgx.CollectRows(rows, pgx.RowToStructByName[HTTPRequest])
	if err != nil {
		return nil, rerror.New(err)
	}
	return reqs, nil
}
