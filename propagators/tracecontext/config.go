package tracecontext

import (
	"errors"

	"go.opentelemetry.io/otel/trace"
)

// TODO Function doc

// _____________________ With option functions _____________________

func WithHeaderKey(value HeaderKey) configFn {
	return func(conf *config) {
		conf.headerKey = value
	}
}

func WithHeaderValueConverter(headerConv HeaderValueConverterPort) configFn {
	return func(conf *config) {
		conf.headerValueConv = headerConv
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

// _____________________ HeaderKey _____________________

var ErrDuplicatedHeaderKey = errors.New("duplicated header key")

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

type HeaderKey struct {
	// TraceID specifies the key that will be used to store the trace ID.
	// It defaults to DefaultTraceIDHeader.
	TraceID string

	// ParentID specifies the key that will be used to store the parent ID.
	// It defaults to DefaultParentIDHeader.
	ParentID string

	// SampledPriority specifies the key that will be used to store the sampling priority.
	// It defaults to DefaultPriorityHeader.
	SampledPriority string
}

func (obj *HeaderKey) setDefaultIfEmpty() {
	if obj.TraceID == "" {
		obj.TraceID = DefaultTraceIDHeader
	}
	if obj.ParentID == "" {
		obj.ParentID = DefaultParentIDHeader
	}
	if obj.SampledPriority == "" {
		obj.SampledPriority = DefaultPriorityHeader
	}
}

// Validate checks if header keys are valid (no duplication, ...)
func (obj *HeaderKey) Validate() error {
	if obj.TraceID == obj.ParentID || obj.TraceID == obj.SampledPriority || obj.ParentID == obj.SampledPriority {
		return ErrDuplicatedHeaderKey
	}
	return nil
}

// _____________________ Configuration _____________________

func newConfig(cfg ...configFn) (*config, error) {
	var conf = &config{}

	// Apply configurations
	for _, v := range cfg {
		if v != nil {
			v(conf)
		}
	}

	// Apply default value on empty
	conf.applyDefault()

	// Check configuration is valid
	if err := conf.headerKey.Validate(); err != nil {
		return nil, err
	}

	return conf, nil
}

type config struct {
	headerKey       HeaderKey
	headerValueConv HeaderValueConverterPort
}

func (obj *config) applyDefault() {
	// Set default header value
	obj.headerKey.setDefaultIfEmpty()

	// Set default header converter
	if obj.headerValueConv == nil {
		obj.headerValueConv = NewHeaderConvBinary()
	}
}
