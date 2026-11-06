package constant

import "testing"

func TestConstants(t *testing.T) {
	if DIR_FILES != "DIR_FILES" {
		t.Errorf("unexpected DIR_FILES: %q", DIR_FILES)
	}
	if ERROR_KEY != "API_ERROR" {
		t.Errorf("unexpected ERROR_KEY: %q", ERROR_KEY)
	}
	if ERROR_MESSAGE != "API_ERROR_MESSAGE" {
		t.Errorf("unexpected ERROR_MESSAGE: %q", ERROR_MESSAGE)
	}
}

func TestErrors(t *testing.T) {
	tests := map[string]error{
		"authorization header is empty":   ErrorAuthorizationHeaderIsEmpty,
		"authorization is empty":          ErrorAuthorizationIsEmpty,
		"authorization header is invalid": ErrorAuthorizationHeaderIsInvalid,
		"authorization token is expired":  ErrorAuthorizationTokenExpired,
		"authentication failed":           ErrorAuthenticationFailed,
		"permission is denied":            ErrorPermissionDenied,
		"user not found":                  ErrorUserNotFound,
		"data not found":                  ErrorDataNotFound,
		"too many request":                ErrorTooManyRequest,
		"request invalid":                 ErrorRequestInvalid,
		"failed to retrieve data":         ErrorRetrieveDataFailed,
		"failed to save data":             ErrorSaveDataFailed,
		"failed to delete data in redis":  ErrorRedisDeleteFailed,
	}
	for msg, err := range tests {
		if err == nil {
			t.Errorf("error %q is nil", msg)
			continue
		}
		if err.Error() != msg {
			t.Errorf("got %q, want %q", err.Error(), msg)
		}
	}
}
