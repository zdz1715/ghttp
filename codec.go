package ghttp

import (
	"net/http"
	"sync"

	"github.com/zdz1715/ghttp/encoding/yaml"

	"github.com/zdz1715/ghttp/encoding/xml"

	"github.com/zdz1715/ghttp/encoding"
	"github.com/zdz1715/ghttp/encoding/json"
	"github.com/zdz1715/ghttp/encoding/proto"
)

var defaultContentType = newDefaultContentType()

type contentType struct {
	subType map[string]string
	mu      sync.RWMutex
}

func newDefaultContentType() *contentType {
	return &contentType{
		subType: map[string]string{
			// default: json
			"*": json.Name,

			"json":       json.Name,
			"x-protobuf": proto.Name,
			"xml":        xml.Name,
			"x-yaml":     yaml.Name,
			"yaml":       yaml.Name,
		},
	}
}

func (c *contentType) Set(name string, cname string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subType[name] = cname
}

func (c *contentType) Get(name string) encoding.Codec {
	return encoding.GetCodec(c.subType[name])
}

func RegisterCodecNameByContentType(contentType string, name string) {
	if name == "" {
		return
	}
	defaultContentType.Set(ContentSubtype(contentType), name)
}

func RegisterCodecByContentType(contentType string, codec encoding.Codec) {
	if codec == nil {
		return
	}
	encoding.RegisterCodec(codec)
	defaultContentType.Set(ContentSubtype(contentType), codec.Name())
}

func GetCodecByContentType(contentType string) encoding.Codec {
	return defaultContentType.Get(ContentSubtype(contentType))
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(r *http.Request, name ...string) (encoding.Codec, bool) {
	headerName := "Content-Type"
	if len(name) > 0 && name[0] != "" {
		headerName = name[0]
	}
	for _, accept := range r.Header[headerName] {
		codec := GetCodecByContentType(accept)
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec(json.Name), false
}

// CodecForResponse get encoding.Codec via http.Response
func CodecForResponse(r *http.Response, name ...string) (encoding.Codec, bool) {
	headerName := "Content-Type"
	if len(name) > 0 && name[0] != "" {
		headerName = name[0]
	}
	for _, accept := range r.Header[headerName] {
		codec := GetCodecByContentType(accept)
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec(json.Name), false
}
