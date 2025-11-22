// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
)

// BasicConfig demonstrates basic usage of ssmconfig.
// SSM Parameters:
//   /myapp/database_url = "postgres://localhost:5432/mydb"
//   /myapp/port = "8080"
//   /myapp/debug = "true"
type BasicConfig struct {
	DatabaseURL string `ssm:"database_url" env:"DB_URL" required:"true"`
	Port        int    `ssm:"port" env:"PORT"`
	Debug       bool   `ssm:"debug" env:"DEBUG"`
}

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := ssmconfig.Load[BasicConfig](ctx, "/myapp/")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Debug: %t\n", cfg.Debug)
}

