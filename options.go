package datadog

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func WithDatadogHeader(cfg tracer.PropagatorConfig) propagatorConfigFn {
	return func(opts *propagator) {
		opts.datadogCfg = cfg
	}
}

func WithHeaderConverter(headerConv HeaderConverterPort) propagatorConfigFn {
	return func(opts *propagator) {
		opts.headerConv = headerConv
	}
}
