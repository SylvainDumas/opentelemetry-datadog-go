package tracecontext

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

//var traceID = trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
//var traceID = trace.TraceID{0x80, 0xf1, 0x98, 0xee, 0x56, 0x34, 0x3b, 0xa8, 0x64, 0xfe, 0x8b, 0x2a, 0x57, 0xd3, 0xef, 0xf7}

func Test_headerConvBinary_traceToDatadog(t *testing.T) {
	var headerConv = NewHeaderConvBinary()

	assert.Equal(t, "16701352862047361693",
		headerConv.traceToDatadog(trace.TraceID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))

	assert.Equal(t, "16701352862047361693",
		headerConv.traceToDatadog(trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))

	assert.Equal(t, "0",
		headerConv.traceToDatadog(trace.TraceID{}))
}

func Test_headerConvBinary_traceFromDatadog(t *testing.T) {
	var headerConv = NewHeaderConvBinary()

	if traceID, err := headerConv.traceFromDatadog("16701352862047361693"); assert.NoError(t, err) {
		assert.Equal(t, trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}, traceID)
	}

	if traceID, err := headerConv.traceFromDatadog("0"); assert.NoError(t, err) {
		assert.Equal(t, trace.TraceID{}, traceID)
	}

	_, err := headerConv.traceFromDatadog("abc")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "abc": invalid syntax`)

	_, err = headerConv.traceFromDatadog("")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "": invalid syntax`)
}

func Test_headerConvBinary_spanToDatadog(t *testing.T) {
	var headerConv = NewHeaderConvBinary()

	assert.Equal(t, "16701352862047361693",
		headerConv.spanToDatadog(trace.SpanID{0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))

	assert.Equal(t, "0",
		headerConv.spanToDatadog(trace.SpanID{}))
}

func Test_headerConvBinary_spanFromDatadog(t *testing.T) {
	var headerConv = NewHeaderConvBinary()

	if spanID, err := headerConv.spanFromDatadog("16701352862047361693"); assert.NoError(t, err) {
		assert.Equal(t, trace.SpanID{0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}, spanID)
	}

	if spanID, err := headerConv.spanFromDatadog("0"); assert.NoError(t, err) {
		assert.Equal(t, trace.SpanID{}, spanID)
	}

	_, err := headerConv.spanFromDatadog("abc")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "abc": invalid syntax`)

	_, err = headerConv.spanFromDatadog("")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "": invalid syntax`)
}

func Test_headerConvString_traceToDatadog(t *testing.T) {
	var headerConv = NewHeaderConvString()

	assert.Equal(t, "16701352862047361693",
		headerConv.traceToDatadog(trace.TraceID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))

	assert.Equal(t, "16701352862047361693",
		headerConv.traceToDatadog(trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))

	assert.Equal(t, "0",
		headerConv.traceToDatadog(trace.TraceID{}))
}

func Test_headerConvString_traceFromDatadog(t *testing.T) {
	var headerConv = NewHeaderConvString()

	if traceID, err := headerConv.traceFromDatadog("16701352862047361693"); assert.NoError(t, err) {
		assert.Equal(t, trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}, traceID)
	}

	_, err := headerConv.traceFromDatadog("0")
	assert.ErrorContains(t, err, "trace-id can't be all zero")

	_, err = headerConv.traceFromDatadog("abc")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "abc": invalid syntax`)

	_, err = headerConv.traceFromDatadog("")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "": invalid syntax`)
}

func Test_headerConvString_spanToDatadog(t *testing.T) {
	var headerConv = NewHeaderConvString()

	assert.Equal(t, "16701352862047361693",
		headerConv.spanToDatadog(trace.SpanID{0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}))

	assert.Equal(t, "0",
		headerConv.spanToDatadog(trace.SpanID{}))
}

func Test_headerConvString_spanFromDatadog(t *testing.T) {
	var headerConv = NewHeaderConvString()

	if spanID, err := headerConv.spanFromDatadog("16701352862047361693"); assert.NoError(t, err) {
		assert.Equal(t, trace.SpanID{0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}, spanID)
	}

	_, err := headerConv.spanFromDatadog("0")
	assert.ErrorContains(t, err, "span-id can't be all zero")

	_, err = headerConv.spanFromDatadog("abc")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "abc": invalid syntax`)

	_, err = headerConv.spanFromDatadog("")
	assert.ErrorContains(t, err, `strconv.ParseUint: parsing "": invalid syntax`)
}

// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/util_test.go#L55-L72
func TestParseUint64(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		id, err := parseUint64("-8809075535603237910")
		assert.NoError(t, err)
		assert.Equal(t, uint64(9637668538106313706), id)
	})

	t.Run("negative invalid", func(t *testing.T) {
		_, err := parseUint64("-abcd")
		assert.Error(t, err)
	})

	t.Run("positive", func(t *testing.T) {
		id, err := parseUint64(fmt.Sprintf("%d", uint64(math.MaxUint64)))
		assert.NoError(t, err)
		assert.Equal(t, uint64(math.MaxUint64), id)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseUint64("abcd")
		assert.Error(t, err)
	})
}

func Benchmark_Conv(b *testing.B) {
	var traceID = trace.TraceID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}

	var stringconv = func(trace.TraceID) (string, error) {
		valueStringHex := traceID.String()[16:]
		valueInt, err := strconv.ParseUint(valueStringHex, 16, 64)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(valueInt, 10), nil
	}

	var headerConv = NewHeaderConvBinary()

	b.Run("stringconv", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stringconv(traceID)
		}
	})

	b.Run("convertTraceOTtoDDHeaderValue", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			headerConv.traceToDatadog(traceID)
		}
	})
}
