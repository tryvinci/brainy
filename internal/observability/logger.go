package observability

import (
	"context"
	"log/slog"
	"os"
)

type traceIDKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey{}).(string); ok {
		return v
	}
	return ""
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func LoggerWithTraceID(logger *slog.Logger, ctx context.Context) *slog.Logger {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return logger
	}
	return logger.With("trace_id", traceID)
}
