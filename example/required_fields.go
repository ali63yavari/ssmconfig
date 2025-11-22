//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
)

// RequiredConfig demonstrates required field validation.
// SSM Parameters (some may be missing):
//
//	/myapp/api_key = "secret-key-123"
//	/myapp/optional_field = "optional-value"
type RequiredConfig struct {
	APIKey        string `ssm:"api_key" env:"API_KEY" required:"true"`
	OptionalField string `ssm:"optional_field" env:"OPTIONAL_FIELD"`
	RequiredInt   int    `ssm:"required_int" env:"REQUIRED_INT" required:"true"`
}

func main() {
	ctx := context.Background()

	// Example 1: Non-strict mode (logs warnings, doesn't panic)
	fmt.Println("=== Non-strict mode ===")
	cfg1, err := ssmconfig.Load[RequiredConfig](ctx, "/myapp/")
	if err != nil {
		log.Printf("Error (non-strict): %v", err)
	} else {
		fmt.Printf("Config loaded: %+v\n", cfg1)
	}

	// Example 2: Strict mode (panics on missing required fields)
	fmt.Println("\n=== Strict mode ===")
	cfg2, err := ssmconfig.Load[RequiredConfig](ctx, "/myapp/",
		ssmconfig.WithStrictMode(true))
	if err != nil {
		log.Printf("Error (strict): %v", err)
	} else {
		fmt.Printf("Config loaded: %+v\n", cfg2)
	}

	// Example 3: With custom logger
	fmt.Println("\n=== With custom logger ===")
	cfg3, err := ssmconfig.Load[RequiredConfig](ctx, "/myapp/",
		ssmconfig.WithLogger(func(format string, args ...interface{}) {
			fmt.Printf("[CUSTOM LOGGER] "+format+"\n", args...)
		}))
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Config loaded: %+v\n", cfg3)
	}
}
