package anysync

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateKeys verifies that new keys can be generated successfully.
func TestGenerateKeys(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	if err := am.GenerateKeys(); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}

	if !am.HasKeys() {
		t.Fatal("Keys should be loaded after generation")
	}

	keys := am.GetKeys()
	if keys == nil {
		t.Fatal("GetKeys returned nil after generation")
	}
	if keys.PeerKey == nil {
		t.Fatal("PeerKey is nil")
	}
	if keys.SignKey == nil {
		t.Fatal("SignKey is nil")
	}
	if keys.PeerId == "" {
		t.Fatal("PeerId is empty")
	}
}

// TestStoreAndLoadKeys verifies that keys can be stored and loaded from disk.
func TestStoreAndLoadKeys(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	// Generate and store keys
	if err := am.GenerateKeys(); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}

	originalKeys := am.GetKeys()
	originalPeerId := originalKeys.PeerId

	if err := am.StoreKeys(); err != nil {
		t.Fatalf("StoreKeys failed: %v", err)
	}

	// Verify files were created
	accountKeyPath := filepath.Join(tmpDir, accountKeyFile)
	deviceKeyPath := filepath.Join(tmpDir, deviceKeyFile)
	if _, err := os.Stat(accountKeyPath); os.IsNotExist(err) {
		t.Fatal("Account key file was not created")
	}
	if _, err := os.Stat(deviceKeyPath); os.IsNotExist(err) {
		t.Fatal("Device key file was not created")
	}

	// Clear keys and load them back
	am.ClearKeys()
	if am.HasKeys() {
		t.Fatal("Keys should be cleared")
	}

	if err := am.LoadKeys(); err != nil {
		t.Fatalf("LoadKeys failed: %v", err)
	}

	// Verify loaded keys match original
	loadedKeys := am.GetKeys()
	if loadedKeys == nil {
		t.Fatal("GetKeys returned nil after loading")
	}
	if loadedKeys.PeerId != originalPeerId {
		t.Fatalf("PeerId mismatch: got %s, want %s", loadedKeys.PeerId, originalPeerId)
	}
}

// TestLoadKeysPersistenceAcrossRestarts verifies keys persist across manager restarts.
func TestLoadKeysPersistenceAcrossRestarts(t *testing.T) {
	tmpDir := t.TempDir()

	// First manager: generate and store keys
	am1 := NewAccountManager(tmpDir)
	if err := am1.GenerateKeys(); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}
	originalPeerId := am1.GetKeys().PeerId
	if err := am1.StoreKeys(); err != nil {
		t.Fatalf("StoreKeys failed: %v", err)
	}

	// Second manager: load existing keys
	am2 := NewAccountManager(tmpDir)
	if err := am2.LoadKeys(); err != nil {
		t.Fatalf("LoadKeys failed on second manager: %v", err)
	}

	loadedPeerId := am2.GetKeys().PeerId
	if loadedPeerId != originalPeerId {
		t.Fatalf("PeerId mismatch across restarts: got %s, want %s", loadedPeerId, originalPeerId)
	}
}

// TestLoadKeysErrorMissingFiles verifies error handling for missing key files.
func TestLoadKeysErrorMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	// Try to load without creating files
	err := am.LoadKeys()
	if err == nil {
		t.Fatal("LoadKeys should fail when key files are missing")
	}
	if am.HasKeys() {
		t.Fatal("Keys should not be loaded after failed load")
	}
}

// TestLoadKeysErrorCorruptedAccountKey verifies error handling for corrupted account key.
func TestLoadKeysErrorCorruptedAccountKey(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	// Create corrupted account key file
	accountKeyPath := filepath.Join(tmpDir, accountKeyFile)
	if err := os.WriteFile(accountKeyPath, []byte("corrupted data"), 0600); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Create valid device key file (though it won't be used)
	deviceKeyPath := filepath.Join(tmpDir, deviceKeyFile)
	if err := os.WriteFile(deviceKeyPath, []byte("some data"), 0600); err != nil {
		t.Fatalf("Failed to create device key file: %v", err)
	}

	// Try to load corrupted keys
	err := am.LoadKeys()
	if err == nil {
		t.Fatal("LoadKeys should fail when account key is corrupted")
	}
	if am.HasKeys() {
		t.Fatal("Keys should not be loaded after failed load")
	}
}

// TestLoadKeysErrorCorruptedDeviceKey verifies error handling for corrupted device key.
func TestLoadKeysErrorCorruptedDeviceKey(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	// Generate valid keys and store them
	if err := am.GenerateKeys(); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}
	if err := am.StoreKeys(); err != nil {
		t.Fatalf("StoreKeys failed: %v", err)
	}

	// Corrupt the device key file
	deviceKeyPath := filepath.Join(tmpDir, deviceKeyFile)
	if err := os.WriteFile(deviceKeyPath, []byte("corrupted data"), 0600); err != nil {
		t.Fatalf("Failed to corrupt device key: %v", err)
	}

	// Clear keys and try to load corrupted device key
	am.ClearKeys()
	err := am.LoadKeys()
	if err == nil {
		t.Fatal("LoadKeys should fail when device key is corrupted")
	}
	if am.HasKeys() {
		t.Fatal("Keys should not be loaded after failed load")
	}
}

// TestKeysExist verifies the KeysExist method.
func TestKeysExist(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	// Initially no keys exist
	if am.KeysExist() {
		t.Fatal("KeysExist should return false initially")
	}

	// Generate and store keys
	if err := am.GenerateKeys(); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}
	if err := am.StoreKeys(); err != nil {
		t.Fatalf("StoreKeys failed: %v", err)
	}

	// Now keys should exist
	if !am.KeysExist() {
		t.Fatal("KeysExist should return true after storing")
	}
}

// TestClearKeys verifies that keys are properly cleared from memory.
func TestClearKeys(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	if err := am.GenerateKeys(); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}

	if !am.HasKeys() {
		t.Fatal("Keys should be loaded after generation")
	}

	am.ClearKeys()

	if am.HasKeys() {
		t.Fatal("Keys should be cleared")
	}
	if am.GetKeys() != nil {
		t.Fatal("GetKeys should return nil after clearing")
	}
}

// TestStoreKeysWithoutGeneration verifies error handling when storing without keys.
func TestStoreKeysWithoutGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	am := NewAccountManager(tmpDir)

	err := am.StoreKeys()
	if err == nil {
		t.Fatal("StoreKeys should fail when no keys are generated")
	}
}
