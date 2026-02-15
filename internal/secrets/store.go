package secrets

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/builtbyrobben/unifi-cli/internal/config"
)

// Store provides access to stored credentials.
type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	Has(key string) (bool, error)
}

// KeyringStore stores credentials in the OS keyring.
type KeyringStore struct {
	ring keyring.Keyring
}

const (
	// Credential keys stored in the keyring.
	KeyHost     = "host"
	KeyUsername = "username"
	KeyPassword = "password"
	KeySite     = "site"

	keyringPasswordEnv = "UNIFI_CLI_KEYRING_PASS" //nolint:gosec // env var name, not a credential
	keyringBackendEnv  = "UNIFI_CLI_KEYRING_BACKEND"
	keyringOpenTimeout = 5 * time.Second
)

var (
	errMissingKey            = errors.New("missing key")
	errMissingValue          = errors.New("missing value")
	errNoTTY                 = errors.New("no TTY available for keyring file backend password prompt")
	errInvalidKeyringBackend = errors.New("invalid keyring backend")
	errKeyringTimeout        = errors.New("keyring connection timed out")
)

// KeyringBackendInfo describes the resolved keyring backend.
type KeyringBackendInfo struct {
	Value  string
	Source string
}

const (
	keyringBackendSourceEnv     = "env"
	keyringBackendSourceDefault = "default"
	keyringBackendAuto          = "auto"
)

// ResolveKeyringBackendInfo determines which keyring backend to use.
func ResolveKeyringBackendInfo() (KeyringBackendInfo, error) {
	if v := normalizeKeyringBackend(os.Getenv(keyringBackendEnv)); v != "" {
		return KeyringBackendInfo{Value: v, Source: keyringBackendSourceEnv}, nil
	}

	return KeyringBackendInfo{Value: keyringBackendAuto, Source: keyringBackendSourceDefault}, nil
}

func allowedBackends(info KeyringBackendInfo) ([]keyring.BackendType, error) {
	switch info.Value {
	case "", keyringBackendAuto:
		return nil, nil
	case "keychain":
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case "file":
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("%w: %q (expected %s, keychain, or file)", errInvalidKeyringBackend, info.Value, keyringBackendAuto)
	}
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	return fileKeyringPasswordFuncFrom(os.Getenv(keyringPasswordEnv), term.IsTerminal(int(os.Stdin.Fd())))
}

func fileKeyringPasswordFuncFrom(password string, isTTY bool) keyring.PromptFunc {
	if password != "" {
		return keyring.FixedStringPrompt(password)
	}

	if isTTY {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
}

func normalizeKeyringBackend(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func shouldForceFileBackend(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == keyringBackendAuto && dbusAddr == ""
}

func shouldUseKeyringTimeout(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == "auto" && dbusAddr != ""
}

func openKeyring() (keyring.Keyring, error) {
	keyringDir, err := config.EnsureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}

	backendInfo, err := ResolveKeyringBackendInfo()
	if err != nil {
		return nil, err
	}

	backends, err := allowedBackends(backendInfo)
	if err != nil {
		return nil, err
	}

	dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")

	if shouldForceFileBackend(runtime.GOOS, backendInfo, dbusAddr) {
		backends = []keyring.BackendType{keyring.FileBackend}
	}

	cfg := keyring.Config{
		ServiceName:              config.AppName,
		KeychainTrustApplication: false,
		AllowedBackends:          backends,
		FileDir:                  keyringDir,
		FilePasswordFunc:         fileKeyringPasswordFunc(),
	}

	if shouldUseKeyringTimeout(runtime.GOOS, backendInfo, dbusAddr) {
		return openKeyringWithTimeout(cfg, keyringOpenTimeout)
	}

	ring, err := keyring.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return ring, nil
}

type keyringResult struct {
	ring keyring.Keyring
	err  error
}

func openKeyringWithTimeout(cfg keyring.Config, timeout time.Duration) (keyring.Keyring, error) {
	ch := make(chan keyringResult, 1)

	go func() {
		ring, err := keyring.Open(cfg)
		ch <- keyringResult{ring, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("open keyring: %w", res.err)
		}

		return res.ring, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("%w after %v (D-Bus SecretService may be unresponsive); "+
			"set UNIFI_CLI_KEYRING_BACKEND=file and UNIFI_CLI_KEYRING_PASS=<password> to use encrypted file storage instead",
			errKeyringTimeout, timeout)
	}
}

// OpenDefault opens the default keyring-backed credential store.
func OpenDefault() (Store, error) {
	ring, err := openKeyring()
	if err != nil {
		return nil, err
	}

	return &KeyringStore{ring: ring}, nil
}

// Get retrieves a credential by key.
func (s *KeyringStore) Get(key string) (string, error) {
	key = strings.TrimSpace(key)

	if key == "" {
		return "", errMissingKey
	}

	item, err := s.ring.Get(key)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", key, err)
	}

	return string(item.Data), nil
}

// Set stores a credential by key.
func (s *KeyringStore) Set(key, value string) error {
	key = strings.TrimSpace(key)

	if key == "" {
		return errMissingKey
	}

	value = strings.TrimSpace(value)

	if value == "" {
		return errMissingValue
	}

	if err := s.ring.Set(keyring.Item{
		Key:  key,
		Data: []byte(value),
	}); err != nil {
		return fmt.Errorf("store %s: %w", key, err)
	}

	return nil
}

// Delete removes a credential by key.
func (s *KeyringStore) Delete(key string) error {
	if err := s.ring.Remove(key); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return fmt.Errorf("delete %s: %w", key, err)
	}

	return nil
}

// Has checks if a credential exists.
func (s *KeyringStore) Has(key string) (bool, error) {
	_, err := s.ring.Get(key)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("check %s: %w", key, err)
	}

	return true, nil
}

// AllCredentialKeys returns all credential key names used by UniFi CLI.
func AllCredentialKeys() []string {
	return []string{KeyHost, KeyUsername, KeyPassword, KeySite}
}
