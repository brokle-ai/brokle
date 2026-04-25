package website

// submitContactBody is the JSON body schema for POST
// /api/v1/website/contact. Validation tags drive
// go-playground/validator/v10 via pkg/request.DecodeJSON.
type submitContactBody struct {
	Name        string `json:"name"                   validate:"required,min=2,max=255"`
	Email       string `json:"email"                  validate:"required,email,max=255"`
	Company     string `json:"company,omitempty"      validate:"omitempty,max=255"`
	Subject     string `json:"subject"                validate:"required,min=5,max=255"`
	Message     string `json:"message"                validate:"required,min=10,max=5000"`
	InquiryType string `json:"inquiry_type,omitempty" validate:"omitempty,max=50"`
}

// submitContactResponse is the 201 body returned to the browser.
type submitContactResponse struct {
	Message string `json:"message"`
}
