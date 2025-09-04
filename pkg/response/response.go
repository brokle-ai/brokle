package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	
	appErrors "brokle/pkg/errors"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Type    string `json:"type,omitempty"`
}

type Meta struct {
	RequestID  string      `json:"request_id,omitempty"`
	Timestamp  string      `json:"timestamp,omitempty"`
	Version    string      `json:"version,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Total      int64       `json:"total,omitempty"`
}

type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
	HasNext   bool  `json:"has_next"`
	HasPrev   bool  `json:"has_prev"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    getMeta(c),
	})
}

func SuccessWithPagination(c *gin.Context, data interface{}, pagination *Pagination) {
	meta := getMeta(c)
	meta.Pagination = pagination
	
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func SuccessWithStatus(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Meta:    getMeta(c),
	})
}

func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	if meta == nil {
		meta = getMeta(c)
	} else {
		defaultMeta := getMeta(c)
		if meta.RequestID == "" {
			meta.RequestID = defaultMeta.RequestID
		}
		if meta.Timestamp == "" {
			meta.Timestamp = defaultMeta.Timestamp
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
		Meta:    getMeta(c),
	})
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, APIResponse{
		Success: true,
		Meta:    getMeta(c),
	})
}

func Error(c *gin.Context, err error) {
	var statusCode int
	var apiError *APIError

	if appErr, ok := appErrors.IsAppError(err); ok {
		statusCode = appErr.StatusCode
		apiError = &APIError{
			Code:    string(appErr.Type),
			Message: appErr.Message,
			Details: appErr.Details,
			Type:    string(appErr.Type),
		}
	} else {
		statusCode = http.StatusInternalServerError
		apiError = &APIError{
			Code:    string(appErrors.InternalError),
			Message: "Internal server error",
			Details: "",
			Type:    string(appErrors.InternalError),
		}
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   apiError,
		Meta:    getMeta(c),
	})
}

func ErrorWithStatus(c *gin.Context, statusCode int, code, message, details string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: getMeta(c),
	})
}

func BadRequest(c *gin.Context, message, details string) {
	ErrorWithStatus(c, http.StatusBadRequest, string(appErrors.BadRequestError), message, details)
}

func NotFound(c *gin.Context, resource string) {
	ErrorWithStatus(c, http.StatusNotFound, string(appErrors.NotFoundError), resource+" not found", "")
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	ErrorWithStatus(c, http.StatusUnauthorized, string(appErrors.UnauthorizedError), message, "")
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	ErrorWithStatus(c, http.StatusForbidden, string(appErrors.ForbiddenError), message, "")
}

func Conflict(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusConflict, string(appErrors.ConflictError), message, "")
}

func ValidationError(c *gin.Context, message, details string) {
	ErrorWithStatus(c, http.StatusBadRequest, string(appErrors.ValidationError), message, details)
}

func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	ErrorWithStatus(c, http.StatusInternalServerError, string(appErrors.InternalError), message, "")
}

func RateLimit(c *gin.Context, message string) {
	if message == "" {
		message = "Rate limit exceeded"
	}
	ErrorWithStatus(c, http.StatusTooManyRequests, string(appErrors.RateLimitError), message, "")
}

func PaymentRequired(c *gin.Context, message string) {
	if message == "" {
		message = "Payment required"
	}
	ErrorWithStatus(c, http.StatusPaymentRequired, string(appErrors.PaymentRequiredError), message, "")
}

func QuotaExceeded(c *gin.Context, message string) {
	if message == "" {
		message = "Quota exceeded"
	}
	ErrorWithStatus(c, http.StatusTooManyRequests, string(appErrors.QuotaExceededError), message, "")
}

func AIProviderError(c *gin.Context, message string) {
	if message == "" {
		message = "AI provider error"
	}
	ErrorWithStatus(c, http.StatusBadGateway, string(appErrors.AIProviderError), message, "")
}

func ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	ErrorWithStatus(c, http.StatusServiceUnavailable, string(appErrors.ServiceUnavailable), message, "")
}

func getMeta(c *gin.Context) *Meta {
	meta := &Meta{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "v1",
	}
	
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			meta.RequestID = id
		}
	}
	
	if timestamp, exists := c.Get("timestamp"); exists {
		if ts, ok := timestamp.(string); ok {
			meta.Timestamp = ts
		}
	}
	
	return meta
}

func NewPagination(page, pageSize int, total int64) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPage < 1 {
		totalPage = 1
	}

	return &Pagination{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPage,
		HasNext:   page < totalPage,
		HasPrev:   page > 1,
	}
}