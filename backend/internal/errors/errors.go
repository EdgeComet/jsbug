package errors

import (
	"fmt"
	"net/http"
)

// Error codes
const (
	CodeInvalidURL       = "INVALID_URL"
	CodeInvalidTimeout   = "INVALID_TIMEOUT"
	CodeInvalidWaitEvent = "INVALID_WAIT_EVENT"
	CodeRenderTimeout    = "RENDER_TIMEOUT"
	CodeRenderFailed     = "RENDER_FAILED"
	CodeFetchFailed      = "FETCH_FAILED"
	CodeChromeFailed     = "CHROME_UNAVAILABLE"
	CodeInternalError    = "INTERNAL_ERROR"
)

// AppError is the base application error type
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Cause      error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ValidationError represents a request validation error (400)
type ValidationError struct {
	AppError
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string) *ValidationError {
	return &ValidationError{
		AppError: AppError{
			Code:       code,
			Message:    message,
			HTTPStatus: http.StatusBadRequest,
		},
	}
}

// InvalidURL creates an invalid URL validation error
func InvalidURL(message string) *ValidationError {
	return NewValidationError(CodeInvalidURL, message)
}

// InvalidTimeout creates an invalid timeout validation error
func InvalidTimeout(message string) *ValidationError {
	return NewValidationError(CodeInvalidTimeout, message)
}

// InvalidWaitEvent creates an invalid wait event validation error
func InvalidWaitEvent(message string) *ValidationError {
	return NewValidationError(CodeInvalidWaitEvent, message)
}

// TimeoutError represents a timeout error (408)
type TimeoutError struct {
	AppError
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, cause error) *TimeoutError {
	return &TimeoutError{
		AppError: AppError{
			Code:       CodeRenderTimeout,
			Message:    message,
			HTTPStatus: http.StatusRequestTimeout,
			Cause:      cause,
		},
	}
}

// RenderTimeout creates a render timeout error
func RenderTimeout(cause error) *TimeoutError {
	return NewTimeoutError("Render timeout exceeded", cause)
}

// FetchTimeout creates a fetch timeout error
func FetchTimeout(cause error) *TimeoutError {
	return NewTimeoutError("Fetch timeout exceeded", cause)
}

// RenderError represents a rendering error (500)
type RenderError struct {
	AppError
}

// NewRenderError creates a new render error
func NewRenderError(message string, cause error) *RenderError {
	return &RenderError{
		AppError: AppError{
			Code:       CodeRenderFailed,
			Message:    message,
			HTTPStatus: http.StatusInternalServerError,
			Cause:      cause,
		},
	}
}

// FetchError represents a fetch error (500)
type FetchError struct {
	AppError
}

// NewFetchError creates a new fetch error
func NewFetchError(message string, cause error) *FetchError {
	return &FetchError{
		AppError: AppError{
			Code:       CodeFetchFailed,
			Message:    message,
			HTTPStatus: http.StatusInternalServerError,
			Cause:      cause,
		},
	}
}

// ChromeError represents a Chrome unavailability error (503)
type ChromeError struct {
	AppError
}

// NewChromeError creates a new Chrome error
func NewChromeError(message string) *ChromeError {
	return &ChromeError{
		AppError: AppError{
			Code:       CodeChromeFailed,
			Message:    message,
			HTTPStatus: http.StatusServiceUnavailable,
		},
	}
}

// ChromeUnavailable creates a Chrome unavailable error
func ChromeUnavailable() *ChromeError {
	return NewChromeError("Chrome renderer is not available")
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	if valErr, ok := err.(*ValidationError); ok {
		return valErr.HTTPStatus
	}
	if toErr, ok := err.(*TimeoutError); ok {
		return toErr.HTTPStatus
	}
	if renErr, ok := err.(*RenderError); ok {
		return renErr.HTTPStatus
	}
	if fetErr, ok := err.(*FetchError); ok {
		return fetErr.HTTPStatus
	}
	if chrErr, ok := err.(*ChromeError); ok {
		return chrErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetCode returns the error code for an error
func GetCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	if valErr, ok := err.(*ValidationError); ok {
		return valErr.Code
	}
	if toErr, ok := err.(*TimeoutError); ok {
		return toErr.Code
	}
	if renErr, ok := err.(*RenderError); ok {
		return renErr.Code
	}
	if fetErr, ok := err.(*FetchError); ok {
		return fetErr.Code
	}
	if chrErr, ok := err.(*ChromeError); ok {
		return chrErr.Code
	}
	return CodeInternalError
}

// GetMessage returns the user-friendly message for an error
func GetMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}
	if valErr, ok := err.(*ValidationError); ok {
		return valErr.Message
	}
	if toErr, ok := err.(*TimeoutError); ok {
		return toErr.Message
	}
	if renErr, ok := err.(*RenderError); ok {
		return renErr.Message
	}
	if fetErr, ok := err.(*FetchError); ok {
		return fetErr.Message
	}
	if chrErr, ok := err.(*ChromeError); ok {
		return chrErr.Message
	}
	return err.Error()
}
