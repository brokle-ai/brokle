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
	SchemaUrl  string     `json:"schemaUrl,omitempty"`
}

// ScopeSpan represents a collection of spans from a single instrumentation scope
type ScopeSpan struct {
	Scope *Scope     `json:"scope,omitempty"`
	Spans []OTLPSpan `json:"spans"`
}

// Scope represents an instrumentation scope
type Scope struct {
	Name       string     `json:"name"`
	Version    string     `json:"version,omitempty"`
	Attributes []KeyValue `json:"attributes,omitempty"`
	SchemaUrl  string     `json:"schemaUrl,omitempty"`
}

// OTLPSpan represents an OTLP span (wire format)
type OTLPSpan struct {
	TraceID           interface{} `json:"traceId"`
	SpanID            interface{} `json:"spanId"`
	ParentSpanID      interface{} `json:"parentSpanId,omitempty"`
	StartTimeUnixNano interface{} `json:"startTimeUnixNano"`
	EndTimeUnixNano   interface{} `json:"endTimeUnixNano,omitempty"`
	Status            *Status     `json:"status,omitempty"`
	Name              string      `json:"name"`
	Attributes        []KeyValue  `json:"attributes,omitempty"`
	Events            []Event     `json:"events,omitempty"`
	Links             []Link      `json:"links,omitempty"`
	Kind              int         `json:"kind,omitempty"`
}

// KeyValue represents an OTLP attribute key-value pair
type KeyValue struct {
	Value interface{} `json:"value"`
	Key   string      `json:"key"`
}

// Event represents an OTLP span event (timestamped annotation within a span)
type Event struct {
	TimeUnixNano           interface{} `json:"timeUnixNano"`
	Name                   string      `json:"name"`
	Attributes             []KeyValue  `json:"attributes,omitempty"`
	DroppedAttributesCount uint32      `json:"droppedAttributesCount,omitempty"` // Number of dropped attributes
}

// Link represents an OTLP span link (reference to span in another trace)
type Link struct {
	TraceID                interface{} `json:"traceId"`                          // Linked trace ID (Buffer or hex string)
	SpanID                 interface{} `json:"spanId"`                           // Linked span ID (Buffer or hex string)
	TraceState             interface{} `json:"traceState,omitempty"`             // W3C TraceState for linked span
	Attributes             []KeyValue  `json:"attributes,omitempty"`             // Link metadata
	DroppedAttributesCount uint32      `json:"droppedAttributesCount,omitempty"` // Number of dropped attributes
}

// Status represents OTLP status
type Status struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}
