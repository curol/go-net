package client

import (
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	// TestGet(t)
	clientReq := &ClientConfig{
		Method:  "GET",
		Address: "http://localhost:8080/home",
		Header: map[string]string{
			"Host":           "localhost:8080",
			"Content-Type":   "text/plain",
			"Content-Length": "13",
		},
		Body: []byte("Hello, World!"),
	}
	client := NewClient(clientReq)
	t.Log("client", client)
	fmt.Println("Client req: ", client.req.ToBytes())
	fmt.Println("Client res: ", client.res.ToBytes())
}
