package params_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/pkg/params"
)

// Compile-time check: Optional[T] must satisfy Huma's ParamWrapper +
// ParamReactor interfaces. If a future Huma release changes the
// interface signatures, this assertion breaks at build time.
var (
	_ huma.ParamWrapper = (*params.Optional[bool])(nil)
	_ huma.ParamReactor = (*params.Optional[bool])(nil)
)

// TestOptional_Zero covers the default state used by handlers that
// haven't yet seen a request: IsSet=false, Ptr()=nil.
func TestOptional_Zero(t *testing.T) {
	var o params.Optional[bool]
	assert.False(t, o.IsSet)
	assert.False(t, o.Value)
	assert.Nil(t, o.Ptr())
}

// TestOptional_PtrReturnsCopy guards against a subtle bug shape —
// Ptr() returns a pointer into a local copy of Value, so the caller
// cannot accidentally mutate the Optional via the returned pointer.
func TestOptional_PtrReturnsCopy(t *testing.T) {
	o := params.Optional[int]{Value: 42, IsSet: true}
	p := o.Ptr()
	require.NotNil(t, p)
	*p = 99
	assert.Equal(t, 42, o.Value, "mutating Ptr() return value must not change the Optional")
}

// TestOptional_EndToEnd_Absent_False_True exercises the full Huma
// request path. The three cases — absent, ?has_error=false,
// ?has_error=true — must produce distinguishable handler inputs
// (IsSet=false, IsSet=true+Value=false, IsSet=true+Value=true).
//
// If ParamReactor wiring breaks in a future Huma upgrade, this test
// fails loudly instead of silently collapsing absent=false.
func TestOptional_EndToEnd_Absent_False_True(t *testing.T) {
	type input struct {
		HasError params.Optional[bool] `query:"has_error"`
	}
	type output struct {
		Body struct {
			IsSet bool `json:"is_set"`
			Value bool `json:"value"`
		}
	}

	_, api := humatest.New(t, huma.DefaultConfig("test", "0.0.0"))
	huma.Register(api, huma.Operation{
		OperationID: "probe",
		Method:      http.MethodGet,
		Path:        "/probe",
	}, func(_ context.Context, in *input) (*output, error) {
		o := &output{}
		o.Body.IsSet = in.HasError.IsSet
		o.Body.Value = in.HasError.Value
		return o, nil
	})

	cases := []struct {
		name        string
		query       string
		wantIsSet   bool
		wantValue   bool
	}{
		{"absent", "", false, false},
		{"explicit-false", "?has_error=false", true, false},
		{"explicit-true", "?has_error=true", true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := api.Get("/probe" + c.query)
			require.Equal(t, http.StatusOK, resp.Code, "body=%s", resp.Body.String())

			var body struct {
				IsSet bool `json:"is_set"`
				Value bool `json:"value"`
			}
			require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
			assert.Equalf(t, c.wantIsSet, body.IsSet,
				"IsSet: want %v, got %v (query=%q)", c.wantIsSet, body.IsSet, c.query)
			assert.Equalf(t, c.wantValue, body.Value,
				"Value: want %v, got %v (query=%q)", c.wantValue, body.Value, c.query)
		})
	}
}

// TestOptional_OpenAPIExposesBoolType asserts the generated OpenAPI
// schema advertises `type: boolean`, not a string enum. Proves
// Optional is transparent to clients — they see the natural scalar
// type they'd have seen with a plain `bool` query param.
func TestOptional_OpenAPIExposesBoolType(t *testing.T) {
	type input struct {
		HasError params.Optional[bool] `query:"has_error"`
	}
	type output struct{ Body struct{} }

	_, api := humatest.New(t, huma.DefaultConfig("test", "0.0.0"))
	huma.Register(api, huma.Operation{
		OperationID: "probe",
		Method:      http.MethodGet,
		Path:        "/probe",
	}, func(_ context.Context, _ *input) (*output, error) { return &output{}, nil })

	b, err := json.Marshal(api.OpenAPI())
	require.NoError(t, err)

	var spec struct {
		Paths map[string]map[string]struct {
			Parameters []struct {
				Name   string `json:"name"`
				In     string `json:"in"`
				Schema struct {
					Type string `json:"type"`
					Enum []any  `json:"enum"`
				} `json:"schema"`
			} `json:"parameters"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(b, &spec))

	var found bool
	for _, op := range spec.Paths["/probe"] {
		for _, p := range op.Parameters {
			if p.Name == "has_error" && p.In == "query" {
				found = true
				assert.Equal(t, "boolean", p.Schema.Type)
				assert.Empty(t, p.Schema.Enum, "schema must not collapse to a string enum")
			}
		}
	}
	assert.True(t, found, "has_error query parameter must be declared in OpenAPI")
}
