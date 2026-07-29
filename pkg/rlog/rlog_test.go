package rlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/ali-gulzar/speechory-core/pkg/rlog"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	return line
}

func Test_Logger_levels(t *testing.T) {
	t.Run("debug_writes_a_debug_level_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf).Level(zerolog.DebugLevel)}

		l.Debug("hello")

		line := decodeLine(t, &buf)
		require.Equal(t, "debug", line["level"])
		require.Equal(t, "hello", line["message"])
	})

	t.Run("debugf_formats_the_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf).Level(zerolog.DebugLevel)}

		l.Debugf("hello %s", "world")

		line := decodeLine(t, &buf)
		require.Equal(t, "debug", line["level"])
		require.Equal(t, "hello world", line["message"])
	})

	t.Run("info_writes_an_info_level_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.Info("hello")

		line := decodeLine(t, &buf)
		require.Equal(t, "info", line["level"])
		require.Equal(t, "hello", line["message"])
	})

	t.Run("infof_formats_the_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.Infof("hello %s", "world")

		line := decodeLine(t, &buf)
		require.Equal(t, "info", line["level"])
		require.Equal(t, "hello world", line["message"])
	})

	t.Run("warn_writes_a_warn_level_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.Warn("hello")

		line := decodeLine(t, &buf)
		require.Equal(t, "warn", line["level"])
		require.Equal(t, "hello", line["message"])
	})

	t.Run("warnf_formats_the_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.Warnf("hello %s", "world")

		line := decodeLine(t, &buf)
		require.Equal(t, "warn", line["level"])
		require.Equal(t, "hello world", line["message"])
	})

	t.Run("error_writes_an_error_level_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.Error("hello")

		line := decodeLine(t, &buf)
		require.Equal(t, "error", line["level"])
		require.Equal(t, "hello", line["message"])
	})

	t.Run("errorf_formats_the_message", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.Errorf("hello %s", "world")

		line := decodeLine(t, &buf)
		require.Equal(t, "error", line["level"])
		require.Equal(t, "hello world", line["message"])
	})

	t.Run("debug_below_the_configured_level_is_a_no_op", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf).Level(zerolog.InfoLevel)}

		l.Debug("hello")

		require.Empty(t, buf.String())
	})
}

func Test_AddMetadata(t *testing.T) {
	t.Run("returns_a_new_logger_with_the_field_attached", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		enriched := l.AddMetadata("request_id", "abc123")
		enriched.Info("hello")

		line := decodeLine(t, &buf)
		require.Equal(t, "abc123", line["request_id"])
		require.Equal(t, "hello", line["message"])
	})

	t.Run("chains_to_attach_multiple_fields", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		l.AddMetadata("method", "GET").AddMetadata("path", "/health").Info("hello")

		line := decodeLine(t, &buf)
		require.Equal(t, "GET", line["method"])
		require.Equal(t, "/health", line["path"])
	})

	t.Run("does_not_mutate_the_receiver", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		_ = l.AddMetadata("request_id", "abc123")
		l.Info("hello")

		line := decodeLine(t, &buf)
		require.NotContains(t, line, "request_id")
	})
}

func Test_Nop(t *testing.T) {
	log := rlog.Nop()

	require.NotPanics(t, func() {
		log.Debug("x")
		log.Debugf("%s", "x")
		log.Info("x")
		log.Infof("%s", "x")
		log.Warn("x")
		log.Warnf("%s", "x")
		log.Error("x")
		log.Errorf("%s", "x")
		log.AddMetadata("k", "v").Info("x")
	})
}

func Test_WithLogger_GetLogger(t *testing.T) {
	t.Run("gin_context_round_trips_the_logger", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		rlog.WithLogger(c, l)

		got := rlog.GetLogger(c)
		got.Info("hello")
		require.Contains(t, buf.String(), "hello")
	})

	t.Run("gin_context_missing_logger_returns_a_nop_logger", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		require.NotPanics(t, func() { rlog.GetLogger(c).Info("hello") })
	})

	t.Run("plain_context_round_trips_the_logger", func(t *testing.T) {
		var buf bytes.Buffer
		l := &rlog.Logger{Inner: zerolog.New(&buf)}

		ctx := rlog.WithLogger[context.Context](context.Background(), l)

		got := rlog.GetLogger(ctx)
		got.Info("hello")
		require.Contains(t, buf.String(), "hello")
	})

	t.Run("plain_context_missing_logger_returns_a_nop_logger", func(t *testing.T) {
		require.NotPanics(t, func() { rlog.GetLogger(context.Background()).Info("hello") })
	})
}

func Test_Initialize(t *testing.T) {
	t.Run("attaches_a_usable_logger_to_the_returned_context", func(t *testing.T) {
		ctx := rlog.Initialize(context.Background())

		require.NotPanics(t, func() { rlog.GetLogger(ctx).Info("hello") })
	})
}
