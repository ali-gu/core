package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/ali-gulzar/speechory-core/pkg/rlog"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

const maxBodyCapture = 64 * 1024

var sensitiveFields = map[string]bool{
	"password":       true,
	"secret":         true,
	"token":          true,
	"access_token":   true,
	"refresh_token":  true,
	"authorization":  true,
	"credit_card":    true,
	"card_number":    true,
	"cvv":            true,
	"ssn":            true,
	"api_key":        true,
	"apikey":         true,
}

var sensitiveHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"set-cookie":    true,
	"x-api-key":     true,
	"proxy-authorization": true,
}

func HTTPRequestLogger(dbMux storage.IDBMux, httpReqStorage storage.IHTTPRequestStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var reqBody string
		if ctx.Request.Body != nil {
			buf, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxBodyCapture+1))
			if err == nil {
				if int64(len(buf)) > maxBodyCapture {
					buf = buf[:maxBodyCapture]
				}
				reqBody = compactAndMaskBody(buf)
				ctx.Request.Body = io.NopCloser(bytes.NewReader(buf))
			}
		}

		capture := &responseCapture{ResponseWriter: ctx.Writer, status: http.StatusOK}
		ctx.Writer = capture

		ctx.Next()

		var practiceID *ksuid.KSUID
		var userID *ksuid.KSUID
		if user, ok := rctx.GetAuthenticatedUser(ctx); ok {
			pid := user.PracticeID
			uid := user.ID
			practiceID = &pid
			userID = &uid
		}

		path := ctx.FullPath()
		if path == "" {
			path = ctx.Request.URL.Path
		}

		record := storage.HTTPRequest{
			ID:           ksuid.New(),
			PracticeID:   practiceID,
			UserID:       userID,
			Method:       ctx.Request.Method,
			Path:         path,
			QueryParams:  compactQueryParams(ctx.Request.URL),
			Headers:      maskHeaders(ctx.Request.Header),
			IPAddress:    ctx.ClientIP(),
			RequestBody:  reqBody,
			ResponseBody: compactAndMaskBody(capture.body.Bytes()),
			StatusCode:   capture.status,
			CreatedAt:    time.Now(),
		}

		db, err := dbMux.BeginWrite()
		if err != nil {
			rlog.GetLogger(ctx).Errorf("http_request_log: begin write: %s", err)
			return
		}

		if err := httpReqStorage.Create(ctx, db, record); err != nil {
			rlog.GetLogger(ctx).Errorf("http_request_log: create: %s", err)
		}
	}
}

func compactQueryParams(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return u.RawQuery
	}
	compacted, err := json.Marshal(values)
	if err != nil {
		return u.RawQuery
	}
	return string(compacted)
}

func maskHeaders(header http.Header) string {
	masked := make(map[string]any, len(header))
	for key, values := range header {
		if sensitiveHeaders[strings.ToLower(key)] {
			masked[key] = "***"
		} else if len(values) == 1 {
			masked[key] = values[0]
		} else {
			masked[key] = values
		}
	}
	raw, err := json.Marshal(masked)
	if err != nil {
		return ""
	}
	return string(raw)
}

func compactAndMaskBody(raw []byte) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ""
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return string(raw)
	}

	maskJSON(parsed)

	compacted, err := json.Marshal(parsed)
	if err != nil {
		return string(raw)
	}
	return string(compacted)
}

func maskJSON(v any) {
	switch val := v.(type) {
	case map[string]any:
		for key, child := range val {
			if sensitiveFields[strings.ToLower(key)] {
				val[key] = "***"
			} else {
				maskJSON(child)
			}
		}
	case []any:
		for _, child := range val {
			maskJSON(child)
		}
	}
}

type responseCapture struct {
	gin.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *responseCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseCapture) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
