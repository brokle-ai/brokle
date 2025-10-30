package observability

// OTLP data structures for OpenTelemetry Protocol trace ingestion
// These types support both JSON and Protobuf formats

// OTLPRequest represents an OTLP trace export request
type OTLPRequest struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

// ResourceSpan represents a collection of spans from a single resource
type ResourceSpan struct {
	Resource   *Resource   `json:"resource,omitempty"`
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

// Resource represents OTEL resource attributes
type Resource struct {
	Attributes []KeyValue `json:"attributes"`
}

// ScopeSpan represents a collection of spans from a single instrumentation scope
type ScopeSpan struct {
	Scope *Scope `json:"scope,omitempty"`
	Spans []Span `json:"spans"`
}

// Scope represents an instrumentation scope
type Scope struct {
	Name       string     `json:"name"`
	Version    string     `json:"version,omitempty"`
	Attributes []KeyValue `json:"attributes,omitempty"`
}

// Span represents an OTLP span
type Span struct {
	TraceID           interface{} `json:"traceId"`                    // Can be Buffer or hex string
	SpanID            interface{} `json:"spanId"`                     // Can be Buffer or hex string
	ParentSpanID      interface{} `json:"parentSpanId,omitempty"`     // Can be Buffer or hex string
	Name              string      `json:"name"`
	Kind              int         `json:"kind,omitempty"`             // 0=UNSPECIFIED, 1=INTERNAL, 2=SERVER, 3=CLIENT, 4=PRODUCER, 5=CONSUMER
	StartTimeUnixNano interface{} `json:"startTimeUnixNano"`          // Can be int64 or {low, high}
	EndTimeUnixNano   interface{} `json:"endTimeUnixNano,omitempty"`
	Attributes        []KeyValue  `json:"attributes,omitempty"`
	Events            []Event     `json:"events,omitempty"`
	Status            *Status     `json:"status,omitempty"`
}

// KeyValue represents an OTLP attribute key-value pair
type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"` // Can be stringValue, intValue, boolValue, arrayValue, etc.
}

// Event represents an OTLP span event
type Event struct {
	TimeUnixNano interface{} `json:"timeUnixNano"`
	Name         string      `json:"name"`
	Attributes   []KeyValue  `json:"attributes,omitempty"`
}

// Status represents OTLP status
type Status struct {
	Code    int    `json:"code,omitempty"`    // 0=UNSET, 1=OK, 2=ERROR
	Message string `json:"message,omitempty"`
}
