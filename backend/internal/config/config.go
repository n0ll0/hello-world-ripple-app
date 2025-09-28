package config

import (
	"log"
	"os"
	"strings"

	"github.com/Netflix/go-env"
)

type AppConfig struct {
	Port    string         `env:"PORT,default=8080"`
	DBPath  string         `env:"DB_PATH,default=app.db"`
	Clients []OAuth2Client // loaded from OAUTH2_CLIENTS
}

type OAuth2Client struct {
	ID     string
	Secret string
	Domain string
}

func Load() *AppConfig {
	var cfg AppConfig
	_, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}
	clientsEnv := os.Getenv("OAUTH2_CLIENTS")
	if clientsEnv != "" {
		for _, entry := range strings.Split(clientsEnv, ",") {
			parts := strings.SplitN(entry, ":", 3)
			if len(parts) == 3 {
				cfg.Clients = append(cfg.Clients, OAuth2Client{
					ID:     parts[0],
					Secret: parts[1],
					Domain: parts[2],
				})
			}
		}
	}
	if len(cfg.Clients) == 0 {
		cfg.Clients = append(cfg.Clients, OAuth2Client{
			ID:     "hello-client",
			Secret: "super-secret",
			Domain: "http://localhost",
		})
	}
	return &cfg
}
