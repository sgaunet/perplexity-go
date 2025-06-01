package perplexity

import (
	"testing"
)

func TestParseErrorMessage(t *testing.T) {
	t.Run("ValidErrorResponse", func(t *testing.T) {
		data := []byte(`{"error": {"message": "An error occurred", "type": "APIError", "code": 400}}`)
		err := ParseErrorMessage(data)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		responseError, ok := err.(*ResponseError)
		if !ok {
			t.Fatalf("expected error of type *ResponseError, got %T", err)
		}

		if responseError.ErrorData.Message != "An error occurred" {
			t.Errorf("expected message 'An error occurred', got '%s'", responseError.ErrorData.Message)
		}
		if responseError.ErrorData.Type != "APIError" {
			t.Errorf("expected type 'APIError', got '%s'", responseError.ErrorData.Type)
		}
		if responseError.ErrorData.Code != 400 {
			t.Errorf("expected code 400, got %d", responseError.ErrorData.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`{"error": {"message": "An error occurred", "type": "APIError", "code": "invalid"}}`)
		err := ParseErrorMessage(data)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		if _, ok := err.(*ErrorResponse); ok {
			t.Fatalf("expected a JSON unmarshal error, got *ErrorResponse")
		}
	})

	t.Run("EmptyData", func(t *testing.T) {
		data := []byte(``)
		err := ParseErrorMessage(data)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		if _, ok := err.(*ErrorResponse); ok {
			t.Fatalf("expected a JSON unmarshal error, got *ErrorResponse")
		}
	})

	t.Run("NilData", func(t *testing.T) {
		var data []byte
		err := ParseErrorMessage(data)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		if _, ok := err.(*ErrorResponse); ok {
			t.Fatalf("expected a JSON unmarshal error, got *ErrorResponse")
		}
	})
}
