// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ali63yavari/ssmconfig"
)

// FileConfig demonstrates loading configuration from YAML, JSON, and TOML files.
// Priority: ENV > File > SSM
//
// Example config.yaml:
//   database:
//     url: "postgres://localhost:5432/mydb"
//     port: 5432
//   server:
//     host: "0.0.0.0"
//     port: 8080
//
// Example config.json:
//   {
//     "database": {
//       "url": "postgres://localhost:5432/mydb",
//       "port": 5432
//     }
//   }
//
// Example config.toml:
//   [database]
//   url = "postgres://localhost:5432/mydb"
//   port = 5432
type FileConfig struct {
	Database struct {
		URL  string `ssm:"database/url" env:"DB_URL"`
		Port int    `ssm:"database/port" env:"DB_PORT"`
	} `ssm:"database"`

	Server struct {
		Host string `ssm:"server/host" env:"SERVER_HOST"`
		Port int    `ssm:"server/port" env:"SERVER_PORT"`
	} `ssm:"server"`
}

func main() {
	ctx := context.Background()

	// Example 1: Load from YAML file
	fmt.Println("=== Loading from YAML ===")
	cfg1, err := ssmconfig.Load[FileConfig](ctx, "/myapp/",
		ssmconfig.WithConfigFiles("config.yaml"))
	if err != nil {
		log.Printf("Error loading from YAML: %v", err)
	} else {
		fmt.Printf("Database: %s:%d\n", cfg1.Database.URL, cfg1.Database.Port)
		fmt.Printf("Server: %s:%d\n", cfg1.Server.Host, cfg1.Server.Port)
	}

	// Example 2: Load from JSON file
	fmt.Println("\n=== Loading from JSON ===")
	cfg2, err := ssmconfig.Load[FileConfig](ctx, "/myapp/",
		ssmconfig.WithConfigFiles("config.json"))
	if err != nil {
		log.Printf("Error loading from JSON: %v", err)
	} else {
		fmt.Printf("Database: %s:%d\n", cfg2.Database.URL, cfg2.Database.Port)
	}

	// Example 3: Load from TOML file
	fmt.Println("\n=== Loading from TOML ===")
	cfg3, err := ssmconfig.Load[FileConfig](ctx, "/myapp/",
		ssmconfig.WithConfigFiles("config.toml"))
	if err != nil {
		log.Printf("Error loading from TOML: %v", err)
	} else {
		fmt.Printf("Database: %s:%d\n", cfg3.Database.URL, cfg3.Database.Port)
	}

	// Example 4: Multiple files (later files override earlier ones)
	fmt.Println("\n=== Loading from multiple files ===")
	cfg4, err := ssmconfig.Load[FileConfig](ctx, "/myapp/",
		ssmconfig.WithConfigFiles("config.yaml", "config.local.yaml"))
	if err != nil {
		log.Printf("Error loading from multiple files: %v", err)
	} else {
		fmt.Printf("Database: %s:%d\n", cfg4.Database.URL, cfg4.Database.Port)
	}

	// Example 5: Priority demonstration (ENV > File > SSM)
	fmt.Println("\n=== Priority: ENV > File > SSM ===")
	os.Setenv("DB_URL", "env-override")
	defer os.Unsetenv("DB_URL")

	cfg5, err := ssmconfig.Load[FileConfig](ctx, "/myapp/",
		ssmconfig.WithConfigFiles("config.yaml"))
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		// Should show "env-override" (ENV takes precedence)
		fmt.Printf("Database URL: %s (from ENV)\n", cfg5.Database.URL)
	}
}

