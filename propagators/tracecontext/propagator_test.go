package tracecontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func Test_propagator_NewDefault(t *testing.T) {
	assert.NotNil(t, NewDefault())
}

func Test_propagator_New(t *testing.T) {
	_, err := New()
	assert.NoError(t, err)

	_, err = New(WithHeaderKey(HeaderKey{"a", "a", ""}))
	assert.ErrorIs(t, err, ErrDuplicatedHeaderKey)
}

func Test_propagator_Inject(t *testing.T) {
	prop, err := New()
	require.NoError(t, err)

	var carrier = propagation.MapCarrier{}

	// Check invalid span -> empty carrier
	prop.Inject(context.Background(), carrier)
	assert.Empty(t, carrier)

	// Create Span Context
	var traceFlag trace.TraceFlags
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d},
		SpanID:     trace.SpanID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61},
		TraceFlags: traceFlag.WithSampled(true),
	})
	require.True(t, sc.IsValid())

	// Check well injected
	prop.Inject(trace.ContextWithSpanContext(context.Background(), sc), carrier)
	assert.Equal(t, "16701352862047361693", carrier.Get(DefaultTraceIDHeader))
	assert.Equal(t, "13263342393987690081", carrier.Get(DefaultParentIDHeader))
	assert.Equal(t, datadogHeaderSampled, carrier.Get(DefaultPriorityHeader))
}

func Test_propagator_Extract(t *testing.T) {
	prop, err := New()
	require.NoError(t, err)

	var carrier = propagation.MapCarrier{}
	var ctx = context.Background()

	// Check no span in context
	var extractedSc = trace.SpanContextFromContext(prop.Extract(ctx, carrier))
	assert.False(t, extractedSc.IsValid())

	// Insert partial Datadog Span Context in carrier
	carrier.Set(DefaultTraceIDHeader, "16701352862047361693")
	// Check no span in context
	extractedSc = trace.SpanContextFromContext(prop.Extract(ctx, carrier))
	assert.False(t, extractedSc.IsValid())

	// Insert Datadog Span Context in carrier
	carrier.Set(DefaultParentIDHeader, "13263342393987690081")
	carrier.Set(DefaultPriorityHeader, datadogHeaderSampled)
	//
	extractedSc = trace.SpanContextFromContext(prop.Extract(ctx, carrier))
	assert.Equal(t, extractedSc.TraceID(), trace.TraceID(trace.TraceID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))
	assert.Equal(t, extractedSc.SpanID(), trace.SpanID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61})
	assert.True(t, extractedSc.TraceFlags().IsSampled())

	// Insert new Datadog Span Context in carrier
	carrier.Set(DefaultTraceIDHeader, "1")
	carrier.Set(DefaultParentIDHeader, "2")
	carrier.Set(DefaultPriorityHeader, datadogHeaderNotSampled)
	// Check if Span Context already exists no override
	var ctxWithSC = trace.ContextWithSpanContext(ctx, extractedSc)
	extractedSc = trace.SpanContextFromContext(prop.Extract(ctxWithSC, carrier))
	assert.Equal(t, extractedSc.TraceID(), trace.TraceID(trace.TraceID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))
	assert.Equal(t, extractedSc.SpanID(), trace.SpanID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61})
	assert.True(t, extractedSc.TraceFlags().IsSampled())
}

func Test_propagator_Fields(t *testing.T) {
	if prop, err := New(); assert.NoError(t, err) {
		assert.ElementsMatch(t,
			prop.Fields(),
			[]string{DefaultParentIDHeader, DefaultPriorityHeader, DefaultTraceIDHeader},
		)
	}
}

func Test_otelToSampledDatadogHeader(t *testing.T) {
	var value trace.TraceFlags
	assert.Equal(t, datadogHeaderNotSampled, otelToSampledDatadogHeader(value))
	assert.Equal(t, datadogHeaderSampled, otelToSampledDatadogHeader(value.WithSampled(true)))
}
