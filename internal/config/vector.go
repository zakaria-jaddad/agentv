package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func (c *Config) PrepareVectorConfig(
	sourcePath string,
	runtimeDir string,
	stateDir string,
	socketPath string,
) (string, error) {

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("read vector config: %v", err)
	}

	var conf map[string]any

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return "", fmt.Errorf("parse vector config: %v", err)
	}

	sinks, ok := conf["sinks"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("vector config missing sinks section")
	}

	sink, ok := sinks["agentv"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("vector config missing agentv sink")
	}

	// sink type verification, it must be of type socket
	if sink["type"] != "socket" || sink["mode"] != "unix" {
		return "", fmt.Errorf("agentv sink must be of type socket and mode unix")
	}

	// Update the socket path in the sink configuration
	sink["path"] = socketPath

	// add vector state dir
	vectorStateDir := filepath.Join(stateDir, "vector")
	if err := os.MkdirAll(vectorStateDir, 0700); err != nil {
		log.Fatalf("create state directory: %v", err)
	}

	// update vector data_dir
	conf["data_dir"] = vectorStateDir
	log.Printf("Vector state dir: %s", vectorStateDir)

	outputData, err := yaml.Marshal(conf)
	if err != nil {
		return "", fmt.Errorf("marshal vector config: %v", err)
	}

	tmp, err := os.CreateTemp(runtimeDir, "vector-*.yml")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(outputData); err != nil {
		tmp.Close()
		return "", err
	}

	// close the temp file
	if err := tmp.Close(); err != nil {
		return "", err
	}

	// create new file on the runtime dir and write the conf to it
	generatedPath := filepath.Join(runtimeDir, "vector.generated.yaml")

	// rename the temp file to the generated path
	if err := os.Rename(tmp.Name(), generatedPath); err != nil {
		return "", err
	}

	return generatedPath, nil
}
