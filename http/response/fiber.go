package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// RespondWithError sends a structured error response.
// It logs the full error details server-side and sends a safe response to the client.
func RespondWithError(c *fiber.Ctx, err error) error {
	requestID := GetRequestID(c)

	var apiErr APIError
	if errors.As(err, &apiErr) {
		// Use the structured APIError
		apiErr.RequestID = requestID
	} else {
		// Fallback for unexpected/non-structured errors
		apiErr = ErrInternalServer.Wrap(err)
		apiErr.RequestID = requestID
	}

	// Log the error with all details for debugging
	logError(c, apiErr)

	// Return safe response to client
	return c.Status(apiErr.HTTPStatus).JSON(apiErr)
}

// RespondWithValidationError sends a validation error with field-level details.
func RespondWithValidationError(c *fiber.Ctx, fieldErrors map[string]string) error {
	requestID := GetRequestID(c)

	apiErr := ErrValidationFailed.WithRequestID(requestID).WithDetails(fieldErrors)

	// Log validation failures at Info level since these are client errors
	logValidationError(c, fieldErrors)

	return c.Status(http.StatusBadRequest).JSON(apiErr)
}

// RespondWithValidationErrors is a variant that accepts a slice of validation errors.
// Useful when using go-playground/validator.
func RespondWithValidationErrors(c *fiber.Ctx, errors []ValidationError) error {
	requestID := GetRequestID(c)

	// Convert to map for consistency
	details := make(map[string]string, len(errors))
	for _, e := range errors {
		details[e.Field] = e.Message
	}

	apiErr := ErrValidationFailed.WithRequestID(requestID).WithDetails(details)

	logValidationError(c, details)

	return c.Status(http.StatusBadRequest).JSON(apiErr)
}

// logError logs the full error details for debugging.
func logError(c *fiber.Ctx, apiErr APIError) {
	attrs := []any{
		slog.String("code", apiErr.Code),
		slog.String("message", apiErr.Message),
		slog.String("user_message", apiErr.UserMessage),
		slog.String("request_id", apiErr.RequestID),
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
	}

	if apiErr.Details != nil {
		attrs = append(attrs, slog.Any("details", apiErr.Details))
	}

	// Log wrapped error if present
	if wrapped := apiErr.Unwrap(); wrapped != nil {
		attrs = append(attrs, slog.String("wrapped_error", wrapped.Error()))
	}

	// Log at appropriate level based on HTTP status
	if apiErr.HTTPStatus >= 500 {
		slog.Error("API error", attrs...)
	} else if apiErr.HTTPStatus >= 400 {
		slog.Warn("API client error", attrs...)
	}
}

// logValidationError logs validation failures.
func logValidationError(c *fiber.Ctx, fieldErrors map[string]string) {
	slog.Info("Validation failed",
		slog.String("request_id", GetRequestID(c)),
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
		slog.Any("field_errors", fieldErrors),
	)
}

// RespondWithAPIError sends a single APIError response with proper logging.
func RespondWithAPIError(c *fiber.Ctx, apiErr APIError) error {
	apiErr.RequestID = GetRequestID(c)
	logError(c, apiErr)
	return c.Status(apiErr.HTTPStatus).JSON(apiErr)
}
