package ghttp

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/guonaihong/gout/encode"
	"github.com/guonaihong/gout/setting"
)

func FullPath(endpoint, path string) string {
	if endpoint == "" {
		return path
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, endpoint) {
		return path
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(endpoint, "/"), strings.TrimLeft(path, "/"))
}

func isZero(arg any) bool {
	if arg == nil {
		return true
	}
	val := reflect.ValueOf(arg)
	if val.Kind() == reflect.Pointer {
		return val.IsNil()

	}
	return false
}

func ContentSubtype(contentType string) string {
	if contentType == "" {
		return ""
	}
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	subContentType := contentType[left+1 : right]
	left = strings.Index(subContentType, "+")
	if left >= 0 {
		return subContentType[left+1:]
	}
	return subContentType
}

func toString(v any) string {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}
	switch s := val.Interface().(type) {
	case []byte:
		return bytesToString(s)
	case string:
		return s
	}
	return ""
}

func bytesToString(s []byte) string {
	//return *(*string)(unsafe.Pointer(&b))
	// go 1.20+
	return unsafe.String(unsafe.SliceData(s), len(s))
}

func EncodeQuery(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	query := toString(v)
	if query != "" {
		return strings.TrimLeft(query, "?"), nil
	}
	enc := encode.NewQueryEncode(setting.Setting{
		NotIgnoreEmpty: false,
	})
	if err := encode.Encode(v, enc); err != nil {
		return "", err
	}
	return enc.End(), nil
}

func Not2xxCode(code int) bool {
	return code < 200 || code > 299
}

func ForceHttps(endpoint string) string {
	index := strings.Index(endpoint, "://")
	if index >= 0 {
		endpoint = endpoint[index+3:]
	}
	return fmt.Sprintf("https://%s", endpoint)
}

// CheckResponse
// returns an error (of type *Error) if the response status code is not 2xx.
func checkResponse(response *http.Response, xxError Not2xxError) error {
	if response == nil {
		return errors.New("http: nil Response")
	}

	if xxError == nil || !Not2xxCode(response.StatusCode) {
		return nil
	}
	var buf strings.Builder

	if response.Request != nil {
		buf.WriteString("method=")
		buf.WriteString(response.Request.Method)
		buf.WriteByte(' ')
	}

	buf.WriteString("code=")
	buf.WriteString(strconv.Itoa(response.StatusCode))

	e := xxError.String()

	if e != "" {
		buf.WriteString(" message=")
		buf.WriteString(e)
	}

	return errors.New(buf.String())
}
