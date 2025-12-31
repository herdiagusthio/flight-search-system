package response

// Error codes used in API responses.
const (
	CodeInvalidRequest     = "invalid_request"
	CodeValidationError    = "validation_error"
	CodeServiceUnavailable = "service_unavailable"
	CodeTimeout            = "timeout"
	CodeInternalError      = "internal_error"
)

// Error messages used in API responses.
const (
	MsgInvalidRequestBody = "Failed to parse request body"
	MsgValidationFailed   = "Request validation failed"
	MsgServiceUnavailable = "All flight providers are currently unavailable"
	MsgTimeout            = "Request timed out"
	MsgRequestCancelled   = "Request was cancelled"
	MsgInternalError      = "An unexpected error occurred"
)

// ErrorDetail represents a standardized error response.
type ErrorDetail struct {
	Code    string            `json:"code" example:"validation_error"`                     // Error code identifier
	Message string            `json:"message" example:"Request validation failed"`         // Human-readable error message
	Details map[string]string `json:"details,omitempty" example:"field:origin,reason:required"` // Additional error details (optional)
}

