package sheets

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// ResourceError is a Client type error.
// It returns from Client's method when server replies with an error.
// It holds the HTTP Method, URL, Status Code and the actual error message came from server.
//
// See `IsResourceError` and `IsStatusError` too.
type ResourceError struct {
	Method     string
	URL        string
	StatusCode int
	Message    string
}

func newResourceError(resp *http.Response) *ResourceError {
	cause := "unspecified"

	if resp.Body != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			cause = string(b)
		}
	}

	endpoint := resp.Request.URL.String()
	return &ResourceError{
		Method:     resp.Request.Method,
		URL:        endpoint,
		StatusCode: resp.StatusCode,
		Message:    cause,
	}
}

// Error implements a Go error and returns a human-readable error text.
func (e *ResourceError) Error() string {
	return fmt.Sprintf("resource error [%s: %s]: %d: %s", e.Method, e.URL, e.StatusCode, e.Message)
}

// IsStatusError reports whether a "target" error is type of `ResourceError` and the status code is the provided "statusCode" one.
// Usage:
// resErr, ok := IsStatusError(http.StatusNotFound, err)
//
//	if ok {
//		[ressErr.Method, URL, StatusCode, Message...]
//	}
//
// See `IsResourceError` too.
func IsStatusError(statusCode int, target error) (*ResourceError, bool) {
	if target == nil {
		return nil, false
	}

	t, ok := target.(*ResourceError)
	if !ok {
		return nil, false
	}

	return t, t.StatusCode == statusCode
}

// IsResourceError reports whether "target" is "e" ResourceError.
// Returns true when all fields of "e" are equal to "target" fields
// or when a "target" matching field is empty.
func IsResourceError(e *ResourceError, target error) bool {
	if target == nil {
		return e == nil
	}

	t, ok := target.(*ResourceError)
	if !ok {
		return false
	}

	return (e.Method == t.Method || t.Method == "") &&
		(e.URL == t.URL || t.URL == "") &&
		(e.StatusCode == t.StatusCode || t.StatusCode <= 0) &&
		(e.Message == t.Message || t.Message == "")
}

// Is implements the standard`errors.Is` internal interface.
// It's equivalent of the `IsResourceError` package-level function.
func (e *ResourceError) Is(target error) bool { // implements Go 1.13 errors.Is internal interface.
	return IsResourceError(e, target)
}
