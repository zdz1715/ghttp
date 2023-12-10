# ghttp
golang http客户端

因为有[gout](https://github.com/guonaihong/gout)这样优秀的http客户端，就不重复造轮子了，只是在[gout](https://github.com/guonaihong/gout)基础上封装了一层方法。
此举是为了解决我写第三方sdk的时候可以统一处理参数和返回值，提高开发效率。

## Contents
- [Installation](#Installation)
- [Quick start](#quick-start)
- [Setting](#setting)
- [SDK](#sdk)
## Installation
```shell
go get -u github.com/zdz1715/ghttp@latest
```

## Quick start
### gitlab-demo
> [!NOTE]
> 只需按照下面初次定义好`gitlab`的`Invoke`方法，然后所有请求以后只需要设置不同的请求方法、url、参数、响应结构体就好了
```go
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zdz1715/ghttp"
)

// Error data-validation-and-error-reporting + OAuth error
// GitLab API docs: https://docs.gitlab.com/ee/api/rest/#data-validation-and-error-reporting
type gitlabError struct {
	Message interface{} `json:"message"`

	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *gitlabError) String() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	if e.Error != "" {
		return e.Error
	}
	if e.Message != nil {
		switch msg := e.Message.(type) {
		case string:
			return msg
		default:
			b, _ := json.Marshal(e.Message)
			return string(b)
		}
	}
	return ""
}

func (e *gitlabError) Reset() {
	e.Message = nil
	e.Error = ""
	e.ErrorDescription = ""
}

func main() {
	args := map[string]string{
		"grant_type": "password",
		"client_id":  "app",
	}
	var reply any
	// 	请求 https://gitlab.com/oauth/token
	err := Invoke(context.Background(), http.MethodPost, "/oauth/token", args, &reply)
	if err != nil {
		fmt.Printf("error: %s", err)
	}
	fmt.Printf("%+v", reply)

	args = map[string]string{
		"page": "1",
	}
	// 	请求 https://gitlab.com/api/v4/projects
	err = Invoke(context.Background(), http.MethodGet, "/api/v4/projects", args, &reply)
	if err != nil {
		fmt.Printf("error: %s", err)
	}
	fmt.Printf("%+v", reply)
}

func Invoke(ctx context.Context, method, path string, args, reply any) error {
	not2xxBody := &gitlabError{}
	clientOps := []ghttp.ClientOption{
		ghttp.WithDebug(true),
		ghttp.WithEndpoint("https://gitlab.com"),
		ghttp.WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		}),
		ghttp.WithUserAgent("sdk/gitlab-v0.0.1"),
		ghttp.WithNot2xxError(not2xxBody),
	}

	client, err := ghttp.NewClient(context.Background(), clientOps...)
	if err != nil {
		return err
	}

	callOptions := &ghttp.CallOptions{

		// Authorization header
		Username: "gitlab",
		Password: "password",

		//BearerToken: "gitlab-token",

		BeforeHook: func(request *http.Request) error {
			request.Header.Set("BeforeHook", "BeforeHook")
			return nil
		},
		AfterHook: func(response *http.Response) error {
			return nil
		},
	}

	// get请求把body换成query
	if method == http.MethodGet && args != nil {
		callOptions.Query = args
		args = nil
	}

	_, err = client.Invoke(ctx, method, path, args, reply, callOptions)
	if err != nil {
		return err
	}

	return nil
}


```
## setting
### path
> [!NOTE]
> 设置的路径若是`http://`或`https://`开头，则不会拼接设置的`Endpoint`, 如下方应对钉钉新老接口的问题
```go
clientOps := []ghttp.ClientOption{
    ghttp.WithEndpoint("https://api.dingtalk.com"),
}
// 钉钉老版接口
// https://oapi.dingtalk.com/department/list
client.Invoke(context.Background(), http.MethodGet, "https://oapi.dingtalk.com/department/list", args, reply)

// 钉钉新版接口
// https://api.dingtalk.com/department/list
client.Invoke(context.Background(), http.MethodPost, "/v1.0/oauth2/accessToken", args, reply)
```

## SDK
- [go-jira](https://github.com/zdz1715/go-jira)
- [go-gitlab](https://github.com/zdz1715/go-gitlab)
- [go-gitee](https://github.com/zdz1715/go-gitee)
- [go-dingtalk](https://github.com/zdz1715/go-dingtalk)