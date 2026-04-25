package request_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
)

// body is the reference struct used by the decode tests. Every common
// tag + pointer/slice/map field that the handler layer uses is
// represented so the test exercises the full classifyValidationError
// path.
type body struct {
	Name    string            `json:"name"             validate:"required,min=1,max=100"`
	Email   string            `json:"email"            validate:"required,email"`
	Count   int               `json:"count"            validate:"required,gte=1,lte=100"`
	Adapter string            `json:"adapter"          validate:"required,oneof=openai anthropic azure"`
	URL     *string           `json:"url,omitempty"    validate:"omitempty,url"`
	Tags    []string          `json:"tags,omitempty"   validate:"omitempty,dive,min=1"`
	Headers map[string]string `json:"headers,omitempty"`
}

func postJSON(t *testing.T, raw string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func mustAppErr(t *testing.T, err error) *appErrors.AppError {
	t.Helper()
	appErr := appErrors.AsAppError(err)
	if appErr == nil {
		t.Fatalf("error is not *appErrors.AppError: %T %v", err, err)
	}
	return appErr
}

func TestDecodeJSON_Valid(t *testing.T) {
	raw := `{"name":"x","email":"a@b.co","count":5,"adapter":"openai"}`
	var b body
	if err := request.DecodeJSON(postJSON(t, raw), &b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Name != "x" || b.Email != "a@b.co" || b.Count != 5 || b.Adapter != "openai" {
		t.Errorf("decoded body mismatch: %+v", b)
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(""))
	var b body
	err := request.DecodeJSON(req, &b)
	appErr := mustAppErr(t, err)
	if appErr.Type != appErrors.TypeInvalidRequest {
		t.Errorf("Type = %q, want %q", appErr.Type, appErrors.TypeInvalidRequest)
	}
}

func TestDecodeJSON_Malformed(t *testing.T) {
	raw := `{"name": "x", "count":`
	var b body
	err := request.DecodeJSON(postJSON(t, raw), &b)
	appErr := mustAppErr(t, err)
	if appErr.Type != appErrors.TypeInvalidRequest {
		t.Errorf("Type = %q, want %q", appErr.Type, appErrors.TypeInvalidRequest)
	}
}

func TestDecodeJSON_UnknownField(t *testing.T) {
	raw := `{"name":"x","email":"a@b.co","count":1,"adapter":"openai","surprise":"hello"}`
	var b body
	err := request.DecodeJSON(postJSON(t, raw), &b)
	appErr := mustAppErr(t, err)
	if appErr.Type != appErrors.TypeValidation {
		t.Errorf("Type = %q, want %q", appErr.Type, appErrors.TypeValidation)
	}
	if appErr.Param != "surprise" {
		t.Errorf("Param = %q, want %q", appErr.Param, "surprise")
	}
}

func TestDecodeJSON_WrongType(t *testing.T) {
	raw := `{"name":"x","email":"a@b.co","count":"five","adapter":"openai"}`
	var b body
	err := request.DecodeJSON(postJSON(t, raw), &b)
	appErr := mustAppErr(t, err)
	if appErr.Type != appErrors.TypeValidation {
		t.Errorf("Type = %q, want %q", appErr.Type, appErrors.TypeValidation)
	}
	if appErr.Param != "count" {
		t.Errorf("Param = %q, want %q", appErr.Param, "count")
	}
}

func TestDecodeJSON_TrailingData(t *testing.T) {
	// Two JSON documents concatenated — must be rejected even if the
	// first is valid.
	raw := `{"name":"x","email":"a@b.co","count":1,"adapter":"openai"}{"name":"y","email":"b@c.co","count":2,"adapter":"azure"}`
	var b body
	err := request.DecodeJSON(postJSON(t, raw), &b)
	appErr := mustAppErr(t, err)
	if appErr.Type != appErrors.TypeInvalidRequest {
		t.Errorf("Type = %q, want %q", appErr.Type, appErrors.TypeInvalidRequest)
	}
}

func TestDecodeJSON_ValidationErrors(t *testing.T) {
	raw := `{"name":"","email":"not-an-email","count":0,"adapter":"???"}`
	var b body
	err := request.DecodeJSON(postJSON(t, raw), &b)
	appErr := mustAppErr(t, err)
	if appErr.Type != appErrors.TypeValidation {
		t.Fatalf("Type = %q, want %q", appErr.Type, appErrors.TypeValidation)
	}
	if len(appErr.Errors) < 3 {
		t.Fatalf("expected >= 3 per-field errors, got %d: %+v", len(appErr.Errors), appErr.Errors)
	}

	// Verify field paths use the JSON tag, not the Go field name.
	fields := map[string]bool{}
	for _, d := range appErr.Errors {
		fields[d.Location] = true
	}
	for _, want := range []string{"body.name", "body.email", "body.adapter"} {
		if !fields[want] {
			t.Errorf("missing %q in Errors[]; got keys: %v", want, fields)
		}
	}
}

func TestDecodeJSON_JSONTagFieldNames(t *testing.T) {
	// The api_key snake-case JSON tag must surface as "body.api_key"
	// in the validation error, not "APIKey".
	type snakeCase struct {
		APIKey string `json:"api_key" validate:"required,min=10"`
	}
	var b snakeCase
	err := request.DecodeJSON(postJSON(t, `{"api_key":"short"}`), &b)
	appErr := mustAppErr(t, err)
	if len(appErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %+v", len(appErr.Errors), appErr.Errors)
	}
	if appErr.Errors[0].Location != "body.api_key" {
		t.Errorf("Location = %q, want %q", appErr.Errors[0].Location, "body.api_key")
	}
}

func TestURLParamUUID(t *testing.T) {
	want := uuid.New()
	r := chi.NewRouter()
	var got uuid.UUID
	var gotErr error
	r.Get("/o/{orgId}", func(w http.ResponseWriter, r *http.Request) {
		got, gotErr = request.URLParamUUID(r, "orgId")
	})

	// Valid.
	req := httptest.NewRequest(http.MethodGet, "/o/"+want.String(), nil)
	r.ServeHTTP(httptest.NewRecorder(), req)
	if gotErr != nil {
		t.Fatalf("unexpected err: %v", gotErr)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	// Invalid.
	req = httptest.NewRequest(http.MethodGet, "/o/not-a-uuid", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)
	appErr := mustAppErr(t, gotErr)
	if appErr.Type != appErrors.TypeValidation {
		t.Errorf("Type = %q, want %q", appErr.Type, appErrors.TypeValidation)
	}
	if appErr.Param != "orgId" {
		t.Errorf("Param = %q, want %q", appErr.Param, "orgId")
	}
}

func TestQueryOptionalBool(t *testing.T) {
	cases := []struct {
		raw     string
		wantNil bool
		wantVal bool
		wantErr bool
	}{
		{raw: "", wantNil: true},
		{raw: "?has_error=true", wantNil: false, wantVal: true},
		{raw: "?has_error=false", wantNil: false, wantVal: false},
		{raw: "?has_error=1", wantNil: false, wantVal: true},
		{raw: "?has_error=xyz", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/x"+c.raw, nil)
			v, err := request.QueryOptionalBool(req, "has_error")
			if c.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if c.wantNil && v != nil {
				t.Errorf("wanted nil, got %v", *v)
			}
			if !c.wantNil && (v == nil || *v != c.wantVal) {
				t.Errorf("wanted %v, got %v", c.wantVal, v)
			}
		})
	}
}

func TestQueryPagination(t *testing.T) {
	cases := []struct {
		raw       string
		wantPage  int
		wantLimit int
		wantErr   bool
	}{
		{raw: "", wantPage: 1, wantLimit: 50},
		{raw: "?page=3", wantPage: 3, wantLimit: 50},
		{raw: "?page=3&limit=25", wantPage: 3, wantLimit: 25},
		{raw: "?page=0", wantErr: true},
		{raw: "?limit=0", wantErr: true},
		{raw: "?limit=5000", wantErr: true},
		{raw: "?page=abc", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/x"+c.raw, nil)
			page, limit, err := request.QueryPagination(req)
			if c.wantErr {
				if err == nil {
					t.Errorf("expected error, got page=%d limit=%d", page, limit)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if page != c.wantPage || limit != c.wantLimit {
				t.Errorf("got page=%d limit=%d, want page=%d limit=%d",
					page, limit, c.wantPage, c.wantLimit)
			}
		})
	}
}
