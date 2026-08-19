package backend

import (
	"fmt"
	"net/http"
)

type AuthError struct {
	Message    string
	StatusCode int
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authtentication failed: %s (status: %d)", e.Message, e.StatusCode)
}

func (e *AuthError) Permanent() bool {
	switch e.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return true
	default:
		return false
	}
}
