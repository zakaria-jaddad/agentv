package config

import (
	"fmt"
)

func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required, how do you want me to connect dumb ass")
	}

	if c.Backend.AuthEndpoint == "" {
		return fmt.Errorf("backend.url is required")
	}

	if c.Vector.Binary == "" {
		return fmt.Errorf("vector.binary is required")
	}

	if c.Vector.Config == "" {
		return fmt.Errorf("vector.config is required")
	}

	if c.Vector.Socket == "" {
		return fmt.Errorf("vector.socket is required")
	}

	return nil
}
