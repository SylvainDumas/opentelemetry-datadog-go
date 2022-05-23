package datadog

import (
	"strings"
)

type propagatorConfigFn func(*propagator)

func WithTraceIDHeader(header string) propagatorConfigFn {
	return func(prop *propagator) {
		prop.headers.traceID = strings.TrimSpace(header)
	}
}

func WithParentIDHeader(header string) propagatorConfigFn {
	return func(prop *propagator) {
		prop.headers.parentID = strings.TrimSpace(header)
	}
}

func WithSampledPriorityHeader(header string) propagatorConfigFn {
	return func(prop *propagator) {
		prop.headers.sampledPriority = strings.TrimSpace(header)
	}
}

func WithHeaderValueConverter(headerConv HeaderValueConverterPort) propagatorConfigFn {
	return func(prop *propagator) {
		prop.headerConv = headerConv
	}
}

// _________________________

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
