package types

import (
	"testing"
)

func TestRobotsRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantErr   bool
		errorCode string
	}{
		{
			name:    "valid http URL",
			url:     "http://example.com/page",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			url:     "https://example.com/page",
			wantErr: false,
		},
		{
			name:    "valid URL with path and query",
			url:     "https://example.com/path/to/page?query=1",
			wantErr: false,
		},
		{
			name:      "empty URL",
			url:       "",
			wantErr:   true,
			errorCode: ErrInvalidURL,
		},
		{
			name:      "missing scheme",
			url:       "example.com/page",
			wantErr:   true,
			errorCode: ErrInvalidURL,
		},
		{
			name:      "ftp scheme not allowed",
			url:       "ftp://example.com/file",
			wantErr:   true,
			errorCode: ErrInvalidURL,
		},
		{
			name:      "file scheme not allowed",
			url:       "file:///etc/passwd",
			wantErr:   true,
			errorCode: ErrInvalidURL,
		},
		{
			name:      "missing host",
			url:       "https:///path",
			wantErr:   true,
			errorCode: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RobotsRequest{URL: tt.url}
			err := req.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if validationErr, ok := err.(*ValidationError); ok {
					if validationErr.Code != tt.errorCode {
						t.Errorf("Validate() error code = %q, want %q", validationErr.Code, tt.errorCode)
					}
				} else {
					t.Errorf("Validate() expected ValidationError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Code:    "TEST_CODE",
		Message: "test message",
	}

	if err.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test message")
	}
}
