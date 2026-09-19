package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Product represents an item/service product entity in the database.
type Product struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Price     float64    `json:"price"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ParsePrice parses user input for product price supporting raw digits, 'k'/'K' (thousands), 'm'/'M' (millions), and 'Rp' prefixes.
func ParsePrice(input string) (float64, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, errors.New("harga tidak boleh kosong")
	}

	// Remove common currency prefixes and symbols
	s = strings.TrimPrefix(s, "Rp")
	s = strings.TrimPrefix(s, "rp")
	s = strings.TrimPrefix(s, "RP")
	s = strings.TrimPrefix(s, "IDR")
	s = strings.TrimPrefix(s, "idr")
	s = strings.TrimSpace(s)

	// Check suffix: k/K (thousands) or m/M/jt/JT (millions)
	multiplier := 1.0
	lower := strings.ToLower(s)

	if strings.HasSuffix(lower, "k") || strings.HasSuffix(lower, "rb") || strings.HasSuffix(lower, "ribu") {
		multiplier = 1000.0
		s = strings.TrimSuffix(lower, "ribu")
		s = strings.TrimSuffix(s, "rb")
		s = strings.TrimSuffix(s, "k")
	} else if strings.HasSuffix(lower, "m") || strings.HasSuffix(lower, "jt") || strings.HasSuffix(lower, "juta") {
		multiplier = 1000000.0
		s = strings.TrimSuffix(lower, "juta")
		s = strings.TrimSuffix(s, "jt")
		s = strings.TrimSuffix(s, "m")
	}

	s = strings.TrimSpace(s)

	// If contains dot or comma as thousand separator (e.g. 400.000 or 400,000) when multiplier is 1
	if multiplier == 1.0 {
		if strings.Count(s, ".") > 0 && !strings.Contains(s, ",") {
			parts := strings.Split(s, ".")
			if len(parts) > 1 && len(parts[len(parts)-1]) == 3 {
				s = strings.ReplaceAll(s, ".", "")
			}
		} else if strings.Count(s, ",") > 0 && !strings.Contains(s, ".") {
			parts := strings.Split(s, ",")
			if len(parts) > 1 && len(parts[len(parts)-1]) == 3 {
				s = strings.ReplaceAll(s, ",", "")
			}
		}
	} else {
		// Decimal point for multiplier (e.g. 1.5k or 1,5k)
		s = strings.ReplaceAll(s, ",", ".")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("format harga '%s' tidak valid: %w", input, err)
	}

	if val < 0 {
		return 0, errors.New("harga tidak boleh negatif")
	}

	return val * multiplier, nil
}

