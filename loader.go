package ssmconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Loader struct {
	ssmClient *ssm.Client
	strict    bool
	logger    func(format string, args ...interface{})
}

type LoaderOption func(*Loader)

// WithStrictMode enables strict mode where missing required fields will cause a panic.
func WithStrictMode(strict bool) LoaderOption {
	return func(l *Loader) {
		l.strict = strict
	}
}

// WithLogger sets a custom logger function for logging missing required fields.
// This allows integration with logging libraries like Sentry, zap, logrus, etc.
// The logger function receives a format string and variadic arguments.
func WithLogger(logger func(format string, args ...interface{})) LoaderOption {
	return func(l *Loader) {
		l.logger = logger
	}
}

func NewLoader(ctx context.Context, opts ...LoaderOption) (*Loader, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	loader := &Loader{
		ssmClient: ssm.NewFromConfig(cfg),
		strict:    false,
		logger:    nil,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// Load loads configuration from AWS SSM Parameter Store and returns a typed struct.
// Environment variables (specified via "env" tags) will override SSM parameter values.
func Load[T any](ctx context.Context, prefix string, opts ...LoaderOption) (*T, error) {
	loader, err := NewLoader(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return LoadWithLoader[T](loader, ctx, prefix)
}

// LoadWithLoader loads configuration using an existing Loader instance.
func LoadWithLoader[T any](loader *Loader, ctx context.Context, prefix string) (*T, error) {
	values, err := loader.loadByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var result T
	if err := mapToStruct(values, &result, loader.strict, loader.logger); err != nil {
		return nil, fmt.Errorf("mapping to struct: %w", err)
	}

	return &result, nil
}

func (l *Loader) loadByPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	out := make(map[string]string)

	var nextToken *string
	for {
		resp, err := l.ssmClient.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           &prefix,
			Recursive:      ToPointerValue(true),
			WithDecryption: ToPointerValue(true),
			NextToken:      nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("fetching parameters: %w", err)
		}

		for _, p := range resp.Parameters {
			name := strings.TrimPrefix(*p.Name, prefix)
			// Remove leading slash if present
			name = strings.TrimPrefix(name, "/")
			out[name] = *p.Value
		}

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return out, nil
}
