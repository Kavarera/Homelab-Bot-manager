package config

import (
	"os"
	"testing"
)

func TestParse_Success(t *testing.T) {
	os.Setenv("TELEGRAM_TOKEN", "mock-token-123")
	os.Setenv("ALLOWED_USER_ID", "12345678")
	os.Setenv("SERVER_NAME", "TestServer")
	os.Setenv("SSH_HOST", "user@host.local")
	os.Setenv("SSH_PRIVATE_KEY_PATH", "/tmp/id_rsa")
	defer func() {
		os.Unsetenv("TELEGRAM_TOKEN")
		os.Unsetenv("ALLOWED_USER_ID")
		os.Unsetenv("SERVER_NAME")
		os.Unsetenv("SSH_HOST")
		os.Unsetenv("SSH_PRIVATE_KEY_PATH")
	}()

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.TelegramToken != "mock-token-123" {
		t.Errorf("expected token 'mock-token-123', got '%s'", cfg.TelegramToken)
	}
	if cfg.AllowedUserID != 12345678 {
		t.Errorf("expected allowed user ID 12345678, got %d", cfg.AllowedUserID)
	}
	if cfg.ServerName != "TestServer" {
		t.Errorf("expected server name 'TestServer', got '%s'", cfg.ServerName)
	}
	if cfg.SSHHost != "user@host.local" {
		t.Errorf("expected ssh host 'user@host.local', got '%s'", cfg.SSHHost)
	}
	if cfg.SSHPrivateKeyPath != "/tmp/id_rsa" {
		t.Errorf("expected ssh key path '/tmp/id_rsa', got '%s'", cfg.SSHPrivateKeyPath)
	}
}

func TestParse_FallbackChatIDAndDefaults(t *testing.T) {
	os.Setenv("TELEGRAM_TOKEN", "mock-token-123")
	os.Unsetenv("ALLOWED_USER_ID")
	os.Setenv("ALLOWED_CHAT_ID", "87654321")
	os.Unsetenv("SERVER_NAME")
	os.Unsetenv("SSH_HOST")
	os.Unsetenv("SSH_PRIVATE_KEY_PATH")
	defer func() {
		os.Unsetenv("TELEGRAM_TOKEN")
		os.Unsetenv("ALLOWED_CHAT_ID")
	}()

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("expected no error with fallback chat id, got: %v", err)
	}

	if cfg.AllowedUserID != 87654321 {
		t.Errorf("expected allowed user ID 87654321, got %d", cfg.AllowedUserID)
	}
	if cfg.ServerName != "Local Server" {
		t.Errorf("expected fallback server name 'Local Server', got '%s'", cfg.ServerName)
	}
	if cfg.SSHHost != "hs1bot@127.0.0.1" {
		t.Errorf("expected fallback ssh host, got '%s'", cfg.SSHHost)
	}
}

func TestParse_MissingToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_TOKEN")
	os.Setenv("ALLOWED_USER_ID", "12345678")
	defer os.Unsetenv("ALLOWED_USER_ID")

	_, err := Parse()
	if err == nil {
		t.Error("expected error for missing TELEGRAM_TOKEN, got nil")
	}
}

func TestParse_MissingUserID(t *testing.T) {
	os.Setenv("TELEGRAM_TOKEN", "mock-token")
	os.Unsetenv("ALLOWED_USER_ID")
	os.Unsetenv("ALLOWED_CHAT_ID")
	defer os.Unsetenv("TELEGRAM_TOKEN")

	_, err := Parse()
	if err == nil {
		t.Error("expected error for missing user id, got nil")
	}
}

func TestParse_InvalidUserID(t *testing.T) {
	os.Setenv("TELEGRAM_TOKEN", "mock-token")
	os.Setenv("ALLOWED_USER_ID", "not-a-number")
	defer func() {
		os.Unsetenv("TELEGRAM_TOKEN")
		os.Unsetenv("ALLOWED_USER_ID")
	}()

	_, err := Parse()
	if err == nil {
		t.Error("expected error for non-integer user id, got nil")
	}
}

func TestLoadEnv_FindsFile(t *testing.T) {
	path, err := LoadEnv()
	if err != nil {
		t.Logf("LoadEnv result: %v (might be expected if no .env in root)", err)
	} else {
		t.Logf("Found .env at: %s", path)
	}
}
