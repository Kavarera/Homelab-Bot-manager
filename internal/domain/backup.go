package domain

import (
	"time"
)

// PostgresContainer represents a discovered PostgreSQL Docker container on VPS.
type PostgresContainer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

// BackupResult encapsulates the output of a remote database dump, compression, and encryption.
type BackupResult struct {
	ContainerName      string    `json:"container_name"`
	Filename           string    `json:"filename"`
	LocalPath          string    `json:"local_path"`
	EncryptedData      []byte    `json:"-"`
	KeyHex             string    `json:"key_hex"`
	RawSizeBytes       int64     `json:"raw_size_bytes"`
	EncryptedSizeBytes int64     `json:"encrypted_size_bytes"`
	CreatedAt          time.Time `json:"created_at"`
}
