[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=SylvainDumas_opentelemetry-datadog-go&metric=coverage)](https://sonarcloud.io/summary/new_code?id=SylvainDumas_opentelemetry-datadog-go)

# opentelemetry datadog for go

[OpenTelemetry](https://opentelemetry.io) propagators are used to extract and inject context data from and into messages exchanged by applications. The propagator supported by this package is the Datadog Trace Context.

## Trace context propagation

| Span Context      | Size     |      | DD header key               | Size    | Text Format     |
|-------------------|----------|------|-----------------------------|---------|-----------------|
| TraceId           | 128 bits | <--> | x-datadog-trace-id          | 64 bits | number base 10  |
| SpanId            | 64 bits  | <--> | x-datadog-parent-id         | 64 bits | number base 10  |
| Sampling decision | 1 bit    | <--> | x-datadog-sampling-priority | bool    | "0" or "1"      |

You can find a getting started guide on [opentelemetry.io](https://opentelemetry.io/docs/instrumentation/go/getting-started).

### Getting Started

```shell
go get github.com/SylvainDumas/opentelemetry-datadog-go
```

If you installed more packages than you intended, you can use `go mod tidy` to remove any unused packages.

## Examples

```go
import (
    //...
	"github.com/SylvainDumas/opentelemetry-datadog-go/propagators/tracecontext"
	"go.opentelemetry.io/otel"
)

func initTracerProvider() {
    // ...
	otel.SetTextMapPropagator(tracecontext.NewDefault())
}

```

### Documentation

- [Datadog](https://www.datadoghq.com)
- [OpenTelemetry data sources](https://opentelemetry.io/docs/concepts/data-sources)
