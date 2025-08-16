// Package response provides standardized API response structures and utilities
// for consistent response handling across microservices.
//
// This package offers a unified way to structure API responses with
// success, error, and validation error scenarios.
//
// Example usage:
//
//	// Success response
//	resp := response.NewSuccessResponse("User created", user)
//
//	// Error response
//	resp := response.NewErrorResponse("User not found")
//
//	// Validation error response
//	resp := response.NewErrorResponseWithValidationErrors("Validation failed",
//		response.ValidationError{Field: "email", Reason: "Required"},
//	)
package response

// Status constants for API responses
const (
	StatusAccept  = "Accepted" // StatusAccept indicates successful operation
	StatusReject  = "Rejected" // StatusReject indicates failed operation
	StatusFailure = "Failed"   // StatusFailure indicates system failure
)

// ValidationError represents a field-level validation error
type ValidationError struct {
	Field  string `json:"field"`  // Field name that failed validation
	Reason string `json:"reason"` // Reason for validation failure
}

// Response represents a standardized API response structure
type Response struct {
	Status  string `json:"status"`            // Status of the operation (Accepted/Rejected/Failed)
	Message string `json:"message,omitempty"` // Human-readable message
	Data    any    `json:"data,omitempty"`    // Response data or validation errors
}

// NewResponse creates a new response with the specified status, message, and data
func NewResponse(status string, message string, data any) Response {
	return Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// NewSuccessResponse creates a successful response with StatusAccept
func NewSuccessResponse(message string, data any) Response {
	return Response{
		Status:  StatusAccept,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response with StatusReject
func NewErrorResponse(message string) Response {
	return Response{
		Status:  StatusReject,
		Message: message,
	}
}

// NewErrorResponseWithValidationErrors creates an error response with validation errors
// The validation errors are included in the Data field
func NewErrorResponseWithValidationErrors(message string, validationErrors ...ValidationError) Response {
	return Response{
		Status:  StatusReject,
		Message: message,
		Data:    validationErrors,
	}
}
