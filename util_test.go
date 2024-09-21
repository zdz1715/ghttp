package ghttp

import (
	"testing"
)

func TestFullPath(t *testing.T) {
	tests := []struct {
		path     string
		endpoint string
		want     string
	}{
		{
			path:     "page",
			endpoint: "https://www.baidu.com",
			want:     "https://www.baidu.com/page",
		},
		{
			path:     "page?limit=0",
			endpoint: "www.baidu.com",
			want:     "http://www.baidu.com/page?limit=0",
		},
		{
			path:     "/page",
			endpoint: "www.baidu.com/",
			want:     "http://www.baidu.com/page",
		},
		{
			path:     "/page?limit=0",
			endpoint: "www.baidu.com/",
			want:     "http://www.baidu.com/page?limit=0",
		},
		{
			path:     "https://www.baidu.com/page?limit=0",
			endpoint: "https://www.baidu.com/",
			want:     "https://www.baidu.com/page?limit=0",
		},
		{
			path:     "http://www.baidu.com/page?limit=0",
			endpoint: "http://www.baidu.com/",
			want:     "http://www.baidu.com/page?limit=0",
		},
		{
			path:     "http://www.baidu.com/page?limit=0",
			endpoint: "http://www.google.com/",
			want:     "http://www.baidu.com/page?limit=0",
		},
	}

	for i, v := range tests {
		target := FullPath(v.endpoint, v.path)
		if target != v.want {
			t.Logf("index: %d, FullPath() failed: target=%s want=%s", i, target, v.want)
		}
	}
}

func TestContentSubtype(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{
			contentType: "application/json",
			want:        "json",
		},
		{
			contentType: "application/xml",
			want:        "xml",
		},
		{
			contentType: "application/x-www-form-urlencoded",
			want:        "x-www-form-urlencoded",
		},
		{
			contentType: "multipart/form-data",
			want:        "form-data",
		},
		{
			contentType: "application/vnd.api+json",
			want:        "json",
		},
		{
			contentType: "multipart/byteranges",
			want:        "byteranges",
		},
		{
			contentType: "application/json; charset=utf-8",
			want:        "json",
		},
		{
			contentType: "application/vnd.docker.distribution.manifest.v2+json; charset=utf-8",
			want:        "json",
		},
	}

	for _, v := range tests {
		target := ContentSubtype(v.contentType)
		if target != v.want {
			t.Logf("ContentSubtype() failed: target=%s want=%s", target, v.want)
		}
	}
}
