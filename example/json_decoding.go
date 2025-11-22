// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
)

// JSONConfig demonstrates JSON decoding from SSM Parameter Store.
// SSM Parameters:
//   /myapp/database = `{"host":"localhost","port":5432,"ssl":true}`
//   /myapp/servers = `[{"name":"api","port":8080},{"name":"web","port":80}]`
//   /myapp/metadata = `{"env":"prod","region":"us-east-1"}`
type JSONConfig struct {
	// Nested struct from JSON
	Database struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		SSL  bool   `json:"ssl"`
	} `ssm:"database" json:"true" required:"true"`

	// Slice of structs from JSON
	Servers []struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	} `ssm:"servers" json:"true"`

	// Map from JSON
	Metadata map[string]string `ssm:"metadata" json:"true"`
}

func main() {
	ctx := context.Background()

	cfg, err := ssmconfig.Load[JSONConfig](ctx, "/myapp/")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Database: %s:%d (SSL: %t)\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.SSL)
	fmt.Printf("Servers: %d servers\n", len(cfg.Servers))
	for _, s := range cfg.Servers {
		fmt.Printf("  - %s:%d\n", s.Name, s.Port)
	}
	fmt.Printf("Metadata: %v\n", cfg.Metadata)
}

