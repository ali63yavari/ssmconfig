//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ali63yavari/ssmconfig"
)

// RefreshConfig demonstrates auto-refreshing configuration.
type RefreshConfig struct {
	DatabaseURL string `ssm:"database_url" env:"DB_URL" required:"true"`
	Port        int    `ssm:"port" env:"PORT"`
	LastUpdated string `ssm:"last_updated" env:"LAST_UPDATED"`
}

func main() {
	ctx := context.Background()

	// Create loader first
	loader, err := ssmconfig.NewLoader(ctx)
	if err != nil {
		log.Fatalf("Failed to create loader: %v", err)
	}

	// Load with auto-refresh every 30 seconds
	refreshingConfig, err := ssmconfig.LoadWithAutoRefreshAndLoader[RefreshConfig](
		loader,
		ctx,
		"/myapp/",
		ssmconfig.WithRefreshInterval[RefreshConfig](30*time.Second),
		ssmconfig.WithOnChange[RefreshConfig](func(old, new *RefreshConfig) {
			fmt.Printf("Config changed!\n")
			fmt.Printf("  Old Database URL: %s\n", old.DatabaseURL)
			fmt.Printf("  New Database URL: %s\n", new.DatabaseURL)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	defer refreshingConfig.Stop()

	// Get current config (thread-safe)
	for i := 0; i < 5; i++ {
		cfg := refreshingConfig.Get()
		fmt.Printf("[%d] Database URL: %s, Port: %d\n", i+1, cfg.DatabaseURL, cfg.Port)
		time.Sleep(10 * time.Second)
	}

	// Manually refresh
	fmt.Println("\nManually refreshing...")
	if err := refreshingConfig.Refresh(); err != nil {
		log.Printf("Refresh error: %v", err)
	}

	// Get a safe copy to modify
	cfgCopy, err := refreshingConfig.GetCopy()
	if err != nil {
		log.Printf("GetCopy error: %v", err)
	} else {
		fmt.Printf("Config copy: %+v\n", cfgCopy)
	}
}
