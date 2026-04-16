// Package apiresponse provides structured error handling for API responses.
//
// This package implements a standardized error response format that includes:
// - Machine-readable error codes
// - Technical/internal messages for debugging
// - User-friendly messages for display
// - Optional details (e.g., validation field errors)
// - Request ID for tracing
//
// Basic Usage:
//
// In your handlers, replace the old error pattern:
//
//	// OLD:
//	c.Status(404).JSON(fiber.Map{"error": "user not found"})
//
//	// NEW:
//	apiresponse.RespondWithAPIError(c, apiresponse.ErrNotFound.WithDetails(map[string]string{
//	    "resource": "user",
//	}))
//
// Handling Domain Errors:
//
// Domain errors from your domain models are automatically mapped:
//
//	err := user.CanLogin(ip, orgDisabled)
//	if err != nil {
//	    return apiresponse.DomainErrorHandler(c, err)
//	}
//
// The DomainErrorHandler automatically converts domain errors like:
// - domain.ErrUserDeleted → ErrUnauthorized with reason "account_deleted"
// - casefile.ErrAddressNotFound → ErrNotFound with resource "address"
//
// Validation Errors:
//
// For validation errors, collect field errors and send them:
//
//	fieldErrors := map[string]string{
//	    "email": "Invalid email format",
//	    "password": "Password must be at least 8 characters",
//	}
//	return apiresponse.RespondWithValidationError(c, fieldErrors)
//
// Generic Error Handling:
//
// For unexpected errors, use RespondWithError which will:
// - Log the full error with stack trace information
// - Return a safe INTERNAL_ERROR to the client
//
//	if err != nil {
//	    return apiresponse.RespondWithError(c, err)
//	}
//
// Error Response Format:
//
// All errors are returned as JSON with this structure:
//
//	{
//	    "code": "NOT_FOUND",
//	    "message": "resource not found",
//	    "userMessage": "The requested resource could not be found.",
//	    "details": {
//	        "resource": "user"
//	    },
//	    "requestId": "1704740123456-550e8400-e29b-41d4-a716-446655440000"
//	}
//
// Predefined Errors:
//
// Common HTTP errors are predefined:
// - ErrBadRequest (400)
// - ErrValidationFailed (400)
// - ErrInvalidJSON (400)
// - ErrUnauthorized (401)
// - ErrInvalidCredentials (401)
// - ErrForbidden (403)
// - ErrNotFound (404)
// - ErrConflict (409)
// - ErrDuplicate (409)
// - ErrUnprocessable (422)
// - ErrTooManyRequests (429)
// - ErrInternalServer (500)
// - ErrServiceUnavailable (503)
//
// Creating Custom Errors:
//
// You can create custom errors for specific cases:
//
//	customErr := apiresponse.New("CUSTOM_ERROR", "technical details", "User-friendly message", http.StatusBadRequest)
//	return apiresponse.RespondWithAPIError(c, customErr)
//
// Logging:
//
// All error responses are automatically logged via slog:
// - 4xx errors are logged at Warn level
// - 5xx errors are logged at Error level
// - Validation errors are logged at Info level
// - Logs include request_id, code, message, path, and details
//
// Domain Error Mappings:
//
// The following domain errors are automatically mapped:
//
// User domain:
// - ErrUserDeleted → ErrUnauthorized (reason: account_deleted)
// - ErrUserDisabled → ErrUnauthorized (reason: account_disabled)
// - ErrInvalidPassword → ErrInvalidCredentials
// - ErrOrganizationDisabled → ErrForbidden (reason: organization_disabled)
// - ErrInvalidIPAddress → ErrForbidden (reason: ip_not_allowed)
//
// CaseFile domain:
// - ErrCaseFileDeleted → ErrNotFound
// - ErrDebtorNotFound → ErrNotFound
// - ErrAddressNotFound → ErrNotFound
// - ErrPaymentPlanNotFound → ErrNotFound
// - And many more...
//
// See domain_mappings.go for the full list.
package response
