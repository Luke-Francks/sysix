package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type PanelConfig struct {
	System    bool `yaml:"system"`
	Processes bool `yaml:"processes"`
	Ports     bool `yaml:"ports"`
	Network   bool `yaml:"network"`
}

type WebConfig struct {
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
	Enabled bool   `yaml:"enabled"`
}

type AgentConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

type ObserverAgent struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

type ObserverConfig struct {
	Enabled      bool            `yaml:"enabled"`
	PollInterval int             `yaml:"poll_interval"`
	Agents       []ObserverAgent `yaml:"agents"`
}

type HistoryConfig struct {
	Enabled       bool   `yaml:"enabled"`
	DBPath        string `yaml:"db_path"`
	RetentionDays int    `yaml:"retention_days"`
	APIPoints     int    `yaml:"api_points"`
}

type Config struct {
	RefreshRate int            `yaml:"refresh_rate"`
	Panels      PanelConfig    `yaml:"panels"`
	Web         WebConfig      `yaml:"web"`
	Agent       AgentConfig    `yaml:"agent"`
	Observer    ObserverConfig `yaml:"observer"`
	History     HistoryConfig  `yaml:"history"`
}

func Default() *Config {
	return &Config{
		RefreshRate: 2,
		Panels: PanelConfig{
			System:    true,
			Processes: true,
			Ports:     true,
			Network:   true,
		},
		Web: WebConfig{
			Port:    8080,
			Host:    "localhost",
			Enabled: true,
		},
		Agent: AgentConfig{
			Enabled: true,
			Host:    "0.0.0.0",
			Port:    9090,
		},
		Observer: ObserverConfig{
			Enabled:      false,
			PollInterval: 2,
			Agents:       []ObserverAgent{},
		},
		History: HistoryConfig{
			Enabled:       true,
			DBPath:        "sysix-history.db",
			RetentionDays: 7,
			APIPoints:     60,
		},
	}
}

func Load() (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile("config.yaml", data, 0o644)
}
