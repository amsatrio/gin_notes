package constant

import "errors"

const (
	ERROR_KEY     = "API_ERROR"
	ERROR_MESSAGE = "API_ERROR_MESSAGE"
)

var (
	ErrorAuthorizationHeaderIsEmpty = errors.New("authorization header is empty")

	ErrorAuthorizationIsEmpty = errors.New("authorization is empty")

	ErrorAuthorizationHeaderIsInvalid = errors.New("authorization header is invalid")

	ErrorAuthorizationTokenExpired = errors.New("authorization token is expired")

	ErrorAuthenticationFailed = errors.New("authentication failed")

	ErrorPermissionDenied = errors.New("permission is denied")

	ErrorUserNotFound = errors.New("user not found")

	ErrorDataNotFound = errors.New("data not found")

	ErrorTooManyRequest = errors.New("too many request")

	ErrorRequestInvalid = errors.New("request invalid")

	ErrorRetrieveDataFailed = errors.New("failed to retrieve data")

	ErrorSaveDataFailed = errors.New("failed to save data")

	ErrorRedisDeleteFailed = errors.New("failed to delete data in redis")
)
