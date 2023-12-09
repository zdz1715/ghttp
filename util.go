package ghttp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/guonaihong/gout/encode"
	"github.com/guonaihong/gout/setting"
	"github.com/zdz1715/go-utils/goutils"
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
	return contentType[left+1 : right]
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
		return goutils.BytesToString(s)
	case string:
		return s
	}
	return ""
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
