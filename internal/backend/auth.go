package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	StatusCode int    `json:"statusCode"`
	Timestamp  string `json:"timestamp"`
	Path       string `json:"path"`
	Message    string `json:"message"`
	Data       data   `json:data`
}

type NormalizedResponse struct {
	Data json.RawMessage `json:"data"`
}

func (c *Client) Authenticate(manager *manager.Manager, ctx context.Context) (*AuthResponse, error) {

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
		return nil, &AuthError{StatusCode: authResponse.StatusCode, Message: authResponse.Message}
	}

	return &authResponse, nil
}

const (
	initialRetryDelay = 2 * time.Second
	maxRetryDelay     = 60 * time.Second
)

func (c *Client) AuthenticateWithRetry(manager *manager.Manager, ctx context.Context) (*AuthResponse, error) {

	delay := initialRetryDelay

	for {
		auth, err := c.Authenticate(manager, ctx)
		// Authentication succeeded
		if err == nil {
			return auth, nil
		}

		// check if permanent authentication failure
		var authErr *AuthError

		if errors.As(err, &authErr) && authErr.Permanent() {
			return nil, err
		}

		// Tempporary backend/network failure
		log.Printf("backend authentication failed: %v; retrying in %s", err, delay)

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()

		case <-timer.C:
		}

		delay *= 2

		if delay > maxRetryDelay {
			delay = maxRetryDelay
		}

	}

}
