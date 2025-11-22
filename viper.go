package ssmconfig

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ViperRemoteProvider implements Viper's remote provider interface for AWS SSM Parameter Store.
// This allows ssmconfig to be used as a remote provider with Viper.
type ViperRemoteProvider struct {
	providerName string
	endpoint     string
	path         string
	secretKeyring string
	loader       *Loader
	mu           sync.RWMutex
	values       map[string]string
	ctx          context.Context
	cancel       context.CancelFunc
}

// Provider returns the provider name for Viper.
func (v *ViperRemoteProvider) Provider() string {
	return v.providerName
}

// Endpoint returns the endpoint (SSM region or endpoint URL).
func (v *ViperRemoteProvider) Endpoint() string {
	return v.endpoint
}

// Path returns the SSM parameter path prefix.
func (v *ViperRemoteProvider) Path() string {
	return v.path
}

// SecretKeyring returns the secret keyring (not used for SSM, but required by interface).
func (v *ViperRemoteProvider) SecretKeyring() string {
	return v.secretKeyring
}

// Get retrieves a value from SSM Parameter Store.
// This method is called by Viper to get configuration values.
func (v *ViperRemoteProvider) Get(key string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Convert Viper key (dot notation) to SSM path format
	ssmKey := v.convertKeyToSSMPath(key)
	
	if val, ok := v.values[ssmKey]; ok {
		return val, nil
	}

	return "", fmt.Errorf("key %s not found in SSM Parameter Store", key)
}

// GetType returns the type of the remote provider.
func (v *ViperRemoteProvider) GetType() string {
	return v.providerName
}

// WatchRemoteProviderOnChannel watches for changes and sends updates to the channel.
// This implements Viper's watch functionality.
func (v *ViperRemoteProvider) WatchRemoteProviderOnChannel() error {
	// Viper's watch mechanism - we'll poll SSM periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-v.ctx.Done():
			return nil
		case <-ticker.C:
			if err := v.refresh(); err != nil {
				return err
			}
		}
	}
}

// refresh reloads all parameters from SSM Parameter Store.
func (v *ViperRemoteProvider) refresh() error {
	values, err := v.loader.loadByPrefix(v.ctx, v.path)
	if err != nil {
		return fmt.Errorf("refreshing SSM parameters: %w", err)
	}

	v.mu.Lock()
	v.values = values
	v.mu.Unlock()

	return nil
}

// convertKeyToSSMPath converts a Viper key (dot notation) to SSM path format.
// Example: "database.url" -> "database/url"
func (v *ViperRemoteProvider) convertKeyToSSMPath(key string) string {
	// Remove the path prefix if it's already included
	key = strings.TrimPrefix(key, v.path)
	key = strings.TrimPrefix(key, "/")
	
	// Convert dot notation to slash notation
	return strings.ReplaceAll(key, ".", "/")
}

// Stop stops watching for changes.
func (v *ViperRemoteProvider) Stop() {
	if v.cancel != nil {
		v.cancel()
	}
}

// NewViperRemoteProvider creates a new Viper remote provider for AWS SSM Parameter Store.
// The providerName should be "ssm" or "awsssm" to identify it as an SSM provider.
// The endpoint can be the AWS region (e.g., "us-east-1") or left empty to use default.
// The path is the SSM parameter path prefix (e.g., "/myapp/config").
func NewViperRemoteProvider(ctx context.Context, providerName, endpoint, path string, opts ...LoaderOption) (*ViperRemoteProvider, error) {
	loader, err := NewLoader(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating SSM loader: %w", err)
	}

	refreshCtx, cancel := context.WithCancel(ctx)

	provider := &ViperRemoteProvider{
		providerName:  providerName,
		endpoint:      endpoint,
		path:          path,
		secretKeyring: "",
		loader:        loader,
		values:        make(map[string]string),
		ctx:           refreshCtx,
		cancel:        cancel,
	}

	// Initial load
	if err := provider.refresh(); err != nil {
		cancel()
		return nil, fmt.Errorf("initial SSM parameter load: %w", err)
	}

	return provider, nil
}

// ViperRemoteProviderOption configures a ViperRemoteProvider.
type ViperRemoteProviderOption func(*ViperRemoteProvider)

// WithViperSecretKeyring sets the secret keyring (for compatibility with Viper interface).
func WithViperSecretKeyring(keyring string) ViperRemoteProviderOption {
	return func(v *ViperRemoteProvider) {
		v.secretKeyring = keyring
	}
}

// ReadRemoteConfig reads all SSM parameters and returns them as a map.
// This is a helper function that can be used to populate Viper with SSM values.
// The keys are converted from SSM path format (with slashes) to Viper dot notation.
// Example usage with Viper:
//   values, err := ssmconfig.ReadRemoteConfig(ctx, "/myapp/config")
//   if err != nil {
//       log.Fatal(err)
//   }
//   for key, value := range values {
//       viper.Set(key, value)
//   }
//   // Or use viper.MergeConfigMap(values)
func ReadRemoteConfig(ctx context.Context, prefix string, opts ...LoaderOption) (map[string]interface{}, error) {
	loader, err := NewLoader(ctx, opts...)
	if err != nil {
		return nil, err
	}

	values, err := loader.loadByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Convert flat map to nested map structure for Viper
	result := make(map[string]interface{})
	for key, value := range values {
		// Convert SSM path format to Viper dot notation
		viperKey := strings.ReplaceAll(key, "/", ".")
		result[viperKey] = value
	}

	return result, nil
}

// SetViperRemoteProvider sets up Viper to use SSM Parameter Store as a remote provider.
// This is a convenience function that integrates ssmconfig with Viper.
// Returns a provider that implements Viper's remote provider interface.
func SetViperRemoteProvider(ctx context.Context, prefix string, opts ...LoaderOption) (*ViperRemoteProvider, error) {
	return NewViperRemoteProvider(ctx, "awsssm", "", prefix, opts...)
}

// GetViperValues returns all SSM parameter values in a format suitable for Viper.
// Keys are converted from SSM path format to Viper dot notation.
// This can be used with viper.Set() or viper.MergeConfigMap().
func (v *ViperRemoteProvider) GetViperValues() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	result := make(map[string]interface{})
	for key, value := range v.values {
		// Convert SSM path format to Viper dot notation
		viperKey := strings.ReplaceAll(key, "/", ".")
		result[viperKey] = value
	}

	return result
}

