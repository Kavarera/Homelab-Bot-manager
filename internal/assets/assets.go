package assets

import (
	_ "embed"
	"os"
)

//go:embed logo.png
var defaultLogo []byte

//go:embed signature.png
var defaultSignature []byte

// GetLogo returns the company logo bytes from disk if found, or from the embedded asset.
func GetLogo(customPath string) []byte {
	if customPath != "" {
		if data, err := os.ReadFile(customPath); err == nil && len(data) > 0 {
			return data
		}
	}
	return defaultLogo
}

// GetSignature returns the director signature bytes from disk if found, or from the embedded asset.
func GetSignature(customPath string) []byte {
	if customPath != "" {
		if data, err := os.ReadFile(customPath); err == nil && len(data) > 0 {
			return data
		}
	}
	return defaultSignature
}
