package ghttp

import "testing"

func TestGetCodecByContentType(t *testing.T) {
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
			contentType: "application/x-yaml",
			want:        "yaml",
		},
		{
			contentType: "application/x-protobuf",
			want:        "proto",
		},
		{
			contentType: "application/vnd.api+json",
			want:        "json",
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
		target := GetCodecByContentType(v.contentType)
		if target.Name() != v.want {
			t.Logf("ContentSubtype() failed: target=%s want=%s", target.Name(), v.want)
		}
	}
}
