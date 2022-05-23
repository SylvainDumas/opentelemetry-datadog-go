package datadog

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
func NewPropagator(cfg ...propagatorConfigFn) propagation.TextMapPropagator {
	var newPropagator = &propagator{}

	// Apply configuration to new propagator
	for _, v := range cfg {
		if v != nil {
			v(newPropagator)
		}
	}

	// Set default header value
	if newPropagator.headers.traceID == "" {
		newPropagator.headers.traceID = DefaultTraceIDHeader
	}
	if newPropagator.headers.parentID == "" {
		newPropagator.headers.parentID = DefaultParentIDHeader
	}
	if newPropagator.headers.sampledPriority == "" {
		newPropagator.headers.sampledPriority = DefaultPriorityHeader
	}

	// Set default header converter
	if newPropagator.headerConv == nil {
		newPropagator.headerConv = newHeaderConvBinary()
	}

	return newPropagator
}

type HeaderValueConverterPort interface {
	// Trace
	traceToDatadog(value trace.TraceID) string
	traceFromDatadog(value string) (trace.TraceID, error)
	// Span
	spanToDatadog(value trace.SpanID) string
	spanFromDatadog(value string) (trace.SpanID, error)
}

// Propagator serializes Span Context to/from Datadog headers.
type propagator struct {
	headers struct {
		// traceID specifies the map key that will be used to store the trace ID.
		// It defaults to DefaultTraceIDHeader.
		traceID string

		// parentID specifies the map key that will be used to store the parent ID.
		// It defaults to DefaultParentIDHeader.
		parentID string

		// sampledPriority specifies the map key that will be used to store the sampling priority.
		// It defaults to DefaultPriorityHeader.
		sampledPriority string
	}

	headerConv HeaderValueConverterPort
}

// Inject injects a context to the carrier following Datadog format.
func (obj *propagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	// If no Span Context or invalid, do not inject it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return
	}

	// Inject Trace ID, Span ID, Sampled in carrier
	carrier.Set(obj.headers.traceID, obj.headerConv.traceToDatadog(spanCtx.TraceID()))
	carrier.Set(obj.headers.parentID, obj.headerConv.spanToDatadog(spanCtx.SpanID()))
	carrier.Set(obj.headers.sampledPriority, otelToSampledDatadogHeader(spanCtx.TraceFlags()))
}

// Extract gets a context from the carrier if it contains Datadog headers.
func (obj *propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	// If an Span Context already defined, do not override it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return ctx
	}

	var (
		traceID = carrier.Get(obj.headers.traceID)
		spanID  = carrier.Get(obj.headers.parentID)
		sampled = carrier.Get(obj.headers.sampledPriority)
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

	if scc.TraceID, err = obj.headerConv.traceFromDatadog(traceID); err != nil {
		return trace.SpanContext{}, errMalformedTraceID
	}

	if scc.SpanID, err = obj.headerConv.spanFromDatadog(spanID); err != nil {
		return trace.SpanContext{}, errMalformedSpanID
	}

	scc.TraceFlags = scc.TraceFlags.WithSampled(sampled == datadogHeaderSampled)

	return trace.NewSpanContext(scc), nil
}

// Fields returns the keys whose values are set with Inject.
func (obj *propagator) Fields() []string {
	return []string{
		obj.headers.traceID,
		obj.headers.parentID,
		obj.headers.sampledPriority,
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
