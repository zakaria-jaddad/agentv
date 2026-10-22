package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func RuntimeDir() (string, error) {

	// RUNTIME_DIRECTORY is provided from systemd
	if dir := os.Getenv("RUNTIME_DIRECTORY"); dir != "" {
		return dir, nil
	}

	// XDG_RUNTIME_DIR is provided by the XDG Base Directory Specification
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "agentv"), nil
	}

	// for macos and windows, we can use the temp directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "agentv", "run"), nil
}

func StateDir() (string, error) {
	// Automatically provided by systemd
	if dir := os.Getenv("STATE_DIRECTORY"); dir != "" {
		return dir, nil
	}

	// macOS: ~/Library/Application Support/agentv
	if runtime.GOOS == "darwin" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "agentv"), nil
	}

	// Linux: ~/.local/state/agentv
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, "agentv"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "state", "agentv"), nil
}
