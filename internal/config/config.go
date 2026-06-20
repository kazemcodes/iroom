package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Upload   UploadConfig   `yaml:"upload"`
	External ExternalConfig `yaml:"external"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	WebRTC   WebRTCConfig   `yaml:"webrtc"`
}

type WebRTCConfig struct {
	PublicIP string `yaml:"public_ip"`
	STUNPort int    `yaml:"stun_port"`
	TurnPort int    `yaml:"turn_port"`
	UDPPort  int    `yaml:"udp_port"`
	TurnSecret string `yaml:"turn_secret"`
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

type UploadConfig struct {
	MaxSize    int64  `yaml:"max_size"`
	UploadDir  string `yaml:"upload_dir"`
}

type ExternalConfig struct {
	APIKey string `yaml:"api_key"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	Enabled  bool   `yaml:"enabled"`
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	if cfg.JWT.Secret == "change-me-in-production" {
		cfg.JWT.Secret = generateRandomSecret(32)
	}

	return cfg, nil
}

func generateRandomSecret(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "change-me-in-production"
	}
	return fmt.Sprintf("%x", b)
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = p
		}
	}
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("JWT_ACCESS_EXPIRY"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.JWT.AccessExpiry = p
		}
	}
	if v := os.Getenv("JWT_REFRESH_EXPIRY"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.JWT.RefreshExpiry = p
		}
	}
	if v := os.Getenv("UPLOAD_MAX_SIZE"); v != "" {
		if p, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Upload.MaxSize = p
		}
	}
	if v := os.Getenv("UPLOAD_DIR"); v != "" {
		cfg.Upload.UploadDir = v
	}
	if v := os.Getenv("EXTERNAL_API_KEY"); v != "" {
		cfg.External.APIKey = v
	}
	if v := os.Getenv("SMTP_HOST"); v != "" {
		cfg.SMTP.Host = v
	}
	if v := os.Getenv("SMTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.SMTP.Port = p
		}
	}
	if v := os.Getenv("SMTP_USERNAME"); v != "" {
		cfg.SMTP.Username = v
	}
	if v := os.Getenv("SMTP_PASSWORD"); v != "" {
		cfg.SMTP.Password = v
	}
	if v := os.Getenv("SMTP_FROM"); v != "" {
		cfg.SMTP.From = v
	}
	if v := os.Getenv("SMTP_ENABLED"); v != "" {
		cfg.SMTP.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("WEBRTC_PUBLIC_IP"); v != "" {
		cfg.WebRTC.PublicIP = v
	}
	if v := os.Getenv("WEBRTC_STUN_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.WebRTC.STUNPort = p
		}
	}
	if v := os.Getenv("WEBRTC_TURN_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.WebRTC.TurnPort = p
		}
	}
	if v := os.Getenv("WEBRTC_TURN_SECRET"); v != "" {
		cfg.WebRTC.TurnSecret = v
	}
	if v := os.Getenv("WEBRTC_UDP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.WebRTC.UDPPort = p
		}
	}
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
		Upload: UploadConfig{
			MaxSize:   50 << 20,
			UploadDir: "uploads",
		},
		External: ExternalConfig{
			APIKey: "",
		},
		SMTP: SMTPConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "",
			Password: "",
			From:     "noreply@iroom.ir",
			Enabled:  false,
		},
		WebRTC: WebRTCConfig{
			PublicIP:   "",
			STUNPort:   3478,
			TurnPort:   3479,
			UDPPort:    8081,
			TurnSecret: "",
		},
	}
}
