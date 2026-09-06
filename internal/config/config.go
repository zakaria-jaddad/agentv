package config

type Config struct {
	Version int `yaml:"version"`
	Token   string
	Backend BackendConfig `yaml:"backend"`
	Agent   AgentConfig   `yaml:"agent"`
	Vector  VectorConfig  `yaml:"vector"`
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
	Socket string `yaml:"socket"`
}
