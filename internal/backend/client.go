package backend

import (
	"net/http"

	"github.com/zishang520/socket.io/clients/socket/v3"
)

type Client struct {
	baseURL           string
	installationToken string
	httpClient        *http.Client

	socket *socket.Socket
}

func NewClient(baseURL string, installationToken string) *Client {
	return &Client{
		baseURL:           baseURL,
		installationToken: installationToken,
		httpClient:        &http.Client{},
	}
}
