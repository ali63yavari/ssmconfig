//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
)

// EdgeCaseConfig demonstrates edge cases and advanced usage.
type EdgeCaseConfig struct {
	// Pointer to struct
	Database *struct {
		Host string `ssm:"host"`
		Port int    `ssm:"port"`
	} `ssm:"database"`

	// String slice (comma-separated)
	AllowedIPs []string `ssm:"allowed_ips"`

	// String slice from JSON
	Tags []string `ssm:"tags" json:"true"`

	// Optional field (no required tag)
	OptionalField string `ssm:"optional_field"`

	// Field with both SSM and env (env takes precedence)
	APIKey string `ssm:"api_key" env:"API_KEY" required:"true"`

	// Nested struct with JSON decoding
	Metadata struct {
		Env     string `json:"env"`
		Region  string `json:"region"`
		Version string `json:"version"`
	} `ssm:"metadata" json:"true"`

	// Field with custom validator
	Email string `ssm:"email" validate:"email"`
}

func main() {
	ctx := context.Background()

	// Example 1: Reusing loader instance
	fmt.Println("=== Reusing Loader ===")
	loader, err := ssmconfig.NewLoader(ctx,
		ssmconfig.WithStrictMode(false),
		ssmconfig.WithLogger(func(format string, args ...interface{}) {
			fmt.Printf("[LOG] "+format+"\n", args...)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create loader: %v", err)
	}

	cfg1, err := ssmconfig.LoadWithLoader[EdgeCaseConfig](loader, ctx, "/myapp/")
	if err != nil {
		log.Printf("Error loading config1: %v", err)
	} else {
		fmt.Printf("Config 1 loaded successfully\n")
	}

	// Load another config with same loader
	cfg2, err := ssmconfig.LoadWithLoader[EdgeCaseConfig](loader, ctx, "/otherapp/")
	if err != nil {
		log.Printf("Error loading config2: %v", err)
	} else {
		fmt.Printf("Config 2 loaded successfully\n")
	}

	// Example 2: Cache invalidation
	fmt.Println("\n=== Cache Invalidation ===")
	loader.InvalidateCache("/myapp/")
	fmt.Println("Cache invalidated for /myapp/")

	// Reload after invalidation
	cfg3, err := ssmconfig.LoadWithLoader[EdgeCaseConfig](loader, ctx, "/myapp/")
	if err != nil {
		log.Printf("Error reloading: %v", err)
	} else {
		fmt.Printf("Config reloaded after cache invalidation\n")
		_ = cfg3
	}

	// Example 3: Strong typing vs JSON
	fmt.Println("\n=== Strong Typing vs JSON ===")
	type TypingConfig struct {
		// Without json tag - uses strongly-typed conversion
		Port1 int `ssm:"port1"`

		// With json tag - uses JSON decoding
		Port2 int `ssm:"port2" json:"true"`
	}

	// With strong typing (default)
	cfg4, _ := ssmconfig.Load[TypingConfig](ctx, "/myapp/")
	fmt.Printf("Port1 (strong typing): %d\n", cfg4.Port1)

	// With JSON preference
	cfg5, _ := ssmconfig.Load[TypingConfig](ctx, "/myapp/",
		ssmconfig.WithStrongTyping(false))
	fmt.Printf("Port2 (JSON): %d\n", cfg5.Port2)
}
