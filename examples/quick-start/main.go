package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zdz1715/ghttp"
)

func main() {

	client := ghttp.NewClient(
		ghttp.WithEndpoint("https://gitlab.com"),
	)

	var reply any
	_, err := client.Invoke(context.Background(), http.MethodGet, "/api/v4/projects", nil, &reply, &ghttp.CallOptions{
		Query: map[string]any{
			"page": "1",
			//"membership": true,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Invoke /api/v4/projects success, reply: %+v\n", reply)
}
