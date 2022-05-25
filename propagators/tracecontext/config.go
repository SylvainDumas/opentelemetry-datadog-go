package tracecontext

import (
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// _____________________ With helpers _____________________

func WithTraceIDHeader(header string) configFn {
	return func(conf *config) {
		conf.headers.traceID = strings.TrimSpace(header)
	}
}

func WithParentIDHeader(header string) configFn {
	return func(conf *config) {
		conf.headers.parentID = strings.TrimSpace(header)
	}
}

func WithSampledPriorityHeader(header string) configFn {
	return func(conf *config) {
		conf.headers.sampledPriority = strings.TrimSpace(header)
	}
}

func WithHeaderValueConverter(headerConv HeaderValueConverterPort) configFn {
	return func(conf *config) {
		conf.headerConv = headerConv
	}
}

// _____________________ Definition _____________________

type configFn func(*config)

type HeaderValueConverterPort interface {
	// Trace
	traceToDatadog(value trace.TraceID) string
	traceFromDatadog(value string) (trace.TraceID, error)
	// Span
	spanToDatadog(value trace.SpanID) string
	spanFromDatadog(value string) (trace.SpanID, error)
}

// _____________________ Configuration _____________________

// Ref https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/tracer/textmap.go#L73-L89

const (
	// DefaultTraceIDHeader specifies the key that will be used in HTTP headers
	// or text maps to store the trace ID.
	DefaultTraceIDHeader = "x-datadog-trace-id"

	// DefaultParentIDHeader specifies the key that will be used in HTTP headers
	// or text maps to store the parent ID.
	DefaultParentIDHeader = "x-datadog-parent-id"

	// DefaultPriorityHeader specifies the key that will be used in HTTP headers
	// or text maps to store the sampling priority value.
	DefaultPriorityHeader = "x-datadog-sampling-priority"
)

func newConfig(cfg ...configFn) *config {
	var conf = &config{}

	// Apply configurations
	for _, v := range cfg {
		if v != nil {
			v(conf)
		}
	}

	conf.applyDefault()

	return conf
}

type config struct {
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

func (obj *config) applyDefault() {
	// Set default header value
	if obj.headers.traceID == "" {
		obj.headers.traceID = DefaultTraceIDHeader
	}
	if obj.headers.parentID == "" {
		obj.headers.parentID = DefaultParentIDHeader
	}
	if obj.headers.sampledPriority == "" {
		obj.headers.sampledPriority = DefaultPriorityHeader
	}

	// Set default header converter
	if obj.headerConv == nil {
		obj.headerConv = NewHeaderConvBinary()
	}
}
