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
