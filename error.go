package ghttp

import (
	"errors"
	"strconv"
	"strings"
)

type Not2xxError interface {
	String() string
}

type HTTPNot2xxError struct {
	Method     string
	StatusCode int
	Err        Not2xxError
}

func (h HTTPNot2xxError) Error() string {
	var buf strings.Builder

	if h.Method != "" {
		buf.WriteString(h.Method)
		buf.WriteByte(' ')
	}

	buf.WriteString(strconv.Itoa(h.StatusCode))
	if h.Err != nil {
		buf.WriteString(": ")
		buf.WriteString(h.Err.String())
	}
	return buf.String()
}

func IsHTTPNot2xxError(err error) bool {
	if err == nil {
		return false
	}
	var e *HTTPNot2xxError
	ok := errors.As(err, &e)
	return ok
}

func ConvertToHTTPNot2xxError(err error) (*HTTPNot2xxError, bool) {
	if err == nil {
		return nil, false
	}
	var e *HTTPNot2xxError
	ok := errors.As(err, &e)
	return e, ok
}
