package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	TelegramToken     string
	AllowedUserID     int64
	ServerName        string
	SSHHost           string
	SSHPrivateKeyPath string
	DatabasePath      string
	BackupBasePath    string

	// SMTP Configuration
	SMTPHost        string
	SMTPPort        int
	SMTPSSL         bool
	SMTPUsername    string
	SMTPPassword    string
	SMTPSenderName  string
	SMTPSenderEmail string
}

// LoadEnv automatically finds and loads the .env file by traversing
// up the directory tree from the current working directory.
func LoadEnv() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if info, err := os.Stat(envPath); err == nil && !info.IsDir() {
			if loadErr := godotenv.Load(envPath); loadErr != nil {
				return envPath, loadErr
			}
			return envPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// Parse extracts and validates configuration directly from the current environment variables.
func Parse() (*Config, error) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_TOKEN"))
	if token == "" {
		return nil, errors.New("TELEGRAM_TOKEN is required but was empty")
	}

	userRaw := strings.TrimSpace(os.Getenv("ALLOWED_USER_ID"))
	if userRaw == "" {
		// Fallback to ALLOWED_CHAT_ID if ALLOWED_USER_ID is not set
		userRaw = strings.TrimSpace(os.Getenv("ALLOWED_CHAT_ID"))
	}

	if userRaw == "" {
		return nil, errors.New("ALLOWED_USER_ID (or ALLOWED_CHAT_ID) is required")
	}

	allowedUserID, err := strconv.ParseInt(userRaw, 10, 64)
	if err != nil || allowedUserID == 0 {
		return nil, fmt.Errorf("invalid ALLOWED_USER_ID: %q is not a valid integer ID", userRaw)
	}

	serverName := strings.TrimSpace(os.Getenv("SERVER_NAME"))
	if serverName == "" {
		serverName = "Local Server"
	}

	sshHost := strings.TrimSpace(os.Getenv("SSH_HOST"))
	if sshHost == "" {
		sshHost = "hs1bot@127.0.0.1"
	}

	sshKeyPath := strings.TrimSpace(os.Getenv("SSH_PRIVATE_KEY_PATH"))
	if sshKeyPath == "" {
		sshKeyPath = "/home/hs1bot/.ssh/bot_hs1_key"
	}

	databasePath := strings.TrimSpace(os.Getenv("DATABASE_PATH"))
	if databasePath == "" {
		databasePath = "data/bot.db"
	}

	backupBasePath := strings.TrimSpace(os.Getenv("BACKUP_BASE_PATH"))
	if backupBasePath == "" {
		backupBasePath = "/home/kava/backup_db"
	}

	// SMTP Settings
	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if smtpHost == "" {
		smtpHost = "smtp.sumopod.com"
	}

	smtpPort := 465
	if portStr := strings.TrimSpace(os.Getenv("SMTP_PORT")); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			smtpPort = p
		}
	}

	smtpSSL := true
	if sslStr := strings.TrimSpace(os.Getenv("SMTP_SSL")); sslStr != "" {
		smtpSSL = strings.ToLower(sslStr) == "true" || sslStr == "1"
	}

	smtpUsername := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	smtpPassword := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))

	smtpSenderName := strings.TrimSpace(os.Getenv("SMTP_SENDER_NAME"))
	if smtpSenderName == "" {
		smtpSenderName = "Kavarera"
	}

	smtpSenderEmail := strings.TrimSpace(os.Getenv("SMTP_SENDER_EMAIL"))
	if smtpSenderEmail == "" {
		smtpSenderEmail = smtpUsername
	}

	cfg := &Config{
		TelegramToken:     token,
		AllowedUserID:     allowedUserID,
		ServerName:        serverName,
		SSHHost:           sshHost,
		SSHPrivateKeyPath: sshKeyPath,
		DatabasePath:      databasePath,
		BackupBasePath:    backupBasePath,
		SMTPHost:          smtpHost,
		SMTPPort:          smtpPort,
		SMTPSSL:           smtpSSL,
		SMTPUsername:      smtpUsername,
		SMTPPassword:      smtpPassword,
		SMTPSenderName:    smtpSenderName,
		SMTPSenderEmail:   smtpSenderEmail,
	}

	return cfg, nil
}

// Load discovers and loads .env from current or parent directories, then parses configuration.
func Load() (*Config, error) {
	_, _ = LoadEnv()
	return Parse()
}
