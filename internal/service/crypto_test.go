package service_test

import (
	"bytes"
	"hs1-bot/internal/service"
	"testing"
)

func TestCrypto_CompressionAndDecompression(t *testing.T) {
	originalData := []byte("CREATE TABLE users (id SERIAL PRIMARY KEY, username VARCHAR(50) NOT NULL); INSERT INTO users VALUES (1, 'kava');")

	compressed, err := service.CompressGzip(originalData)
	if err != nil {
		t.Fatalf("failed to compress gzip: %v", err)
	}

	if len(compressed) == 0 {
		t.Fatal("compressed bytes should not be empty")
	}

	decompressed, err := service.DecompressGzip(compressed)
	if err != nil {
		t.Fatalf("failed to decompress gzip: %v", err)
	}

	if !bytes.Equal(originalData, decompressed) {
		t.Errorf("decompressed data does not match original: expected %q, got %q", string(originalData), string(decompressed))
	}
}

func TestCrypto_KeyGenerationAndEncryptionCycle(t *testing.T) {
	rawKey, keyHex, err := service.GenerateAES256Key()
	if err != nil {
		t.Fatalf("failed to generate AES-256 key: %v", err)
	}

	if len(rawKey) != 32 {
		t.Fatalf("expected key length 32 bytes, got %d", len(rawKey))
	}
	if len(keyHex) != 64 {
		t.Fatalf("expected hex key length 64 chars, got %d", len(keyHex))
	}

	plaintext := []byte("SUPER_SECRET_DATABASE_BACKUP_SQL_DATA_12345")

	// 1. Encrypt
	ciphertext, err := service.EncryptAESGCM(plaintext, rawKey)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	// 2. Decrypt with correct key
	decrypted, err := service.DecryptAESGCM(ciphertext, rawKey)
	if err != nil {
		t.Fatalf("decryption with correct key failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("expected decrypted text %q, got %q", string(plaintext), string(decrypted))
	}

	// 3. Decrypt with wrong key should fail
	wrongKey, _, _ := service.GenerateAES256Key()
	_, err = service.DecryptAESGCM(ciphertext, wrongKey)
	if err == nil {
		t.Error("expected decryption with wrong key to return error, got nil")
	}
}
