package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func main() {
	args := map[string]any{
		"grant_type": "password",
		"client_id":  "app",
	}

	gitlab := NewGitlab()

	var reply any
	// 	请求 https://gitlab.com/oauth/token
	err := gitlab.Invoke(context.Background(), http.MethodPost, "/oauth/token", args, &reply)
	if err != nil {
		fmt.Printf("Invoke /oauth/token, error: %s\n", err)
	} else {
		fmt.Printf("Invoke /oauth/token success, reply: %+v\n", reply)
	}

	args = map[string]any{
		"page":       "1",
		"membership": true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 	请求 https://gitlab.com/api/v4/projects
	err = gitlab.Invoke(ctx, http.MethodGet, "/api/v4/projects", args, &reply)
	if err != nil {
		fmt.Printf("Invoke /api/v4/projects, error: %s\n", err)
	} else {
		fmt.Printf("Invoke /api/v4/projects success, reply: %+v\n", reply)
	}

}

type Gitlab struct {
	cc *ghttp.Client
}

func NewGitlab() *Gitlab {
	clientOps := []ghttp.ClientOption{
		ghttp.WithDebug(ghttp.DefaultDebug),
		ghttp.WithEndpoint("https://gitlab.com"),
		ghttp.WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		}),
		ghttp.WithUserAgent("sdk/gitlab-v0.0.1"),
		ghttp.WithNot2xxError(func() ghttp.Not2xxError {
			return new(gitlabError)
		}),
	}
	client := ghttp.NewClient(clientOps...)
	return &Gitlab{
		cc: client,
	}

}

func (g *Gitlab) Invoke(ctx context.Context, method, path string, args, reply any) error {
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
	var err error
	if method == http.MethodGet && args != nil {
		callOptions.Query = args
		_, err = g.cc.Invoke(ctx, method, path, nil, reply, callOptions)
	} else {
		_, err = g.cc.Invoke(ctx, method, path, args, reply, callOptions)
	}

	return err
}
