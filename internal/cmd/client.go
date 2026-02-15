package cmd

import (
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/secrets"
	"github.com/builtbyrobben/unifi-cli/internal/unifi"
)

func getUniFiClient() (*unifi.Client, error) {
	// Check for environment variable override first
	apiKey := os.Getenv("UNIFI_API_KEY")

	if apiKey == "" {
		// Try to get from keyring
		store, err := secrets.OpenDefault()
		if err != nil {
			return nil, fmt.Errorf("open credential store: %w", err)
		}

		apiKey, err = store.GetAPIKey()
		if err != nil {
			return nil, fmt.Errorf("get API key: %w (set UNIFI_API_KEY or run 'unifi-cli auth set-key --stdin')", err)
		}
	}

	return unifi.NewClient(apiKey), nil
}
