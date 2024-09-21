package query

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var timeType = reflect.TypeOf(time.Time{})

// Values returns the url.Values encoding of v.
//
// Values expects to be passed a struct, and traverses it recursively using the
// following encoding rules.
//
// Each exported struct field is encoded as a URL parameter unless
//
//   - the field's tag is "-", or
//   - the field is empty and its tag specifies the "omitempty" option
//
// The empty values are false, 0, any nil pointer or interface value, any array
// slice, map, or string of length zero, and any type (such as time.Time) that
// returns true for IsZero().
//
// The URL parameter name defaults to the struct field name but can be
// specified in the struct field's tag value.  The "query" key in the struct
// field's tag value is the key name, followed by an optional comma and
// options.  For example:
//
//	// Field is ignored by this package.
//	Field int `query:"-"`
//
//	// Field appears as URL parameter "myName".
//	Field int `query:"myName"`
//
//	// Field appears as URL parameter "myName" and the field is omitted if
//	// its value is empty
//	Field int `query:"myName,omitempty"`
//
//	// Field appears as URL parameter "Field" (the default), but the field
//	// is skipped if empty.  Note the leading comma.
//	Field int `query:",omitempty"`
//
// For encoding individual field values, the following type-dependent rules
// apply:
//
// Boolean values default to encoding as the strings "true" or "false".
// Including the "int" option signals that the field should be encoded as the
// strings "1" or "0".
//
// time.Time values default to encoding as RFC3339 timestamps.  Including the
// "unix" option signals that the field should be encoded as a Unix time (see
// time.Unix()).  The "unixmilli" and "unixnano" options will encode the number
// of milliseconds and nanoseconds, respectively, since January 1, 1970 (see
// time.UnixNano()).  Including the "layout" struct tag (separate from the
// "query" tag) will use the value of the "layout" tag as a layout passed to
// time.Format.  For example:
//
//	// Encode a time.Time as YYYY-MM-DD HH:ii:ss
//	Field time.Time `query:,time_format:"2006-01-02 15:04:05"`
//
// Slice and Array values default to encoding as multiple URL values of the
// same name.  Including the "comma" option signals that the field should be
// encoded as a single comma-delimited value.  Including the "space" option
// similarly encodes the value as a single space-delimited string. Including
// the "semicolon" option will encode the value as a semicolon-delimited string.
// Including the "brackets" option signals that the multiple URL values should
// have "[]" appended to the value name. "numbered" will append a number to
// the end of each incidence of the value name, example:
// name0=value0&name1=value1, etc.  Including the "del" struct tag (separate
// from the "query" tag) will use the value of the "del" tag as the delimiter.
// For example:
//
//	// Encode a slice of bools as ints ("1" for true, "0" for false),
//	// separated by exclamation points "!".
//	Field []bool `query:",int" del:"!"`
//
// Anonymous struct fields are usually encoded as if their inner exported
// fields were fields in the outer struct, subject to the standard Go
// visibility rules.  An anonymous struct field with a name given in its URL
// tag is treated as having that name, rather than being anonymous.
//
// Non-nil pointer values are encoded as the value pointed to.
//
// Nested structs have their fields processed recursively and are encoded
// including parent fields in value names for scoping. For example,
//
//	"user[name]=acme&user[addr][postcode]=1234&user[addr][city]=SFO"
//
// All other values are encoded using their default string representation.
//
// Multiple fields that encode to the same URL parameter name will be included
// as multiple URL values of the same name.
func Values(v interface{}) (url.Values, error) {
	values := make(url.Values)

	if v == nil {
		return values, nil
	}

	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return values, nil
		}
		val = val.Elem()
	}

	switch str := val.Interface().(type) {
	case string:
		return parseQueryString(str)
	case []byte:
		queryString := unsafe.String(unsafe.SliceData(str), len(str))
		return parseQueryString(queryString)
	}

	err := reflectValue(values, val)
	return values, err
}

func parseQueryString(queryString string) (url.Values, error) {
	return url.ParseQuery(strings.TrimLeft(queryString, "?"))
}

func reflectValue(values url.Values, val reflect.Value) error {
	switch val.Kind() {
	case reflect.Map:
		return reflectMap(values, val)
	case reflect.Slice, reflect.Array:
		l := val.Len()
		if l == 0 {
			return nil
		}
		for i := 0; i < l; i += 2 {
			endIndex := i + 1

			if endIndex > l-1 {
				continue
			}

			key := valueString(val.Index(i), nil)
			values.Add(key, valueString(val.Index(endIndex), nil))
		}
	case reflect.Struct:
		return reflectStruct(values, val, "")
	default:
		return fmt.Errorf("query: Values() unsupported kind input. Got %v", val.Kind())
	}
	return nil
}

func reflectStruct(values url.Values, val reflect.Value, scope string) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}

		sv := val.Field(i)
		tag := sf.Tag.Get(Tag)

		if tag == "-" {
			continue
		}

		fieldName, opts := parseTag(tag)

		name := fieldName
		if name == "" {
			name = sf.Name
		}

		if scope != "" {
			name = scope + "[" + name + "]"
		}

		if opts.Contains(OmitemptyTagOpt) && isEmptyValue(sv) {
			continue
		}

		// recursively dereference pointers. break on nil pointers
		for sv.Kind() == reflect.Ptr {
			if sv.IsNil() {
				break
			}
			sv = sv.Elem()
		}

		if sv.Kind() == reflect.Interface {
			sv = sv.Elem()
		}

		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
			l := sv.Len()
			if l == 0 {
				// skip if slice or array is empty
				continue
			}

			var del string
			if delValue := opts.Get(DelTagOpt); delValue != "" {
				switch delValue {
				case "comma":
					del = ","
				case "space":
					del = " "
				case "semicolon":
					del = ";"
				case "brackets":
					name = name + "[]"
				default:
					del = delValue
				}
			}
			if del != "" {
				s := new(bytes.Buffer)
				first := true
				for j := 0; j < l; j++ {
					if first {
						first = false
					} else {
						s.WriteString(del)
					}

					s.WriteString(valueString(sv.Index(j), opts))
				}
				values.Add(name, s.String())
			} else {
				for j := 0; j < l; j++ {
					values.Add(name, valueString(sv.Index(j), opts))
				}
			}
			continue
		}

		if sv.Type() == timeType {
			values.Add(name, valueString(sv, opts))
			continue
		}

		if sv.Kind() == reflect.Struct {
			if ok := opts.Contains(InlineTagOpt); fieldName == "" && ok {
				if err := reflectStruct(values, sv, ""); err != nil {
					return err
				}
			} else {
				if err := reflectStruct(values, sv, name); err != nil {
					return err
				}
			}
			continue
		}

		values.Add(name, valueString(sv, opts))
	}

	return nil
}

// isEmptyValue checks if a value should be considered empty for the purposes
// of omitting fields with the "omitempty" option.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}

	type zeroable interface {
		IsZero() bool
	}

	if z, ok := v.Interface().(zeroable); ok {
		return z.IsZero()
	}

	return false
}

func reflectMap(values url.Values, val reflect.Value) error {
	iter := val.MapRange()
	for iter.Next() {
		key := valueString(iter.Key(), nil)
		sv := iter.Value()

		// interface{}
		if sv.Kind() == reflect.Interface {
			sv = sv.Elem()
		}

		switch sv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < sv.Len(); i++ {
				values.Add(key, valueString(sv.Index(i), nil))
			}
		default:
			values.Add(key, valueString(sv, nil))
		}

	}
	return nil
}

// valueString returns the string representation of a value
func valueString(v reflect.Value, opts tagOptions) string {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// query:"name,int"
	if v.Kind() == reflect.Bool && opts.Contains(IntTagOpt) {
		if v.Bool() {
			return "1"
		}
		return "0"
	}

	if v.Type() == timeType {
		t := v.Interface().(time.Time)
		if t.IsZero() {
			return ""
		}
		// query:"create_time,unix"
		if opts.Contains(UnixTagOpt) {
			return strconv.FormatInt(t.Unix(), 10)
		}
		// query:"create_time,unixmilli"
		if opts.Contains(UnixmilliTagOpt) {
			return strconv.FormatInt((t.UnixNano() / 1e6), 10)
		}
		// query:"create_time,unixnano"
		if opts.Contains(UnixnanoTagOpt) {
			return strconv.FormatInt(t.UnixNano(), 10)
		}

		// query:"create_time,time_format:2006-01-02 15:04:05"
		if layout := opts.Get(TimeFormatTagOpt); layout != "" {
			return t.Format(layout)
		}

		return t.Format(time.RFC3339)
	}

	// bytes to string
	if b, ok := v.Interface().([]byte); ok {
		return unsafe.String(unsafe.SliceData(b), len(b))
	}

	return fmt.Sprint(v.Interface())
}

// tagOptions is the string following a comma in a struct field's "query" tag, or
// the empty string. It does not include the leading comma.
type tagOptions map[string]string

// parseTag splits a struct field's url tag into its name and comma-separated
// options.
func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	opts := s[1:]
	tagOpts := make(tagOptions)
	if len(opts) > 0 {
		for _, v := range opts {
			if v == "" {
				continue
			}
			keys := strings.Split(v, ":")
			if len(keys) == 0 {
				continue
			}
			tagOpts[keys[0]] = strings.Join(keys[1:], ":")
		}
	}
	return s[0], tagOpts
}

// Contains checks whether the tagOptions contains the specified option.
func (o tagOptions) Contains(option string) bool {
	_, ok := o[option]
	return ok
}

func (o tagOptions) Get(option string) string {
	return o[option]
}
