// +build ignore

// This example demonstrates Viper integration.
// To run this example, install Viper first:
//   go get github.com/spf13/viper
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
	"github.com/spf13/viper"
)

// ViperConfig demonstrates integration with Viper.
// SSM Parameters:
//   /myapp/database.url = "postgres://localhost:5432/mydb"
//   /myapp/database.port = "5432"
//   /myapp/server.host = "0.0.0.0"
//   /myapp/server.port = "8080"
func main() {
	ctx := context.Background()

	// Method 1: Read all SSM parameters and set in Viper
	fmt.Println("=== Method 1: ReadRemoteConfig ===")
	values, err := ssmconfig.ReadRemoteConfig(ctx, "/myapp/")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Set values in Viper
	for key, value := range values {
		viper.Set(key, value)
	}

	// Use Viper normally
	dbURL := viper.GetString("database.url")
	dbPort := viper.GetInt("database.port")
	fmt.Printf("Database: %s:%d\n", dbURL, dbPort)

	// Method 2: Using ViperRemoteProvider
	fmt.Println("\n=== Method 2: ViperRemoteProvider ===")
	provider, err := ssmconfig.SetViperRemoteProvider(ctx, "/myapp/")
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Stop()

	// Get all values and merge into Viper
	viperValues := provider.GetViperValues()
	viper.MergeConfigMap(viperValues)

	// Or get individual values
	serverHost, _ := provider.Get("server.host")
	serverPort, _ := provider.Get("server.port")
	fmt.Printf("Server: %s:%s\n", serverHost, serverPort)

	// Use Viper
	fmt.Printf("Server Host (from Viper): %s\n", viper.GetString("server.host"))
}

