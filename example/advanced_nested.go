// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ali63yavari/ssmconfig"
)

// AdvancedNestedConfig demonstrates complex nested structures.
type DatabaseConfig struct {
	Host     string   `ssm:"host" env:"DB_HOST" required:"true" validate:"minlen:3"`
	Port     int      `ssm:"port" env:"DB_PORT" validate:"min:1,max:65535"`
	SSL      bool     `ssm:"ssl" env:"DB_SSL"`
	Replicas []string `ssm:"replicas" json:"true"` // JSON array
}

type CacheConfig struct {
	Redis struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		DB   int    `json:"db"`
	} `ssm:"redis" json:"true" required:"true"` // Entire struct from JSON
}

type LoggingConfig struct {
	Level  string `ssm:"level" env:"LOG_LEVEL"`
	Format string `ssm:"format" env:"LOG_FORMAT"`
}

// AdvancedNestedConfig shows multiple levels of nesting.
// SSM Parameters:
//   /myapp/database/host = "db.example.com"
//   /myapp/database/port = "5432"
//   /myapp/database/ssl = "true"
//   /myapp/database/replicas = `["replica1.example.com","replica2.example.com"]`
//   /myapp/cache/redis = `{"host":"redis.example.com","port":6379,"db":0}`
//   /myapp/logging/level = "info"
//   /myapp/logging/format = "json"
type AdvancedNestedConfig struct {
	Database DatabaseConfig `ssm:"database" required:"true"`
	Cache    CacheConfig    `ssm:"cache" required:"true"`
	Logging  LoggingConfig  `ssm:"logging"`
}

func main() {
	ctx := context.Background()

	cfg, err := ssmconfig.Load[AdvancedNestedConfig](ctx, "/myapp/",
		ssmconfig.WithStrictMode(true),
		ssmconfig.WithLogger(func(format string, args ...interface{}) {
			fmt.Printf("[CONFIG] "+format+"\n", args...)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Database: %s:%d (SSL: %t)\n",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.SSL)
	fmt.Printf("Replicas: %v\n", cfg.Database.Replicas)
	fmt.Printf("Redis: %s:%d (DB: %d)\n",
		cfg.Cache.Redis.Host, cfg.Cache.Redis.Port, cfg.Cache.Redis.DB)
	fmt.Printf("Logging: %s (%s)\n", cfg.Logging.Level, cfg.Logging.Format)
}

