package common

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BaseError struct {
	Code    ErrorCode              `json:"code,omitempty"`
	Message string                 `json:"message,omitempty"`
	Cause   error                  `json:"cause,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *BaseError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("error code: %v", e.Code)
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

func (e *BaseError) ToGRPCStatus() *status.Status {
	// Map our error codes to gRPC codes
	grpcCode := mapErrorCodeToGRPCCode(e.Code)

	// Create base status
	st := status.New(grpcCode, e.Message)

	// Build error details
	details := &ErrorDetails{
		Code:    e.Code,
		Message: e.Message,
		Details: make(map[string]string),
	}

	// Convert details map to string map
	for k, v := range e.Details {
		if str, ok := v.(string); ok {
			details.Details[k] = str
		} else {
			// Try to JSON encode non-string values
			if data, err := json.Marshal(v); err == nil {
				details.Details[k] = string(data)
			}
		}
	}

	// Add cause if present
	if e.Cause != nil {
		if baseErr, ok := e.Cause.(*BaseError); ok {
			details.Cause = baseErr.toErrorDetails()
		} else {
			details.Cause = &ErrorDetails{
				Message: e.Cause.Error(),
			}
		}
	}

	// Attach details to status
	if st, err := st.WithDetails(details); err == nil {
		return st
	}

	return st
}

func (e *BaseError) toErrorDetails() *ErrorDetails {
	details := &ErrorDetails{
		Code:    e.Code,
		Message: e.Message,
		Details: make(map[string]string),
	}

	for k, v := range e.Details {
		if str, ok := v.(string); ok {
			details.Details[k] = str
		} else if data, err := json.Marshal(v); err == nil {
			details.Details[k] = string(data)
		}
	}

	if e.Cause != nil {
		if baseErr, ok := e.Cause.(*BaseError); ok {
			details.Cause = baseErr.toErrorDetails()
		} else {
			details.Cause = &ErrorDetails{
				Message: e.Cause.Error(),
			}
		}
	}

	return details
}

func mapErrorCodeToGRPCCode(code ErrorCode) codes.Code {
	switch code {
	case ErrorCode_INVALID_REQUEST, ErrorCode_INVALID_PARAMETER:
		return codes.InvalidArgument
	case ErrorCode_RATE_LIMITED:
		return codes.ResourceExhausted
	case ErrorCode_INTERNAL_ERROR:
		return codes.Internal
	case ErrorCode_TIMEOUT_ERROR:
		return codes.DeadlineExceeded
	case ErrorCode_RANGE_TOO_LARGE:
		return codes.ResourceExhausted
	case ErrorCode_RANGE_OUTSIDE_AVAILABLE:
		return codes.OutOfRange
	case ErrorCode_UNSUPPORTED_METHOD:
		return codes.Unimplemented
	case ErrorCode_UNSUPPORTED_BLOCK_TAG:
		return codes.Unimplemented
	default:
		return codes.Unknown
	}
}

func NewError(code ErrorCode, message string) *BaseError {
	return &BaseError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

func (e *BaseError) WithCause(cause error) *BaseError {
	e.Cause = cause
	return e
}

func (e *BaseError) WithDetail(key string, value interface{}) *BaseError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

func (e *BaseError) WithDetails(details map[string]interface{}) *BaseError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

func FromGRPCStatus(st *status.Status) (*BaseError, bool) {
	for _, detail := range st.Details() {
		if errDetails, ok := detail.(*ErrorDetails); ok {
			return fromErrorDetails(errDetails), true
		}
	}
	return nil, false
}

func fromErrorDetails(details *ErrorDetails) *BaseError {
	e := &BaseError{
		Code:    details.Code,
		Message: details.Message,
		Details: make(map[string]interface{}),
	}

	// Copy string details
	for k, v := range details.Details {
		// Try to parse JSON values
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			e.Details[k] = parsed
		} else {
			e.Details[k] = v
		}
	}

	// Convert cause if present
	if details.Cause != nil {
		e.Cause = fromErrorDetails(details.Cause)
	}

	return e
}

func parseErrorCode(code string) ErrorCode {
	if val, ok := ErrorCode_value[code]; ok {
		return ErrorCode(val)
	}
	return ErrorCode_ERROR_CODE_UNSPECIFIED
}

func ToStatus(err error) *status.Status {
	if err == nil {
		return status.New(codes.OK, "")
	}

	if baseErr, ok := err.(*BaseError); ok {
		return baseErr.ToGRPCStatus()
	}

	// Check if it's already a gRPC status
	if st, ok := status.FromError(err); ok {
		return st
	}

	// Default to internal error
	return status.New(codes.Internal, err.Error())
}
