package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrConfigNotFound = errors.New("config file not found")

var defaultConfigNames = []string{
	"kyn.config.yaml",
	"kyn.config.yml",
	".kyn.yaml",
	".kyn.yml",
}

func Load(cwd string, explicitPath string) (Config, string, error) {
	cfgPath, err := resolveConfigPath(cwd, explicitPath)
	if err != nil {
		return Config{}, "", err
	}

	cfg, err := decodeConfigFile(cfgPath)
	if err != nil {
		return Config{}, "", err
	}

	if err := Validate(cfg); err != nil {
		return Config{}, "", err
	}

	return cfg, cfgPath, nil
}

func resolveConfigPath(cwd string, explicitPath string) (string, error) {
	if explicitPath != "" {
		p := explicitPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, explicitPath)
		}
		if _, err := os.Stat(p); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("%w: %s", ErrConfigNotFound, explicitPath)
			}
			return "", fmt.Errorf("stat config: %w", err)
		}
		return filepath.Clean(p), nil
	}

	for _, name := range defaultConfigNames {
		p := filepath.Join(cwd, name)
		if _, err := os.Stat(p); err == nil {
			return filepath.Clean(p), nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat config: %w", err)
		}
	}

	return "", fmt.Errorf("%w (searched: %s)", ErrConfigNotFound, strings.Join(defaultConfigNames, ", "))
}

func decodeConfigFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse yaml config: %w", err)
	}

	return cfg, nil
}
