package datadog

import (
	"bytes"
	"strconv"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func Test_NewParse(t *testing.T) {
	//var traceID = trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	//var traceID = trace.TraceID{0x80, 0xf1, 0x98, 0xee, 0x56, 0x34, 0x3b, 0xa8, 0x64, 0xfe, 0x8b, 0x2a, 0x57, 0xd3, 0xef, 0xf7}
	var traceID = trace.TraceID{0xb8, 0x10, 0xdb, 0xa2, 0x98, 0x03, 0xee, 0x61, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}
	//16701352862047361693

	var stringconv = func(trace.TraceID) (string, error) {
		valueStringHex := traceID.String()[16:]
		valueInt, err := strconv.ParseUint(valueStringHex, 16, 64)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(valueInt, 10), nil
	}

	var headerConv = newHeaderConvBinary()

	var header = headerConv.traceToDatadog(traceID)
	if expected, err := stringconv(traceID); err != nil {
		t.Error(err)
	} else if expected != header {
		t.Errorf("Excpected %+v, get %+v", expected, header)
	}

	traceID2, err := headerConv.traceFromDatadog(header)
	if err != nil {
		t.Error(err)
	}

	//var expectedTraceID = trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	var expectedTraceID = trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0xe7, 0xc7, 0x1f, 0xf0, 0xc2, 0xc9, 0x5a, 0x9d}
	if bytes.Equal(traceID2[:], expectedTraceID[:]) == false {
		t.Errorf("Excpected %+v, get %+v", expectedTraceID, traceID2)
	}
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

	var headerConv = newHeaderConvBinary()

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
