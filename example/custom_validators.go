// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ali63yavari/ssmconfig"
)

// ValidatedConfig demonstrates custom validators.
// SSM Parameters:
//   /myapp/email = "user@example.com"
//   /myapp/website = "https://example.com"
//   /myapp/username = "alice123"
//   /myapp/port = "8080"
type ValidatedConfig struct {
	Email    string `ssm:"email" validate:"email"`
	Website  string `ssm:"website" validate:"url"`
	Username string `ssm:"username" validate:"alphanumeric,minlen:3,maxlen:20"`
	Port     int    `ssm:"port" validate:"min:1,max:65535"`
}

func main() {
	ctx := context.Background()

	// Register custom validators
	registerCustomValidators()

	cfg, err := ssmconfig.Load[ValidatedConfig](ctx, "/myapp/")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Email: %s\n", cfg.Email)
	fmt.Printf("Website: %s\n", cfg.Website)
	fmt.Printf("Username: %s\n", cfg.Username)
	fmt.Printf("Port: %d\n", cfg.Port)
}

func registerCustomValidators() {
	// Register alphanumeric validator
	ssmconfig.RegisterValidator("alphanumeric", func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("alphanumeric validator requires string type")
		}
		for _, r := range str {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				return fmt.Errorf("string contains non-alphanumeric characters")
			}
		}
		return nil
	})

	// Register regex validator (parameterized)
	ssmconfig.RegisterParameterizedValidator("regex", func(value interface{}, params string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("regex validator requires string type")
		}
		matched, err := regexp.MatchString(params, str)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("value does not match pattern: %s", params)
		}
		return nil
	})

	// Register domain validator
	ssmconfig.RegisterValidator("domain", func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("domain validator requires string type")
		}
		if !strings.Contains(str, ".") {
			return fmt.Errorf("invalid domain format")
		}
		parts := strings.Split(str, ".")
		if len(parts) < 2 {
			return fmt.Errorf("invalid domain format")
		}
		return nil
	})
}

