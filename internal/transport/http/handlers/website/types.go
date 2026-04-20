package website

// Huma operation types for the website package. One operation today
// (submit contact form); additional operations append their
// Input/Output + body/response pairs here.

// SubmitContactInput is the typed request shape Huma uses to
// validate the request body before invoking the handler. Validation
// tags map to OpenAPI 3.1 constraints + runtime checks; failures
// surface as 422 validation_error responses through the apiResponse
// envelope override.
type SubmitContactInput struct {
	Body submitContactBody
}

// submitContactBody is the JSON body schema. Field tags carry the
// same constraints the gin handler enforced via `binding:` tags;
// Huma reads them from the `minLength`/`maxLength`/`format` form
// for both validation and OpenAPI doc emission.
type submitContactBody struct {
	Name        string `json:"name" minLength:"2" maxLength:"255" doc:"Contact name"`
	Email       string `json:"email" format:"email" maxLength:"255" doc:"Contact email"`
	Company     string `json:"company,omitempty" maxLength:"255" doc:"Optional company name"`
	Subject     string `json:"subject" minLength:"5" maxLength:"255" doc:"Subject line"`
	Message     string `json:"message" minLength:"10" maxLength:"5000" doc:"Message body"`
	InquiryType string `json:"inquiry_type,omitempty" maxLength:"50" doc:"Optional inquiry classification (sales, support, partnership, …)"`
}

// SubmitContactOutput is the typed response. The Body field is what
// Huma serialises into the response envelope.
type SubmitContactOutput struct {
	Body submitContactResponse
}

// submitContactResponse is the response shape returned to the
// browser. Carries a human-readable confirmation message safe to
// render verbatim.
type submitContactResponse struct {
	Message string `json:"message" doc:"Confirmation message safe to render verbatim"`
}
