package config

import (
	"fmt"
)

func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required, how do you want me to connect dumb ass")
	}

	if c.Backend.URL == "" {
		return fmt.Errorf("backend.url is required")
	}

	if c.Vector.Binary == "" {
		return fmt.Errorf("vector.binary is required")
	}

	return nil
}
