//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
)

// DatabaseConfig is a nested configuration struct.
type DatabaseConfig struct {
	Host     string `ssm:"host" env:"DB_HOST" required:"true"`
	Port     int    `ssm:"port" env:"DB_PORT"`
	Username string `ssm:"username" env:"DB_USER"`
	Password string `ssm:"password" env:"DB_PASS"`
}

// ServerConfig is another nested configuration struct.
type ServerConfig struct {
	Host string `ssm:"host" env:"SERVER_HOST"`
	Port int    `ssm:"port" env:"SERVER_PORT" required:"true"`
}

// NestedConfig demonstrates nested struct support.
// SSM Parameters:
//
//	/myapp/database/host = "localhost"
//	/myapp/database/port = "5432"
//	/myapp/database/username = "admin"
//	/myapp/database/password = "secret"
//	/myapp/server/host = "0.0.0.0"
//	/myapp/server/port = "8080"
type NestedConfig struct {
	Database DatabaseConfig `ssm:"database" required:"true"`
	Server   ServerConfig   `ssm:"server" required:"true"`
}

func main() {
	ctx := context.Background()

	cfg, err := ssmconfig.Load[NestedConfig](ctx, "/myapp/")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Database: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
