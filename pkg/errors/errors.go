// Package errors implements Brokle's unified domain error type. The
// shape mirrors Stripe / OpenAI / Anthropic on the wire — a closed
// coarse Reason (used by SDK consumers for retry/alert/error-class
// branching) plus an open fine Code (used for observability and
// SDK-side specific subclasses):
//
//	{
//	  "type":    "validation_error",   // closed enum, derived from Reason
//	  "code":    "project_not_found",   // open snake_case enum, optional
//	  "message": "...",
//	  "details": "...",                 // optional supplementary text
//	  "param":   "projectId",           // optional input field hint
//	  "errors":  [{...}]                // optional per-field validation
//	}
//
// HTTP status is a pure function of Reason via Reason.HTTPStatus —
// never derive status from a stored field, and never let callers
// override it. Framework-level errors arriving by status code (chi
// middleware, request decoders) are categorised via FromHTTPStatus.
//
// Pattern verified against k8s apimachinery (`apierrors.StatusError`),
// Grafana `errutil`, HashiCorp Boundary `internal/errors`, CockroachDB
// `cockroachdb/errors`. Single error type at a top-level shared path
// is the industry default; the earlier two-package experiment
// (`derrors` + `appErrors`) was folded into this one on 2026-04-28.
//
// USAGE:
//
//	// Repository (the single translation point):
//	row, err := q.GetProjectByID(ctx, id)
//	if db.IsNoRows(err) {
//	    return nil, appErrors.NotFound("project", appErrors.WithOp("repo.project.get_by_id"))
//	}
//	if err != nil {
//	    return nil, appErrors.Internal("get project", err, appErrors.WithOp("repo.project.get_by_id"))
//	}
//
//	// Service (passthrough is correct — error is self-describing):
//	func (s *ProjectService) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
//	    return s.projectRepo.GetByID(ctx, id)
//	}
//
//	// Caller branching on a kind:
//	if appErrors.IsNotFound(err) { … }
//	if appErrors.IsAlreadyExists(err) { … }
package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
)

// Reason is the semantic kind of an error. Each transport adapter maps
// Reason → its native status (HTTP code, gRPC Code, …). Closed and
// stable — adding a value is a deliberate API change. Keep in lockstep
// with reasonToStatus / reasonToHTTPType below.
type Reason int

const (
	// ReasonUnspecified is the iota=0 placeholder. NEVER set this on a
	// real *Error — it's reserved as the "absent reason" sentinel per
	// proto3 / k8s `StatusReasonUnknown` / gRPC `codes.Unknown`
	// convention. HTTPStatus() falls through to 500 and HTTPType()
	// falls through to "api_error" if encountered, but production
	// code should always set a real reason via the typed constructors.
	// Reserving iota=0 prevents the "&Error{} matches Internal"
	// sentinel collision that bit us before this rule was codified.
	ReasonUnspecified Reason = iota

	// ReasonInternal — unexpected internal failure. Wraps a Cause for
	// log/trace export. HTTP 500 / gRPC Internal.
	ReasonInternal

	// ReasonNotFound — the addressed resource does not exist. HTTP 404 /
	// gRPC NotFound.
	ReasonNotFound

	// ReasonAlreadyExists — the resource being created already exists
	// (UNIQUE constraint). HTTP 409 / gRPC AlreadyExists.
	ReasonAlreadyExists

	// ReasonInvalidInput — request body parsed but failed declarative
	// validation (struct-tag constraints, business rules). HTTP 422 /
	// gRPC InvalidArgument.
	ReasonInvalidInput

	// ReasonBadRequest — the request was syntactically rejected before
	// validation (malformed body, oversize, unsupported media type,
	// method not allowed). HTTP 400 / gRPC InvalidArgument.
	ReasonBadRequest

	// ReasonUnauthenticated — credentials missing/invalid. HTTP 401 /
	// gRPC Unauthenticated.
	ReasonUnauthenticated

	// ReasonPermissionDenied — caller is authenticated but lacks the
	// required permission. HTTP 403 / gRPC PermissionDenied.
	ReasonPermissionDenied

	// ReasonConflict — state-change conflict (optimistic concurrency,
	// invalid lifecycle transition). HTTP 409 / gRPC FailedPrecondition.
	ReasonConflict

	// ReasonRateLimit — request rate or quota exceeded. HTTP 429 /
	// gRPC ResourceExhausted.
	ReasonRateLimit

	// ReasonPaymentRequired — the action requires a billable plan or
	// has exhausted included credits. HTTP 402.
	ReasonPaymentRequired

	// ReasonUpstream — an upstream provider (LLM, external API)
	// returned an error. HTTP 502 / gRPC Unavailable.
	ReasonUpstream

	// ReasonUnavailable — service is temporarily unable to handle the
	// request (maintenance, overload). HTTP 503 / gRPC Unavailable.
	ReasonUnavailable

	// ReasonNotImplemented — operation exists in the schema but is not
	// implemented in this build. HTTP 501 / gRPC Unimplemented.
	ReasonNotImplemented
)

// String returns the snake_case name of the reason; useful for
// structured-log fields and error messages.
func (r Reason) String() string {
	switch r {
	case ReasonUnspecified:
		return "unspecified"
	case ReasonInternal:
		return "internal"
	case ReasonNotFound:
		return "not_found"
	case ReasonAlreadyExists:
		return "already_exists"
	case ReasonInvalidInput:
		return "invalid_input"
	case ReasonBadRequest:
		return "bad_request"
	case ReasonUnauthenticated:
		return "unauthenticated"
	case ReasonPermissionDenied:
		return "permission_denied"
	case ReasonConflict:
		return "conflict"
	case ReasonRateLimit:
		return "rate_limit"
	case ReasonPaymentRequired:
		return "payment_required"
	case ReasonUpstream:
		return "upstream"
	case ReasonUnavailable:
		return "unavailable"
	case ReasonNotImplemented:
		return "not_implemented"
	default:
		return fmt.Sprintf("unknown(%d)", int(r))
	}
}

// HTTPStatus returns the canonical HTTP status code for a Reason.
// Pure function; the wire renderer (pkg/response.WriteError) reads
// this. Unknown reasons fall back to 500.
func (r Reason) HTTPStatus() int {
	switch r {
	case ReasonNotFound:
		return http.StatusNotFound
	case ReasonAlreadyExists, ReasonConflict:
		return http.StatusConflict
	case ReasonInvalidInput:
		return http.StatusUnprocessableEntity
	case ReasonBadRequest:
		return http.StatusBadRequest
	case ReasonUnauthenticated:
		return http.StatusUnauthorized
	case ReasonPermissionDenied:
		return http.StatusForbidden
	case ReasonRateLimit:
		return http.StatusTooManyRequests
	case ReasonPaymentRequired:
		return http.StatusPaymentRequired
	case ReasonUpstream:
		return http.StatusBadGateway
	case ReasonUnavailable:
		return http.StatusServiceUnavailable
	case ReasonNotImplemented:
		return http.StatusNotImplemented
	default: // ReasonInternal + unknown
		return http.StatusInternalServerError
	}
}

// HTTPType returns the wire-shape `error.type` string for a Reason.
// SDK consumers branch on this closed enum; the values are stable.
func (r Reason) HTTPType() string {
	switch r {
	case ReasonNotFound:
		return "not_found_error"
	case ReasonAlreadyExists, ReasonConflict:
		return "conflict_error"
	case ReasonInvalidInput:
		return "validation_error"
	case ReasonBadRequest:
		return "invalid_request_error"
	case ReasonUnauthenticated:
		return "authentication_error"
	case ReasonPermissionDenied:
		return "permission_error"
	case ReasonRateLimit:
		return "rate_limit_error"
	case ReasonPaymentRequired:
		return "payment_required"
	case ReasonUpstream:
		return "upstream_provider_error"
	case ReasonUnavailable:
		return "service_unavailable"
	case ReasonNotImplemented:
		return "not_implemented"
	default: // ReasonInternal + unknown
		return "api_error"
	}
}

// ErrorDetail carries a per-field validation diagnostic. One entry per
// rejected field, populated by pkg/request.DecodeJSON from
// go-playground/validator errors.
//
// Location is a dotted path from the root of the request document —
// "body.items[3].tags" for a nested JSON field, "query.page" for a
// query parameter. Message is the humanised validation failure
// ("required", "must be ≥ 1"). Value is the offending input echoed
// back verbatim.
type ErrorDetail struct {
	Location string `json:"location,omitempty"`
	Message  string `json:"message"`
	Value    any    `json:"value,omitempty"`
}

// Error is the canonical error type. Construct via the typed helpers
// (NotFound, AlreadyExists, …) — direct struct-literal construction is
// reserved for tests.
//
// Fields beyond Reason are all optional. Code is opt-in (Stripe shape:
// only emit when programmatically actionable). Param hints at the
// offending input field. Errors carries per-field validator details.
// Op + Cause are log-only (never serialised to clients).
type Error struct {
	Reason   Reason
	Resource string        // entity name involved ("project", "user", "api_key")
	Message  string        // optional override; default derived from Reason+Resource
	Details  string        // optional supplementary text (lowercase, ≤ one sentence)
	Code     string        // optional open-enum subcode (see codes.go for the catalogue)
	Param    string        // optional input field hint (matches OpenAI's error.param)
	Errors   []ErrorDetail // optional per-field validation diagnostics
	Op       string        // failing op, e.g. "repo.project.get_by_id" — log-only
	Cause    error         // wrapped underlying error — Unwrap chain, log-only
}

// Error implements the error interface. Format:
// "<Op>: <message>[: <Cause>]". Never written to clients.
func (e *Error) Error() string {
	parts := make([]string, 0, 3)
	if e.Op != "" {
		parts = append(parts, e.Op+":")
	}
	if msg := e.publicMessage(); msg != "" {
		parts = append(parts, msg)
	} else {
		parts = append(parts, e.Reason.String())
	}
	out := strings.Join(parts, " ")
	if e.Cause != nil {
		out = fmt.Sprintf("%s: %v", out, e.Cause)
	}
	return out
}

// Unwrap exposes Cause for errors.Is / errors.As chains.
func (e *Error) Unwrap() error { return e.Cause }

// HTTPStatus returns the canonical HTTP status for the error's Reason.
func (e *Error) HTTPStatus() int { return e.Reason.HTTPStatus() }

// Is matches when target is also a *Error with the same Reason. When
// target.Resource is non-empty it must also match. When target.Code is
// non-empty it must also match.
//
// There is no wildcard reason value — ReasonUnspecified is the iota=0
// placeholder reserved as the "absent" sentinel and is never set on
// real errors. To check "is this any *Error" use appErrors.As(err) !=
// nil, the idiomatic Go pattern (k8s apierrors / pgx / Boundary all
// expose package-level predicates + errors.As, never an Is wildcard;
// see the 2026-04-29 Lessons Learned for the precedent survey and the
// sentinel-collision regression that motivated this design).
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	if t.Reason != e.Reason {
		return false
	}
	if t.Resource != "" && t.Resource != e.Resource {
		return false
	}
	if t.Code != "" && t.Code != e.Code {
		return false
	}
	return true
}

// PublicMessage returns the wire-safe human-readable message. When
// Message is set, returns it verbatim; otherwise derived from
// Reason+Resource ("project not found", "user already exists"). Never
// includes Cause — that's log-only.
func (e *Error) PublicMessage() string { return e.publicMessage() }

func (e *Error) publicMessage() string {
	if e.Message != "" {
		return e.Message
	}
	switch e.Reason {
	case ReasonUnauthenticated:
		return "authentication required"
	case ReasonRateLimit:
		return "rate limit exceeded"
	case ReasonPaymentRequired:
		return "payment required"
	case ReasonBadRequest:
		return "bad request"
	}
	if e.Resource == "" {
		return e.Reason.String()
	}
	switch e.Reason {
	case ReasonNotFound:
		return e.Resource + " not found"
	case ReasonAlreadyExists:
		return e.Resource + " already exists"
	case ReasonConflict:
		return e.Resource + " conflict"
	case ReasonInvalidInput:
		return "invalid " + e.Resource
	case ReasonPermissionDenied:
		return "permission denied for " + e.Resource
	case ReasonUpstream:
		return "upstream " + e.Resource + " error"
	case ReasonUnavailable:
		return e.Resource + " temporarily unavailable"
	case ReasonNotImplemented:
		return e.Resource + " not implemented"
	default:
		return e.Reason.String()
	}
}

// MarshalJSON serialises Error as the canonical Brokle error envelope:
// `{"error":{type, code?, message, details?, param?, errors?}}`.
//
// The shape matches Stripe / OpenAI / Anthropic verbatim. There is
// deliberately NO top-level `success` boolean — HTTP status is the
// canonical success/failure signal (per RFC 9110 §15).
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Error inner `json:"error"`
	}{
		Error: inner{
			Type:    e.Reason.HTTPType(),
			Code:    e.Code,
			Message: e.publicMessage(),
			Details: e.Details,
			Param:   e.Param,
			Errors:  e.Errors,
		},
	})
}

// inner is the on-the-wire shape of the `error` field in the envelope.
// Kept package-private; consumers read via the APIError component in
// pkg/response.
type inner struct {
	Type    string        `json:"type"`
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message"`
	Details string        `json:"details,omitempty"`
	Param   string        `json:"param,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

// ----- Functional options ----------------------------------------------------

// Option mutates an *Error during construction. Functional-options
// pattern keeps constructor signatures stable as new optional fields
// are added without growing the constructor matrix.
type Option func(*Error)

// WithMessage overrides the default message derived from Reason+Resource.
// Use when the caller has a more specific phrasing than "project not found".
func WithMessage(message string) Option {
	return func(e *Error) { e.Message = message }
}

// WithDetails sets a free-form details string supplementing Message.
// Use lowercase per the established convention:
// `WithDetails("projectId must be a valid UUID")`.
func WithDetails(details string) Option {
	return func(e *Error) { e.Details = details }
}

// WithCode sets the open-enum domain code (see codes.go for the
// catalogue). Stripe behaviour: omit unless programmatically
// actionable.
func WithCode(code string) Option {
	return func(e *Error) { e.Code = code }
}

// WithParam annotates the error with the input field that triggered
// it. SDK clients use this to render per-field error messages and to
// map failures back to form inputs. Matches OpenAI's `error.param`.
func WithParam(param string) Option {
	return func(e *Error) { e.Param = param }
}

// WithFieldErrors attaches per-field validation diagnostics. Populated
// by the request-binding layer (pkg/request.DecodeJSON) when
// go-playground/validator rejects fields, and by any other callsite
// that needs to surface multiple independent input problems on one
// response. Empty slices are omitted from the wire envelope.
func WithFieldErrors(details []ErrorDetail) Option {
	return func(e *Error) { e.Errors = details }
}

// WithOp annotates the error with the failing operation (e.g.
// "repo.project.get_by_id"). Surfaced in Error() and structured logs;
// never on the wire.
func WithOp(op string) Option {
	return func(e *Error) { e.Op = op }
}

// WithCause attaches an underlying error for errors.Is/As traversal.
// Surfaced in Error() and the Unwrap chain; never on the wire.
func WithCause(cause error) Option {
	return func(e *Error) { e.Cause = cause }
}

// ----- Typed constructors (resource-required) -----------------------------

// NotFound: the addressed resource does not exist.
func NotFound(resource string, opts ...Option) *Error {
	return newErr(ReasonNotFound, resource, opts)
}

// AlreadyExists: the resource being created already exists.
func AlreadyExists(resource string, opts ...Option) *Error {
	return newErr(ReasonAlreadyExists, resource, opts)
}

// Conflict: state-change conflict (optimistic concurrency, invalid
// lifecycle transition). Message is required — the default
// "<resource> conflict" is too vague to be useful, and 100% of
// historical callsites set one.
func Conflict(resource, message string, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message)}, opts...)
	return newErr(ReasonConflict, resource, opts)
}

// InvalidParam: a single named input field couldn't parse or violates
// its constraint. HTTP 422 with `error.param` populated so SDK
// consumers can surface per-field errors. Pattern verified against
// k8s `apimachinery/util/validation/field` (per-cause constructors).
//
// Use InvalidFields(details) when multiple fields failed validation
// in one request (validator output). Use Conflict / BadRequest for
// entity-level or request-shape rejections.
func InvalidParam(field, message string, opts ...Option) *Error {
	opts = append([]Option{WithParam(field), WithMessage(message)}, opts...)
	return newErr(ReasonInvalidInput, "", opts)
}

// InvalidFields: multiple input fields failed validation in one
// request — typically the output of go-playground/validator on a
// struct decode. HTTP 422 with `error.errors[]` populated; the first
// field's location is also set as Param so single-param SDK consumers
// (Stripe-style) can highlight one input. Each ErrorDetail names its
// field via Location and carries a humanised message + the offending
// value.
//
// Use InvalidParam(field, message) when there's only one bad field —
// the multi-field constructor is for the validator-boundary case
// where N fields failed in one request.
//
// Pattern verified against k8s `apierrors.NewInvalid(kind, name,
// errs field.ErrorList)` — the canonical multi-field validator
// adapter at the transport boundary.
func InvalidFields(details []ErrorDetail, opts ...Option) *Error {
	primary := []Option{WithFieldErrors(details)}
	if len(details) > 0 {
		primary = append(primary, WithParam(details[0].Location))
	}
	return newErr(ReasonInvalidInput, "", append(primary, opts...))
}

// PermissionDenied: caller is authenticated but lacks the required
// permission. Message is required — every callsite specifies which
// permission was missing.
func PermissionDenied(resource, message string, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message)}, opts...)
	return newErr(ReasonPermissionDenied, resource, opts)
}

// Upstream: an upstream provider (LLM, external API) returned an error.
func Upstream(resource string, cause error, opts ...Option) *Error {
	opts = append([]Option{WithCause(cause)}, opts...)
	return newErr(ReasonUpstream, resource, opts)
}

// Unavailable: service is temporarily unable to handle the request.
func Unavailable(resource string, opts ...Option) *Error {
	return newErr(ReasonUnavailable, resource, opts)
}

// NotImplemented: operation exists in the schema but is not
// implemented. Message is required — say what's not implemented.
func NotImplemented(resource, message string, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message)}, opts...)
	return newErr(ReasonNotImplemented, resource, opts)
}

// ----- Typed constructors (resource-optional) -----------------------------

// BadRequest: request was syntactically rejected before validation
// (malformed body, oversize, unsupported media type, method not
// allowed). HTTP 400. Message is required — say what's wrong with
// the request.
func BadRequest(message string, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message)}, opts...)
	return newErr(ReasonBadRequest, "", opts)
}

// Unauthenticated: credentials missing/invalid. Message is required
// — say which credentials are missing or why they were rejected.
func Unauthenticated(message string, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message)}, opts...)
	return newErr(ReasonUnauthenticated, "", opts)
}

// RateLimit: request rate or quota exceeded. Message is required.
func RateLimit(message string, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message)}, opts...)
	return newErr(ReasonRateLimit, "", opts)
}

// PaymentRequired: the action requires a billable plan or has
// exhausted included credits. HTTP 402.
func PaymentRequired(opts ...Option) *Error {
	return newErr(ReasonPaymentRequired, "", opts)
}

// Internal: unexpected internal failure. Cause is REQUIRED — callers
// should always wrap the underlying error so logs and traces preserve
// the chain.
func Internal(message string, cause error, opts ...Option) *Error {
	opts = append([]Option{WithMessage(message), WithCause(cause)}, opts...)
	return newErr(ReasonInternal, "", opts)
}

func newErr(reason Reason, resource string, opts []Option) *Error {
	e := &Error{Reason: reason, Resource: resource}
	for _, o := range opts {
		o(e)
	}
	return e
}

// ----- Inspection helpers, shaped after stdlib `errors` --------------------

// As returns the wrapped *Error if err is or wraps one, nil otherwise.
// Mirrors errors.As ergonomics for the common single-type case:
//
//	if e := appErrors.As(err); e != nil {
//	    switch e.Reason { … }
//	}
func As(err error) *Error {
	var e *Error
	if stderrors.As(err, &e) {
		return e
	}
	return nil
}

// IsReason reports whether err is or wraps a *Error with the given
// Reason. Convenience predicate for call sites that don't care about
// the resource.
func IsReason(err error, reason Reason) bool {
	if e := As(err); e != nil {
		return e.Reason == reason
	}
	return false
}

// IsNotFound is the canonical convenience predicate (matches k8s
// apierrors.IsNotFound shape). Callers that care about the resource
// can use As(err) and inspect e.Resource.
func IsNotFound(err error) bool { return IsReason(err, ReasonNotFound) }

// IsAlreadyExists is the AlreadyExists convenience predicate.
func IsAlreadyExists(err error) bool { return IsReason(err, ReasonAlreadyExists) }

// IsConflict is the Conflict convenience predicate.
func IsConflict(err error) bool { return IsReason(err, ReasonConflict) }

// HTTPStatus returns the HTTP status implied by err. Falls back to 500
// for non-*Error errors.
func HTTPStatus(err error) int {
	if e := As(err); e != nil {
		return e.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// ----- Framework-boundary categorisation -----------------------------------

// FromHTTPStatus synthesises an *Error from an HTTP status code, used
// by the framework boundary for errors that arrive without an *Error
// context (request body decode 422, content-type negotiation 406/415,
// method routing 405, body size 413, …).
//
// Class-fallback safe: any 4xx not in the explicit map becomes
// ReasonBadRequest, any 5xx becomes ReasonInternal. This eliminates
// the regression class where unmapped framework statuses leaked into
// the server-fault category.
//
// cause is variadic so callers can pass either zero or one originating
// error; only the first non-nil cause is recorded.
func FromHTTPStatus(status int, message string, cause ...error) *Error {
	r := reasonFromStatus(status)
	opts := []Option{WithMessage(message)}
	for _, c := range cause {
		if c != nil {
			opts = append(opts, WithCause(c))
			break
		}
	}
	return newErr(r, "", opts)
}

// reasonFromStatus is the inverse of Reason.HTTPStatus, used only at
// the framework boundary where a bare HTTP status arrives without
// *Error context.
func reasonFromStatus(status int) Reason {
	switch status {
	case http.StatusBadRequest,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusRequestTimeout,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusMisdirectedRequest,
		http.StatusLocked,
		http.StatusFailedDependency,
		http.StatusTooEarly,
		http.StatusUpgradeRequired,
		http.StatusPreconditionRequired,
		http.StatusRequestHeaderFieldsTooLarge,
		http.StatusUnavailableForLegalReasons:
		return ReasonBadRequest
	case http.StatusUnauthorized:
		return ReasonUnauthenticated
	case http.StatusPaymentRequired:
		return ReasonPaymentRequired
	case http.StatusForbidden:
		return ReasonPermissionDenied
	case http.StatusNotFound:
		return ReasonNotFound
	case http.StatusConflict:
		return ReasonConflict
	case http.StatusUnprocessableEntity:
		return ReasonInvalidInput
	case http.StatusTooManyRequests:
		return ReasonRateLimit
	case http.StatusNotImplemented:
		return ReasonNotImplemented
	case http.StatusBadGateway:
		return ReasonUpstream
	case http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected:
		return ReasonUnavailable
	default:
		if status >= 400 && status < 500 {
			return ReasonBadRequest
		}
		return ReasonInternal
	}
}
