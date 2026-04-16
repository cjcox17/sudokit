package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// APIError represents a structured error response for API clients.
// It includes machine-readable codes, user-friendly messages, and optional details.
type APIError struct {
	Code        string `json:"code"`                // e.g., "ADDRESS_NOT_FOUND", "VALIDATION_FAILED"
	Message     string `json:"message"`             // Technical/internal message for debugging
	UserMessage string `json:"userMessage"`         // Friendly message shown to end users
	Details     any    `json:"details,omitempty"`   // Optional: field errors, additional context
	RequestID   string `json:"requestId,omitempty"` // Request ID for debugging/tracing
	HTTPStatus  int    `json:"-"`                   // HTTP status code (not serialized)
	err         error  `json:"-"`                   // Wrapped underlying error (not serialized)
}

// Error implements the error interface.
func (e APIError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.err)
	}
	return e.Message
}

// Unwrap allows errors.As and errors.Is to work with wrapped errors.
func (e APIError) Unwrap() error {
	return e.err
}

// WithDetails returns a new APIError with the given details.
func (e APIError) WithDetails(details any) APIError {
	e.Details = details
	return e
}

// WithRequestID sets the request ID on the error.
func (e APIError) WithRequestID(requestID string) APIError {
	e.RequestID = requestID
	return e
}

// Wrap wraps an underlying error, preserving the APIError structure.
func (e APIError) Wrap(err error) APIError {
	e.err = err
	if e.Message == "" && err != nil {
		e.Message = err.Error()
	}
	return e
}

// IsAPIError checks if an error is an APIError (or wraps one).
func IsAPIError(err error) bool {
	var apiErr APIError
	return errors.As(err, &apiErr)
}

// New creates a new APIError with the given parameters.
func New(code, message, userMessage string, status int) APIError {
	return APIError{
		Code:        code,
		Message:     message,
		UserMessage: userMessage,
		HTTPStatus:  status,
	}
}

// Common predefined API errors
var (
	// 400 - Bad Request
	ErrBadRequest = APIError{
		Code:        "BAD_REQUEST",
		Message:     "invalid request",
		UserMessage: "The request could not be understood. Please check your input and try again.",
		HTTPStatus:  http.StatusBadRequest,
	}

	ErrValidationFailed = APIError{
		Code:        "VALIDATION_FAILED",
		Message:     "validation failed",
		UserMessage: "Some fields are invalid. Please check your input and try again.",
		HTTPStatus:  http.StatusBadRequest,
	}

	ErrInvalidJSON = APIError{
		Code:        "INVALID_JSON",
		Message:     "invalid JSON in request body",
		UserMessage: "The request body contains invalid JSON. Please check your input and try again.",
		HTTPStatus:  http.StatusBadRequest,
	}

	// 401 - Unauthorized
	ErrUnauthorized = APIError{
		Code:        "UNAUTHORIZED",
		Message:     "authentication required",
		UserMessage: "You need to sign in to access this resource.",
		HTTPStatus:  http.StatusUnauthorized,
	}

	ErrInvalidCredentials = APIError{
		Code:        "INVALID_CREDENTIALS",
		Message:     "invalid email or password",
		UserMessage: "Invalid email or password. Please try again.",
		HTTPStatus:  http.StatusUnauthorized,
	}

	// 403 - Forbidden
	ErrForbidden = APIError{
		Code:        "FORBIDDEN",
		Message:     "access denied",
		UserMessage: "You do not have permission to access this resource.",
		HTTPStatus:  http.StatusForbidden,
	}

	// 404 - Not Found
	ErrNotFound = APIError{
		Code:        "NOT_FOUND",
		Message:     "resource not found",
		UserMessage: "The requested resource could not be found.",
		HTTPStatus:  http.StatusNotFound,
	}

	// 409 - Conflict
	ErrConflict = APIError{
		Code:        "CONFLICT",
		Message:     "resource conflict",
		UserMessage: "The request conflicts with the current state of the resource.",
		HTTPStatus:  http.StatusConflict,
	}

	ErrDuplicate = APIError{
		Code:        "DUPLICATE",
		Message:     "resource already exists",
		UserMessage: "This resource already exists. Please try a different value.",
		HTTPStatus:  http.StatusConflict,
	}

	// 422 - Unprocessable Entity
	ErrUnprocessable = APIError{
		Code:        "UNPROCESSABLE_ENTITY",
		Message:     "unable to process request",
		UserMessage: "We couldn't process your request. Please check your information and try again.",
		HTTPStatus:  http.StatusUnprocessableEntity,
	}

	// 429 - Too Many Requests
	ErrTooManyRequests = APIError{
		Code:        "TOO_MANY_REQUESTS",
		Message:     "rate limit exceeded",
		UserMessage: "Too many requests. Please wait a moment and try again.",
		HTTPStatus:  http.StatusTooManyRequests,
	}

	// 500 - Internal Server Error
	ErrInternalServer = APIError{
		Code:        "INTERNAL_ERROR",
		Message:     "internal server error",
		UserMessage: "Something went wrong on our end. Please try again later or contact support if the problem persists.",
		HTTPStatus:  http.StatusInternalServerError,
	}

	// 503 - Service Unavailable
	ErrServiceUnavailable = APIError{
		Code:        "SERVICE_UNAVAILABLE",
		Message:     "service temporarily unavailable",
		UserMessage: "The service is temporarily unavailable. Please try again later.",
		HTTPStatus:  http.StatusServiceUnavailable,
	}
)

// GetRequestID extracts the request ID from Fiber context locals.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("request_id").(string); ok {
		return id
	}
	return ""
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FromStatus creates an appropriate APIError from an HTTP status code.
// Useful when you only have a status code and need to generate a proper error.
func FromStatus(status int) APIError {
	switch status {
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrConflict
	case http.StatusUnprocessableEntity:
		return ErrUnprocessable
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusInternalServerError:
		return ErrInternalServer
	case http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	default:
		return ErrInternalServer.WithDetails(map[string]int{"status": status})
	}
}

// MarshalJSON implements custom JSON marshaling to ensure consistent formatting.
func (e APIError) MarshalJSON() ([]byte, error) {
	type Alias APIError
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&e),
	})
}
