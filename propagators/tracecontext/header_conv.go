package tracecontext

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"go.opentelemetry.io/otel/trace"
)

// ____________________ Binary converter ____________________

func NewHeaderConvBinary() HeaderValueConverterPort {
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

func (obj *headerConvBinary) traceToDatadog(value trace.TraceID) string {
	// Convert OpenTelemetry 128-bits trace ID to Datadog 64-bits trace IDs
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

func (obj *headerConvBinary) spanToDatadog(value trace.SpanID) string {
	// Convert OpenTelemetry 64-bits span ID to Datadog 64-bits span IDs
	return obj.uint64ByteArrayToString(value[:])
}

func (obj *headerConvBinary) spanFromDatadog(value string) (spanID trace.SpanID, err error) {
	err = obj.uint64StringToByteArray(value, spanID[:])
	return
}

func (obj *headerConvBinary) uint64ByteArrayToString(data []byte) string {
	// Format use is strconv.FormatUint(id, 10)
	// https://github.com/DataDog/dd-trace-go/blob/v1.38.1/ddtrace/tracer/textmap.go#L246
	return strconv.FormatUint(obj.endian.Uint64(data), 10)
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

// ____________________ String converter ____________________

func NewHeaderConvString() HeaderValueConverterPort { return &headerConvString{} }

type headerConvString struct{}

func (obj headerConvString) traceToDatadog(value trace.TraceID) string {
	// Convert OpenTelemetry 128-bits trace ID to Datadog 64-bits trace IDs
	valueStringHex := value.String()[16:]
	// No error can happen since string comes from a fixed byte array
	valueInt, _ := strconv.ParseUint(valueStringHex, 16, 64)
	return strconv.FormatUint(valueInt, 10)
}

func (obj headerConvString) traceFromDatadog(value string) (trace.TraceID, error) {
	// Datadog uses 64-bits data
	id64b, err := parseUint64(value)
	if err != nil {
		return trace.TraceID{}, err
	}
	return trace.TraceIDFromHex(fmt.Sprintf("%032x", id64b))
}

func (obj headerConvString) spanToDatadog(value trace.SpanID) string {
	// No error can happen since string comes from a fixed byte array
	valueInt, _ := strconv.ParseUint(value.String(), 16, 64)
	return strconv.FormatUint(valueInt, 10)
}

func (obj headerConvString) spanFromDatadog(value string) (trace.SpanID, error) {
	id64b, err := parseUint64(value)
	if err != nil {
		return trace.SpanID{}, err
	}
	return trace.SpanIDFromHex(fmt.Sprintf("%016x", id64b))
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
