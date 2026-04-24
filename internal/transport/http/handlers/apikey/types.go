package apikey

// createAPIKeyBody — POST /api/v1/projects/{projectId}/api-keys
type createAPIKeyBody struct {
	Name         string `json:"name"          validate:"required,min=2,max=100"`
	ExpiryOption string `json:"expiry_option" validate:"required,oneof=30days 90days never"`
}
