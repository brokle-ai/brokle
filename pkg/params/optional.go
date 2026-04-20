// Package params defines Huma v2 parameter-helper types shared across
// handler packages. The canonical Optional[T] here mirrors Huma's own
// godoc example (huma.go:1278-1303 in v2.37.3) and implements
// huma.ParamWrapper + huma.ParamReactor so it can represent a
// tri-state optional query / header / path / cookie parameter
// without using a pointer.
//
// Background: Huma panics at huma.go:189 if a request parameter field
// is a pointer type, because the framework would have no way to
// distinguish "absent" from "zero value" at the struct-binding
// layer. Huma solves this via two interfaces:
//
//   - ParamWrapper.Receiver() — the wrapping type exposes the
//     inner scalar by reflect.Value, so Huma binds the parsed
//     value directly without seeing a pointer.
//   - ParamReactor.OnParamSet(isSet, parsed) — the wrapping type
//     observes whether the request carried the parameter at all.
//
// Huma's runtime at huma.go:881-901 only fires OnParamSet with
// isSet=true when the parameter was actually present in the
// request, so we can distinguish absent (IsSet=false) from zero
// value (IsSet=true, Value=false).
//
// The OpenAPI spec emitted for a field of type Optional[bool]
// renders `type: boolean` — clients send `?has_error=true` /
// `?has_error=false` / omitted exactly as they would for a non-
// wrapped `bool`. The wrapping is invisible on the wire.
package params

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// Optional wraps a scalar query / header / path / cookie parameter
// so the handler can distinguish "absent" from the zero value
// without a pointer. Implements huma.ParamWrapper +
// huma.ParamReactor.
//
// Typical usage:
//
//	type ListTracesInput struct {
//	    HasError params.Optional[bool] `query:"has_error" doc:"Filter by error presence"`
//	}
//	func (h *handler) list(ctx context.Context, in *ListTracesInput) (...) {
//	    filter.HasError = in.HasError.Ptr() // &true / &false / nil
//	}
//
// The receiver must be a pointer for Huma's interface-satisfaction
// check; the zero value (IsSet=false) is the correct initial state.
type Optional[T any] struct {
	Value T
	IsSet bool
}

// Schema satisfies huma.SchemaProvider so the generated OpenAPI
// parameter schema advertises `type: T` (e.g. boolean) instead of
// the wrapping struct's object shape. Without this, the default
// reflection walker (schema.go:565-612) sees Value + IsSet as
// struct fields and emits `type: object`, which then fails
// validation on inbound `true`/`false` strings. Huma's own
// huma_test.go:97 shows this as the required pattern.
func (o Optional[T]) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(o.Value))
}

// Receiver satisfies huma.ParamWrapper. Huma parses the request
// value into Value via the returned reflect.Value handle, bypassing
// the pointer-param panic path entirely.
//
// The returned value points at the Value field directly — Huma
// writes into it, and then OnParamSet fires with isSet=true.
func (o *Optional[T]) Receiver() reflect.Value {
	return reflect.ValueOf(o).Elem().Field(0)
}

// OnParamSet satisfies huma.ParamReactor. Huma invokes this AFTER
// parameter parsing with isSet=true only when the parameter was
// present in the request; absent parameters leave IsSet=false.
func (o *Optional[T]) OnParamSet(isSet bool, _ any) {
	o.IsSet = isSet
}

// Ptr returns &Value when IsSet, nil otherwise. Convenient for
// forwarding to domain filter types that use pointer fields to
// express tri-state (e.g. `filter.HasError = in.HasError.Ptr()`).
// The returned pointer addresses a local copy of Value; mutating it
// does NOT affect the Optional.
func (o Optional[T]) Ptr() *T {
	if !o.IsSet {
		return nil
	}
	v := o.Value
	return &v
}
