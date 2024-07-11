package ghttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestWithDebug(t *testing.T) {
	client := NewClient(
		WithEndpoint("https://gitlab.com"),
		WithDebug(func() DebugInterface {
			return &Debug{
				Writer: os.Stdout,
				Trace:  true,
				TraceCallback: func(w io.Writer, info TraceInfo) {
					_, _ = w.Write(info.Table())
				},
			}
		}),
	)

	data := map[string]interface{}{
		"grant_type": "password",
	}

	var reply any
	_, err := client.Invoke(context.Background(), http.MethodPost, "/oauth/token", data, &reply, &CallOptions{
		Query: map[string]any{
			"page": "1",
			//"membership": true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Invoke /oauth/token success, reply: %+v\n", reply)
}
