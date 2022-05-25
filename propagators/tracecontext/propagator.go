package tracecontext

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	errMalformedTraceID = errors.New("cannot parse Datadog trace ID as 64bit unsigned int from header")
	errMalformedSpanID  = errors.New("cannot parse Datadog span ID as 64bit unsigned int from header")
)

// NewPropagator returns a new propagator which uses TextMap to inject
// and extract values. It propagates trace and span IDs and baggage.
// To use the defaults, nil may be provided in place of the config.
func NewPropagator(cfg ...configFn) propagation.TextMapPropagator {
	return &propagator{
		conf: newConfig(cfg...),
	}
}

// propagator serializes Span Context to/from Datadog headers.
type propagator struct {
	conf *config
}

// Inject injects a context to the carrier following Datadog format.
func (obj *propagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	// If no Span Context or invalid, do not inject it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return
	}

	// Inject Trace ID, Span ID, Sampled in carrier
	carrier.Set(obj.conf.headers.traceID, obj.conf.headerConv.traceToDatadog(spanCtx.TraceID()))
	carrier.Set(obj.conf.headers.parentID, obj.conf.headerConv.spanToDatadog(spanCtx.SpanID()))
	carrier.Set(obj.conf.headers.sampledPriority, otelToSampledDatadogHeader(spanCtx.TraceFlags()))
}

// Extract gets a context from the carrier if it contains Datadog headers.
func (obj *propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	// If an Span Context already defined, do not override it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return ctx
	}

	var (
		traceID = carrier.Get(obj.conf.headers.traceID)
		spanID  = carrier.Get(obj.conf.headers.parentID)
		sampled = carrier.Get(obj.conf.headers.sampledPriority)
	)
	sc, err := obj.extract(traceID, spanID, sampled)
	if err != nil || !sc.IsValid() {
		return ctx
	}

	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (obj *propagator) extract(traceID, spanID, sampled string) (trace.SpanContext, error) {
	var (
		scc trace.SpanContextConfig
		err error
	)

	if scc.TraceID, err = obj.conf.headerConv.traceFromDatadog(traceID); err != nil {
		return trace.SpanContext{}, errMalformedTraceID
	}

	if scc.SpanID, err = obj.conf.headerConv.spanFromDatadog(spanID); err != nil {
		return trace.SpanContext{}, errMalformedSpanID
	}

	scc.TraceFlags = scc.TraceFlags.WithSampled(sampled == datadogHeaderSampled)

	return trace.NewSpanContext(scc), nil
}

// Fields returns the keys whose values are set with Inject.
func (obj *propagator) Fields() []string {
	return []string{
		obj.conf.headers.traceID,
		obj.conf.headers.parentID,
		obj.conf.headers.sampledPriority,
	}
}

// ________________ sampling ________________

const (
	datadogHeaderNotSampled = "0"
	datadogHeaderSampled    = "1"
)

func otelToSampledDatadogHeader(value trace.TraceFlags) string {
	if value.IsSampled() {
		return datadogHeaderSampled
	}
	return datadogHeaderNotSampled
}
