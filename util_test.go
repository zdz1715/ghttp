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
			endpoint: "https://www.baidu.com",
			want:     "https://www.baidu.com/page?limit=0",
		},
		{
			path:     "/page",
			endpoint: "https://www.baidu.com/",
			want:     "https://www.baidu.com/page",
		},
		{
			path:     "/page?limit=0",
			endpoint: "https://www.baidu.com/",
			want:     "https://www.baidu.com/page?limit=0",
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

	for _, v := range tests {
		target := FullPath(v.path, v.endpoint)
		if target != v.want {
			t.Logf("FullPath() failed: target=%s want=%s", target, v.want)
		}
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		args any
		want bool
	}{
		{
			args: nil,
			want: true,
		},
		{
			args: "1",
			want: false,
		},
		{
			args: (*clientOptions)(nil),
			want: true,
		},
		{
			args: (CallOption)(nil),
			want: true,
		},
	}

	for _, v := range tests {
		target := isZero(v.args)
		if target != v.want {
			t.Logf("isZero() failed: target=%t want=%t", target, v.want)
		}
	}
}

func TestEncodeQuery(t *testing.T) {
	tests := []struct {
		args any
		want string
	}{
		// 字符串会去掉前缀'?'
		{
			args: "?field1=1&field2=",
			want: "field1=1&field2=",
		},
		// map
		{
			args: map[string]string{
				"field1": "1",
				"field2": "2",
			},
			want: "field1=1&field2=2",
		},
		// struct tag: query
		{
			args: struct {
				Name string `query:"name"`
				Age  int
			}{
				Name: "tom",
				Age:  2,
			},
			want: "Age=2&name=tom",
		},
		// 字符串默认为空就不传
		// int默认零值也会传参
		// *int默认零值不会传参
		{
			args: struct {
				Name string
				Age  int  `query:"age"`
				Sex  *int `query:"sex"`
			}{},
			want: "age=0",
		},
		// 数组
		{
			args: []string{"field1", "1", "field2", "2"},
			want: "field1=1&field2=2",
		},
		// 单个字段数组
		{
			args: map[string]any{
				"field": []string{
					"1",
					"2",
				},
			},
			want: "field=1&field=2",
		},
	}

	for _, v := range tests {
		target, err := EncodeQuery(v.args)
		if err != nil {
			t.Error(err)
			continue
		}
		if target != v.want {
			t.Logf("EncodeQuery() failed: target=%s want=%s", target, v.want)
		}
	}
}
