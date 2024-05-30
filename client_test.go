package ghttp

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/zdz1715/ghttp/debug"
)

func TestWithDebug(t *testing.T) {
	client := NewClient(
		WithEndpoint("https://gitlab.com"),
		WithDebug(func() debug.Interface {
			return &debug.Debug{
				Writer: os.Stdout,
				Trace:  true,
				TraceCallback: func(info *debug.TraceInfo) {
					fmt.Printf("trace info: %+v\n", info)
				},
			}
		}),
	)

	req, err := http.NewRequest(http.MethodGet, "/api/v4/projects", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(req, &CallOptions{
		Query: map[string]any{
			"page":       "1",
			"membership": true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
