package config

import (
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type APIConfig struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
}

type Config struct {
	API       APIConfig `yaml:"api"`
	SphCookie string    `yaml:"sph_cookie"`
	mu        sync.RWMutex
}

func Load() *Config {
	cfg := &Config{
		API: APIConfig{
			Hostname: "0.0.0.0",
			Port:     2022,
		},
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("[config] config.yaml not found, using defaults: %v", err)
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Printf("[config] parse config.yaml failed: %v, using defaults", err)
		return cfg
	}

	log.Printf("[config] loaded: %s:%d", cfg.API.Hostname, cfg.API.Port)
	return cfg
}

func (c *Config) GetSphCookie() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SphCookie
}

func (c *Config) SetSphCookie(cookie string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SphCookie = cookie
}
