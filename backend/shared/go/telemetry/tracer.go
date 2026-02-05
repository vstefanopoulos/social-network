package tele

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type tracing struct {
	tracer trace.Tracer
}

func NewTracer(name string) *tracing {
	t := &tracing{
		tracer: otel.Tracer(name),
	}
	return t
}
