package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	LiveKit  LiveKitConfig  `yaml:"livekit"`
	Upload   UploadConfig   `yaml:"upload"`
	External ExternalConfig `yaml:"external"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type JWTConfig struct {
	Secret          string `yaml:"secret"`
	AccessExpiry    int    `yaml:"access_expiry"`
	RefreshExpiry   int    `yaml:"refresh_expiry"`
}

type LiveKitConfig struct {
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
	URL       string `yaml:"url"`
}

type UploadConfig struct {
	MaxSize    int64  `yaml:"max_size"`
	UploadDir  string `yaml:"upload_dir"`
}

type ExternalConfig struct {
	APIKey string `yaml:"api_key"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "iroom.db",
		},
		JWT: JWTConfig{
			Secret:        "change-me-in-production",
			AccessExpiry:  15,
			RefreshExpiry: 10080,
		},
		LiveKit: LiveKitConfig{
			APIKey:    "",
			APISecret: "",
			URL:       "ws://localhost:7880",
		},
		Upload: UploadConfig{
			MaxSize:   50 << 20,
			UploadDir: "uploads",
		},
		External: ExternalConfig{
			APIKey: "",
		},
	}
}
