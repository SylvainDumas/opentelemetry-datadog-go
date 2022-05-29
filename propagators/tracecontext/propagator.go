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

// NewDefault returns a new propagator with default configuration which uses
// TextMap to inject and extract values. It propagates trace and span IDs.
func NewDefault() propagation.TextMapPropagator {
	prop, err := New(nil)
	if err != nil {
		return nil
	}
	return prop
}

// New returns a new propagator which uses TextMap to inject and extract
// values. It propagates trace and span IDs.
// To use the defaults, call with nothing.
func New(cfg ...configFn) (propagation.TextMapPropagator, error) {
	propagatorConf, err := newConfig(cfg...)
	if err != nil {
		return nil, err
	}

	return &propagator{conf: propagatorConf}, nil
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
	carrier.Set(obj.conf.headerKey.TraceID, obj.conf.headerValueConv.traceToDatadog(spanCtx.TraceID()))
	carrier.Set(obj.conf.headerKey.ParentID, obj.conf.headerValueConv.spanToDatadog(spanCtx.SpanID()))
	carrier.Set(obj.conf.headerKey.SampledPriority, otelToSampledDatadogHeader(spanCtx.TraceFlags()))
}

// Extract gets a context from the carrier if it contains Datadog headers.
func (obj *propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	// If an Span Context already defined, do not override it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return ctx
	}

	var (
		traceID = carrier.Get(obj.conf.headerKey.TraceID)
		spanID  = carrier.Get(obj.conf.headerKey.ParentID)
		sampled = carrier.Get(obj.conf.headerKey.SampledPriority)
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

	if scc.TraceID, err = obj.conf.headerValueConv.traceFromDatadog(traceID); err != nil {
		return trace.SpanContext{}, errMalformedTraceID
	}

	if scc.SpanID, err = obj.conf.headerValueConv.spanFromDatadog(spanID); err != nil {
		return trace.SpanContext{}, errMalformedSpanID
	}

	scc.TraceFlags = scc.TraceFlags.WithSampled(sampled == datadogHeaderSampled)

	return trace.NewSpanContext(scc), nil
}

// Fields returns the keys whose values are set with Inject.
func (obj *propagator) Fields() []string {
	return []string{
		obj.conf.headerKey.TraceID,
		obj.conf.headerKey.ParentID,
		obj.conf.headerKey.SampledPriority,
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
