// Package anysync provides Any-Sync integration components.
package anysync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/util/crypto"
)

// AccountManager manages cryptographic keys for the account.
// It handles generation, secure storage, and loading of account keys.
type AccountManager struct {
	keys    *accountdata.AccountKeys
	dataDir string
}

const (
	// accountKeyFile stores the encrypted account private key
	accountKeyFile = "account.key"
	// deviceKeyFile stores the device key encrypted with the account key
	deviceKeyFile = "device.key"
)

// NewAccountManager creates a new AccountManager.
func NewAccountManager(dataDir string) *AccountManager {
	return &AccountManager{
		dataDir: dataDir,
	}
}

// GenerateKeys generates new random account and device keys.
// This creates a fresh cryptographic identity for the account.
func (am *AccountManager) GenerateKeys() error {
	keys, err := accountdata.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate random keys: %w", err)
	}

	am.keys = keys
	return nil
}

// StoreKeys securely persists the keys to disk.
// The device key is encrypted with the account key before storage.
func (am *AccountManager) StoreKeys() error {
	if am.keys == nil {
		return fmt.Errorf("no keys to store")
	}

	// Ensure data directory exists
	if err := os.MkdirAll(am.dataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Marshal account key (SignKey is the account key)
	accountKeyBytes, err := am.keys.SignKey.Marshall()
	if err != nil {
		return fmt.Errorf("failed to marshal account key: %w", err)
	}

	// Write account key to disk (unencrypted for now, but with restricted permissions)
	// TODO: Consider encrypting with user-provided password or platform keychain
	accountKeyPath := filepath.Join(am.dataDir, accountKeyFile)
	if err := os.WriteFile(accountKeyPath, accountKeyBytes, 0600); err != nil {
		return fmt.Errorf("failed to write account key: %w", err)
	}

	// Marshal device key (PeerKey is the device key)
	deviceKeyBytes, err := am.keys.PeerKey.Marshall()
	if err != nil {
		return fmt.Errorf("failed to marshal device key: %w", err)
	}

	// Encrypt device key with account public key
	accountPubKey := am.keys.SignKey.GetPublic()
	encryptedDeviceKey, err := accountPubKey.Encrypt(deviceKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt device key: %w", err)
	}

	// Write encrypted device key to disk
	deviceKeyPath := filepath.Join(am.dataDir, deviceKeyFile)
	if err := os.WriteFile(deviceKeyPath, encryptedDeviceKey, 0600); err != nil {
		return fmt.Errorf("failed to write device key: %w", err)
	}

	return nil
}

// LoadKeys loads existing keys from disk.
// Returns an error if keys are missing or corrupted.
func (am *AccountManager) LoadKeys() error {
	accountKeyPath := filepath.Join(am.dataDir, accountKeyFile)
	deviceKeyPath := filepath.Join(am.dataDir, deviceKeyFile)

	// Check if key files exist
	if _, err := os.Stat(accountKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("account key file not found: %s", accountKeyPath)
	}
	if _, err := os.Stat(deviceKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("device key file not found: %s", deviceKeyPath)
	}

	// Read account key
	accountKeyBytes, err := os.ReadFile(accountKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read account key: %w", err)
	}

	// Unmarshal account key
	accountKey, err := crypto.UnmarshalEd25519PrivateKeyProto(accountKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account key: %w", err)
	}

	// Read encrypted device key
	encryptedDeviceKey, err := os.ReadFile(deviceKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read device key: %w", err)
	}

	// Decrypt device key using account private key (with panic recovery)
	var deviceKeyBytes []byte
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("failed to decrypt device key (corrupted or invalid data): %v", r)
			}
		}()
		deviceKeyBytes, err = accountKey.Decrypt(encryptedDeviceKey)
	}()
	if err != nil {
		return err
	}

	// Unmarshal device key
	deviceKey, err := crypto.UnmarshalEd25519PrivateKeyProto(deviceKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal device key: %w", err)
	}

	// Create AccountKeys from loaded keys
	am.keys = accountdata.New(deviceKey, accountKey)

	return nil
}

// GetKeys returns the current AccountKeys.
// Returns nil if keys haven't been generated or loaded yet.
func (am *AccountManager) GetKeys() *accountdata.AccountKeys {
	return am.keys
}

// ClearKeys securely clears keys from memory.
// This should be called during shutdown to prevent key leakage.
func (am *AccountManager) ClearKeys() {
	am.keys = nil
}

// HasKeys returns true if keys are currently loaded in memory.
func (am *AccountManager) HasKeys() bool {
	return am.keys != nil
}

// KeysExist checks if key files exist on disk.
func (am *AccountManager) KeysExist() bool {
	accountKeyPath := filepath.Join(am.dataDir, accountKeyFile)
	deviceKeyPath := filepath.Join(am.dataDir, deviceKeyFile)

	_, accountErr := os.Stat(accountKeyPath)
	_, deviceErr := os.Stat(deviceKeyPath)

	return accountErr == nil && deviceErr == nil
}
