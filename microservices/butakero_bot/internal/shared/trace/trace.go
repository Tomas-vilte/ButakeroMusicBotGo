package trace

import (
	"context"
	"github.com/google/uuid"
)

// GetTraceID obtiene el ID de traza del contexto para logging distribuido
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return "no_trace_id"
}

func WithTraceID(ctx context.Context) context.Context {
	traceID := uuid.New().String()
	return context.WithValue(ctx, "trace_id", traceID)
}

func GenerateTraceID() string {
	return uuid.New().String()
}
