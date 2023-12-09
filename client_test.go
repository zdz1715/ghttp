package ghttp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"
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

func TestClient_Invoke(t *testing.T) {
	not2xxBody := &gitlabError{}
	clientOps := []ClientOption{
		WithDebug(true),
		WithEndpoint("https://gitlab.com"),
		WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		}),
		WithUserAgent("sdk/gitlab-v0.0.1"),
		WithNot2xxError(not2xxBody),
	}

	client, err := NewClient(context.Background(), clientOps...)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]string{
		"grant_type": "password",
		"client_id":  "app",
	}
	var reply any
	res, err := client.Invoke(context.Background(), http.MethodPost, "/oauth/token", body, &reply, &CallOptions{
		Query: body,

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
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("response: %+v\n reply: %+v\nnot2xxBody: %+v\n", res.Body, reply, not2xxBody)
}
