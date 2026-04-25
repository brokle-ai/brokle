package observability

// Cross-plane shared types for the observability package.
//
// Per-plane DTOs live in dashboard_types.go and sdk_types.go; OTLP
// has no JSON DTOs at all (its request/response shapes are raw OTEL
// protobufs handled directly in otlp.go). Anything shared across ≥2
// planes (e.g. a common pagination envelope if SDK ever adopts one)
// lands here to avoid duplication.
