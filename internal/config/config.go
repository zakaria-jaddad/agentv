package config

const (
	// AgentvDir        = "/var/lib/agentv"
	// AgentvSocketPath = "/var/lib/agentv/agentv.sock"
	AgentvDir        = "/tmp/agentv"
	AgentvSocketPath = "/tmp/agentv/agentv.sock"
)

type Config struct {
	Version int `yaml:"version"`
	Token   string
	Backend BackendConfig `yaml:"backend"`
	Agent   AgentConfig   `yaml:"agent"`
	Vector  VectorConfig  `yaml:"vector"`
	Socket  string
}

type BackendConfig struct {
	URL string `yaml:"url"`
}

type AgentConfig struct {
	Name string `yaml:"name"`
}

type VectorConfig struct {
	Binary string `yaml:"binary"`
	Config string `yaml:"config"`
}
