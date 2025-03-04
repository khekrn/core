package base

const (
	StatusAccept  = "Accepted"
	StatusReject  = "Rejected"
	StatusFailure = "Failed"
)

type ValidationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewResponse(status string, message string, data any) Response {
	return Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

func NewSuccessResponse(message string, data any) Response {
	return Response{
		Status:  StatusAccept,
		Message: message,
		Data:    data,
	}
}

func NewErrorResponse(message string) Response {
	return Response{
		Status:  StatusReject,
		Message: message,
	}
}

func NewErrorResponseWithValidationErrors(message string, validationErrors ...ValidationError) Response {
	return Response{
		Status:  StatusReject,
		Message: message,
		Data:    validationErrors,
	}
}
