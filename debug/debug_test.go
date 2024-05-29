package debug

import (
	"net/http"
	"os"
	"testing"

	"github.com/zdz1715/ghttp"
)

func TestDebug(t *testing.T) {
	client := ghttp.NewClient(
		ghttp.WithEndpoint("https://gitlab.com"),
	)

	debug := &Debug{
		Writer: os.Stdout,
		Trace:  true,
	}
	req, err := http.NewRequest(http.MethodGet, "/api/v4/projects", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(req, &ghttp.CallOptions{
		Query: map[string]any{
			"page":       "1",
			"membership": true,
		},
	}, debug)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", debug.TraceInfo())

}
