package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
	"github.com/builtbyrobben/unifi-cli/internal/secrets"
)

type AuthCmd struct {
	SetCredentials AuthSetCredentialsCmd `cmd:"set-credentials" help:"Set UniFi controller credentials"`
	Status         AuthStatusCmd         `cmd:"" help:"Show authentication status"`
	Remove         AuthRemoveCmd         `cmd:"" help:"Remove stored credentials"`
}

type AuthSetCredentialsCmd struct {
	Host     string `help:"UniFi controller URL (e.g., https://192.168.1.1)" env:"UNIFI_HOST"`
	Username string `help:"Admin username" env:"UNIFI_USERNAME"`
	Password string `help:"Admin password (will prompt if not provided)" env:"UNIFI_PASSWORD"`
	SiteName string `help:"Site name" default:"default" name:"site-name"`
}

func (cmd *AuthSetCredentialsCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	host := strings.TrimSpace(cmd.Host)
	username := strings.TrimSpace(cmd.Username)
	password := strings.TrimSpace(cmd.Password)
	site := strings.TrimSpace(cmd.SiteName)

	// Interactive prompts for missing fields
	if host == "" {
		host, err = promptString("Controller URL: ")
		if err != nil {
			return fmt.Errorf("read host: %w", err)
		}
	}

	if username == "" {
		username, err = promptString("Username: ")
		if err != nil {
			return fmt.Errorf("read username: %w", err)
		}
	}

	if password == "" {
		password, err = promptPassword("Password: ")
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
	}

	// Store all credentials
	creds := map[string]string{
		secrets.KeyHost:     strings.TrimRight(host, "/"),
		secrets.KeyUsername: username,
		secrets.KeyPassword: password,
		secrets.KeySite:     site,
	}

	for key, value := range creds {
		if storeErr := store.Set(key, value); storeErr != nil {
			return fmt.Errorf("store %s: %w", key, storeErr)
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Credentials stored in keyring",
		})
	}

	fmt.Fprintln(os.Stderr, "Credentials stored in keyring")

	return nil
}

type AuthStatusCmd struct{}

func (cmd *AuthStatusCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	status := map[string]any{
		"storage_backend": "keyring",
	}

	// Check keyring credentials
	host, _ := store.Get(secrets.KeyHost)
	username, _ := store.Get(secrets.KeyUsername)
	hasPassword, _ := store.Has(secrets.KeyPassword)
	site, _ := store.Get(secrets.KeySite)

	// Check env var overrides
	envHost := os.Getenv("UNIFI_HOST")
	envUser := os.Getenv("UNIFI_USERNAME")
	envPass := os.Getenv("UNIFI_PASSWORD")
	envSite := os.Getenv("UNIFI_SITE")

	effectiveHost := firstNonEmpty(envHost, host)
	effectiveUser := firstNonEmpty(envUser, username)
	effectiveSite := firstNonEmpty(envSite, site, "default")
	hasAnyPassword := envPass != "" || hasPassword

	status["host"] = effectiveHost
	status["username"] = effectiveUser
	status["site"] = effectiveSite
	status["has_password"] = hasAnyPassword
	status["env_override"] = envHost != "" || envUser != "" || envPass != "" || envSite != ""

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, status)
	}

	fmt.Fprintf(os.Stderr, "Storage:  %s\n", status["storage_backend"])
	fmt.Fprintf(os.Stderr, "Host:     %s\n", displayValue(effectiveHost))
	fmt.Fprintf(os.Stderr, "Username: %s\n", displayValue(effectiveUser))
	fmt.Fprintf(os.Stderr, "Password: %s\n", displayBool(hasAnyPassword))
	fmt.Fprintf(os.Stderr, "Site:     %s\n", effectiveSite)

	if envHost != "" || envUser != "" || envPass != "" || envSite != "" {
		fmt.Fprintln(os.Stderr, "\nEnvironment variable overrides active")
	}

	if effectiveHost == "" || effectiveUser == "" || !hasAnyPassword {
		fmt.Fprintln(os.Stderr, "\nRun: unifi-cli auth set-credentials")
	}

	return nil
}

type AuthRemoveCmd struct{}

func (cmd *AuthRemoveCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	for _, key := range secrets.AllCredentialKeys() {
		if deleteErr := store.Delete(key); deleteErr != nil {
			return fmt.Errorf("remove %s: %w", key, deleteErr)
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "All credentials removed",
		})
	}

	fmt.Fprintln(os.Stderr, "All credentials removed")

	return nil
}

func promptString(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	var input string

	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytePassword)), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

func displayValue(v string) string {
	if v == "" {
		return "(not set)"
	}

	return v
}

func displayBool(v bool) string {
	if v {
		return "(set)"
	}

	return "(not set)"
}
