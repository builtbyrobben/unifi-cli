package cmd

import (
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/api"
	"github.com/builtbyrobben/unifi-cli/internal/secrets"
	"github.com/builtbyrobben/unifi-cli/internal/unifi"
)

func getUniFiClient(flags *RootFlags) (*unifi.Client, error) {
	host := os.Getenv("UNIFI_HOST")
	username := os.Getenv("UNIFI_USERNAME")
	password := os.Getenv("UNIFI_PASSWORD")
	site := os.Getenv("UNIFI_SITE")

	// Fill in from keyring if not set via env
	if host == "" || username == "" || password == "" {
		store, err := secrets.OpenDefault()
		if err != nil {
			return nil, fmt.Errorf("open credential store: %w", err)
		}

		if host == "" {
			host, err = store.Get(secrets.KeyHost)
			if err != nil {
				return nil, fmt.Errorf("get host: %w (run 'unifi-cli auth set-credentials')", err)
			}
		}

		if username == "" {
			username, err = store.Get(secrets.KeyUsername)
			if err != nil {
				return nil, fmt.Errorf("get username: %w (run 'unifi-cli auth set-credentials')", err)
			}
		}

		if password == "" {
			password, err = store.Get(secrets.KeyPassword)
			if err != nil {
				return nil, fmt.Errorf("get password: %w (run 'unifi-cli auth set-credentials')", err)
			}
		}

		if site == "" {
			storedSite, _ := store.Get(secrets.KeySite)
			if storedSite != "" {
				site = storedSite
			}
		}
	}

	// CLI flag overrides everything
	if flags.Site != "" {
		site = flags.Site
	}

	if site == "" {
		site = "default"
	}

	creds := api.Credentials{
		Host:     host,
		Username: username,
		Password: password,
		Site:     site,
	}

	apiClient, err := api.NewClient(creds,
		api.WithInsecure(flags.Insecure),
		api.WithUserAgent("unifi-cli/1.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}

	return unifi.NewClient(apiClient), nil
}
