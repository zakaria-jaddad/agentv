package backend

import "net/http"

type Client struct {
	baseURL           string
	installationToken string
	httpClient        *http.Client
}

func NewClient(baseURL string, installationToken string) *Client {
	return &Client{
		baseURL:           baseURL,
		installationToken: installationToken,
		httpClient:        &http.Client{},
	}
}
