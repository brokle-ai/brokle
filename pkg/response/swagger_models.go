package response

// This file contains Swagger-specific models for common API structures

// SuccessResponse represents a successful API response
// @Description Standard successful response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true" description:"Always true for successful responses"`
	Data    interface{} `json:"data" description:"Response data payload"`
	Meta    *Meta       `json:"meta,omitempty" description:"Response metadata"`
}

// ErrorResponse represents an error API response
// @Description Standard error response
type ErrorResponse struct {
	Success bool      `json:"success" example:"false" description:"Always false for error responses"`
	Error   *APIError `json:"error" description:"Error details"`
	Meta    *Meta     `json:"meta,omitempty" description:"Response metadata"`
}

// MessageResponse represents a simple message response
// @Description Simple message response for actions
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully" description:"Response message"`
}

// IDResponse represents a response with a single ID
// @Description Response containing an ID reference
type IDResponse struct {
	ID string `json:"id" example:"ulid_01h2x3y4z5" description:"Resource identifier"`
}

// ListResponse represents a paginated list response
// @Description Paginated list response wrapper
type ListResponse struct {
	Success bool        `json:"success" example:"true" description:"Request success status"`
	Data    interface{} `json:"data" description:"Array of items"`
	Meta    struct {
		RequestID  string      `json:"request_id,omitempty" example:"req_01h2x3y4z5"`
		Timestamp  string      `json:"timestamp,omitempty" example:"2023-12-01T10:30:00Z"`
		Version    string      `json:"version,omitempty" example:"v1"`
		Pagination *Pagination `json:"pagination" description:"Pagination information"`
		Total      int64       `json:"total" example:"150"`
	} `json:"meta"`
}

// HealthResponse represents health check response
// @Description Health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy" description:"Overall health status"`
	Services  map[string]string `json:"services" description:"Individual service health status"`
	Timestamp string            `json:"timestamp" example:"2023-12-01T10:30:00Z" description:"Health check timestamp"`
	Version   string            `json:"version" example:"1.0.0" description:"Application version"`
}