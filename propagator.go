package datadog

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	errMalformedTraceID = errors.New("cannot parse Datadog trace ID as 64bit unsigned int from header")
	errMalformedSpanID  = errors.New("cannot parse Datadog span ID as 64bit unsigned int from header")
)

type propagatorConfigFn func(*propagator)

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
	if newPropagator.datadogCfg.BaggagePrefix == "" {
		newPropagator.datadogCfg.BaggagePrefix = tracer.DefaultBaggageHeaderPrefix
	}
	if newPropagator.datadogCfg.TraceHeader == "" {
		newPropagator.datadogCfg.TraceHeader = tracer.DefaultTraceIDHeader
	}
	if newPropagator.datadogCfg.ParentHeader == "" {
		newPropagator.datadogCfg.ParentHeader = tracer.DefaultParentIDHeader
	}
	if newPropagator.datadogCfg.PriorityHeader == "" {
		newPropagator.datadogCfg.PriorityHeader = tracer.DefaultPriorityHeader
	}

	// Set default header converter
	if newPropagator.headerConv == nil {
		newPropagator.headerConv = newHeaderConvBinary()
	}

	return newPropagator
}

type HeaderConverterPort interface {
	// Trace
	traceToDatadog(value trace.TraceID) string
	traceFromDatadog(value string) (trace.TraceID, error)
	// Span
	spanToDatadog(value trace.SpanID) string
	spanFromDatadog(value string) (trace.SpanID, error)
}

// Propagator serializes Span Context to/from Datadog headers.
type propagator struct {
	datadogCfg tracer.PropagatorConfig
	headerConv HeaderConverterPort
}

// Inject injects a context to the carrier following Datadog format.
func (obj *propagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	// If no Span Context or invalid, do not inject it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return
	}

	// Inject Trace ID, Span ID, Sampled in carrier
	carrier.Set(obj.datadogCfg.TraceHeader, obj.headerConv.traceToDatadog(spanCtx.TraceID()))
	carrier.Set(obj.datadogCfg.ParentHeader, obj.headerConv.spanToDatadog(spanCtx.SpanID()))
	carrier.Set(obj.datadogCfg.PriorityHeader, otelToSampledDatadogHeader(spanCtx.TraceFlags()))
}

// Extract gets a context from the carrier if it contains Datadog headers.
func (obj *propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	// If an Span Context already defined, do not override it
	var spanCtx = trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return ctx
	}

	var (
		traceID = carrier.Get(obj.datadogCfg.TraceHeader)
		spanID  = carrier.Get(obj.datadogCfg.ParentHeader)
		sampled = carrier.Get(obj.datadogCfg.PriorityHeader)
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
		obj.datadogCfg.TraceHeader,
		obj.datadogCfg.ParentHeader,
		obj.datadogCfg.PriorityHeader,
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
