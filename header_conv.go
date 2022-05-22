package datadog

import (
	"encoding/binary"
	"strconv"
	"strings"
	"unsafe"

	"go.opentelemetry.io/otel/trace"
)

// ____________________ Binary converter ____________________

func newHeaderConvBinary() HeaderConverterPort {
	// Trace and Span are a byte array representing 128-bits or 64-bits value stored
	// Use binary.ByteOrder to transform byte array in/from uint64
	var c uint16 = 1
	if (*[2]byte)(unsafe.Pointer(&c))[0] == 1 {
		return &headerConvBinary{binary.BigEndian}
	} else {
		return &headerConvBinary{binary.LittleEndian}
	}
}

type headerConvBinary struct {
	endian binary.ByteOrder
}

//convert OpenTelemetry 128-bits trace ID to Datadog 64-bits trace IDs
func (obj *headerConvBinary) traceToDatadog(value trace.TraceID) string {
	// Datadog only uses last 64-bits data like for his propagator B3 extractTextMap function:
	// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/textmap.go#L370-L377
	return obj.uint64ByteArrayToString(value[8:])
}

func (obj *headerConvBinary) traceFromDatadog(value string) (traceID trace.TraceID, err error) {
	// Datadog only uses last 64-bits data like for his propagator B3 extractTextMap function:
	// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/textmap.go#L370-L377
	err = obj.uint64StringToByteArray(value, traceID[8:])
	return
}

//convert OpenTelemetry 64-bits span ID to Datadog 64-bits span IDs
func (obj *headerConvBinary) spanToDatadog(value trace.SpanID) string {
	return obj.uint64ByteArrayToString(value[:])
}

func (obj *headerConvBinary) spanFromDatadog(value string) (spanID trace.SpanID, err error) {
	err = obj.uint64StringToByteArray(value, spanID[:])
	return
}

func (obj *headerConvBinary) uint64ByteArrayToString(data []byte) string {
	var id64b = obj.endian.Uint64(data)

	// Format use is strconv.FormatUint(id, 10)
	// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/textmap.go#L246
	return strconv.FormatUint(id64b, 10)
}

func (obj *headerConvBinary) uint64StringToByteArray(value string, dst []byte) error {
	// Datadog uses a special function to transform header value to uint64
	// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/util.go#L64
	id64b, err := parseUint64(value)
	if err != nil {
		return err
	}

	obj.endian.PutUint64(dst, id64b)
	return nil
}

// ___________________________ Convert from header ___________________________

// parseUint64 parses a uint64 from either an unsigned 64 bit base-10 string
// or a signed 64 bit base-10 string representing an unsigned integer
// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/util.go#L64
func parseUint64(str string) (uint64, error) {
	if strings.HasPrefix(str, "-") {
		id, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint64(id), nil
	}
	return strconv.ParseUint(str, 10, 64)
}
