package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zakaria-jaddad/agentv/internal/manager"
)

// export interface NormalizedResponse<T> {
//   success: boolean;
//   statusCode: number;
//   message: string;
//   data: T;
//   timestamp: string;
//   path: string;
// }

type data struct {
	AgentID   int64  `json:"agentID"`
	SocketURL string `json:"socketUrl"`
}

type AuthResponse struct {
	Success    bool   `json:"success"`
	StatusCode int32  `json:"statusCode"`
	Timestamp  string `json:"timestamp"`
	Path       string `json:"path"`
	Message    string `json:"message"`
	Data       data   `json:data`
}

type NormalizedResponse struct {
	Data json.RawMessage `json:"data"`
}

func (c *Client) Authenticate(manager *manager.Manager) (*AuthResponse, error) {

	payload := map[string]string{
		"hostname":     manager.Agent.Hostname,
		"platform":     manager.Agent.OS,
		"architecture": manager.Agent.Architecture,
		"name":         manager.Agent.Name,
		"token":        c.installationToken,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/agent/auth", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var authResponse AuthResponse

	if err := json.NewDecoder(res.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("decode authentication response  %w", err)
	}

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("authentication failed: server returned  %d, message: %s", res.StatusCode, authResponse.Message)
	}

	return &authResponse, nil
}
