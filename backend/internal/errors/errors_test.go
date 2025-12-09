package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError(CodeInvalidURL, "URL is required")

	if err.Code != CodeInvalidURL {
		t.Errorf("Code = %s, want %s", err.Code, CodeInvalidURL)
	}
	if err.Message != "URL is required" {
		t.Errorf("Message = %s", err.Message)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusBadRequest)
	}
}

func TestValidationErrorHelpers(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		wantCode string
	}{
		{"InvalidURL", InvalidURL("test"), CodeInvalidURL},
		{"InvalidTimeout", InvalidTimeout("test"), CodeInvalidTimeout},
		{"InvalidWaitEvent", InvalidWaitEvent("test"), CodeInvalidWaitEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %s, want %s", tt.err.Code, tt.wantCode)
			}
			if tt.err.HTTPStatus != http.StatusBadRequest {
				t.Errorf("HTTPStatus = %d, want %d", tt.err.HTTPStatus, http.StatusBadRequest)
			}
		})
	}
}

func TestTimeoutError(t *testing.T) {
	cause := errors.New("context deadline exceeded")
	err := NewTimeoutError("Request timed out", cause)

	if err.Code != CodeRenderTimeout {
		t.Errorf("Code = %s, want %s", err.Code, CodeRenderTimeout)
	}
	if err.HTTPStatus != http.StatusRequestTimeout {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusRequestTimeout)
	}
	if err.Cause != cause {
		t.Error("Cause not set correctly")
	}
}

func TestTimeoutErrorHelpers(t *testing.T) {
	cause := errors.New("deadline exceeded")

	renderErr := RenderTimeout(cause)
	if renderErr.HTTPStatus != http.StatusRequestTimeout {
		t.Errorf("RenderTimeout HTTPStatus = %d", renderErr.HTTPStatus)
	}

	fetchErr := FetchTimeout(cause)
	if fetchErr.HTTPStatus != http.StatusRequestTimeout {
		t.Errorf("FetchTimeout HTTPStatus = %d", fetchErr.HTTPStatus)
	}
}

func TestRenderError(t *testing.T) {
	cause := errors.New("navigation failed")
	err := NewRenderError("Failed to render page", cause)

	if err.Code != CodeRenderFailed {
		t.Errorf("Code = %s, want %s", err.Code, CodeRenderFailed)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusInternalServerError)
	}
}

func TestFetchError(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewFetchError("Failed to fetch page", cause)

	if err.Code != CodeFetchFailed {
		t.Errorf("Code = %s, want %s", err.Code, CodeFetchFailed)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusInternalServerError)
	}
}

func TestChromeError(t *testing.T) {
	err := NewChromeError("Chrome not available")

	if err.Code != CodeChromeFailed {
		t.Errorf("Code = %s, want %s", err.Code, CodeChromeFailed)
	}
	if err.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusServiceUnavailable)
	}
}

func TestChromeUnavailable(t *testing.T) {
	err := ChromeUnavailable()

	if err.Code != CodeChromeFailed {
		t.Errorf("Code = %s, want %s", err.Code, CodeChromeFailed)
	}
	if err.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusServiceUnavailable)
	}
}

func TestAppError_Error(t *testing.T) {
	// Without cause
	err1 := &AppError{Code: "TEST", Message: "test message"}
	expected1 := "TEST: test message"
	if err1.Error() != expected1 {
		t.Errorf("Error() = %s, want %s", err1.Error(), expected1)
	}

	// With cause
	cause := errors.New("underlying error")
	err2 := &AppError{Code: "TEST", Message: "test message", Cause: cause}
	if err2.Error() == expected1 {
		t.Error("Error() should include cause")
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &AppError{Code: "TEST", Message: "test", Cause: cause}

	if err.Unwrap() != cause {
		t.Error("Unwrap() should return cause")
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"ValidationError", InvalidURL("test"), http.StatusBadRequest},
		{"TimeoutError", RenderTimeout(nil), http.StatusRequestTimeout},
		{"RenderError", NewRenderError("test", nil), http.StatusInternalServerError},
		{"FetchError", NewFetchError("test", nil), http.StatusInternalServerError},
		{"ChromeError", ChromeUnavailable(), http.StatusServiceUnavailable},
		{"Unknown error", errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := GetHTTPStatus(tt.err)
			if status != tt.status {
				t.Errorf("GetHTTPStatus() = %d, want %d", status, tt.status)
			}
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{"ValidationError", InvalidURL("test"), CodeInvalidURL},
		{"TimeoutError", RenderTimeout(nil), CodeRenderTimeout},
		{"RenderError", NewRenderError("test", nil), CodeRenderFailed},
		{"FetchError", NewFetchError("test", nil), CodeFetchFailed},
		{"ChromeError", ChromeUnavailable(), CodeChromeFailed},
		{"Unknown error", errors.New("unknown"), CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetCode(tt.err)
			if code != tt.code {
				t.Errorf("GetCode() = %s, want %s", code, tt.code)
			}
		})
	}
}

func TestGetMessage(t *testing.T) {
	err := InvalidURL("URL is required")
	msg := GetMessage(err)
	if msg != "URL is required" {
		t.Errorf("GetMessage() = %s, want %s", msg, "URL is required")
	}

	// Unknown error returns error string
	unknownErr := errors.New("unknown error")
	msg = GetMessage(unknownErr)
	if msg != "unknown error" {
		t.Errorf("GetMessage() = %s, want %s", msg, "unknown error")
	}
}
