package rlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type format string

const (
	formatText format = "text"
	formatJSON format = "json"
)

type Logger struct {
	Inner zerolog.Logger
}

type ILogger interface {
	Debug(msg string)
	Debugf(format string, args ...any)
	Info(msg string)
	Infof(format string, args ...any)
	Warn(msg string)
	Warnf(format string, args ...any)
	Error(msg string)
	Errorf(format string, args ...any)
	AddMetadata(key string, value any) ILogger
}

func (l *Logger) log(level zerolog.Level, msg string) {
	l.Inner.WithLevel(level).Msg(msg)
}

func (l *Logger) Debug(msg string) {
	l.log(zerolog.DebugLevel, msg)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(zerolog.DebugLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(msg string) {
	l.log(zerolog.InfoLevel, msg)
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(zerolog.InfoLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	l.log(zerolog.WarnLevel, msg)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(zerolog.WarnLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(msg string) {
	l.log(zerolog.ErrorLevel, msg)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.log(zerolog.ErrorLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) AddMetadata(key string, value any) ILogger {
	return &Logger{Inner: l.Inner.With().Interface(key, value).Logger()}
}

func Nop() ILogger {
	return &Logger{Inner: zerolog.Nop()}
}

func newLogger(w io.Writer, f format) *Logger {
	switch f {
	case formatText:
		w = zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	}

	return &Logger{
		Inner: zerolog.New(w).With().Timestamp().Logger(),
	}
}

func Initialize[T context.Context](ctx T) T {
	l := newLogger(os.Stdout, formatJSON)
	log.Logger = l.Inner
	return WithLogger(ctx, l)
}

func WithLogger[T context.Context](ctx T, logger ILogger) T {
	return rctx.Set(ctx, rctx.ContextKeyLogger, logger)
}

func GetLogger[T context.Context](ctx T) ILogger {
	ctxLogger, ok := rctx.Get[T, ILogger](ctx, rctx.ContextKeyLogger)
	if !ok {
		return Nop()
	}
	return ctxLogger
}
