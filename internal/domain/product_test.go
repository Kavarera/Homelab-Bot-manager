package domain_test

import (
	"hs1-bot/internal/domain"
	"testing"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"400000", 400000, false},
		{"400k", 400000, false},
		{"400K", 400000, false},
		{"400rb", 400000, false},
		{"400 ribu", 400000, false},
		{"1.5M", 1500000, false},
		{"1,5jt", 1500000, false},
		{"Rp 400.000", 400000, false},
		{"IDR 500,000", 500000, false},
		{"150000", 150000, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"-500k", 0, true},
	}

	for _, tt := range tests {
		got, err := domain.ParsePrice(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParsePrice(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParsePrice(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParsePrice(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		}
	}
}

func TestClient_ProductListMethods(t *testing.T) {
	c := &domain.Client{
		ProductOrdered: "PayrollPro License (Month), Cloud Hosting",
	}

	list := c.GetProductList()
	if len(list) != 2 || list[0] != "PayrollPro License (Month)" || list[1] != "Cloud Hosting" {
		t.Fatalf("unexpected GetProductList: %#v", list)
	}

	if !c.HasProduct("Cloud Hosting") {
		t.Error("expected HasProduct('Cloud Hosting') to be true")
	}

	c.AddProduct("VPN Access")
	if !c.HasProduct("VPN Access") || len(c.GetProductList()) != 3 {
		t.Errorf("expected product to be added, got %#v", c.GetProductList())
	}

	c.RemoveProduct("Cloud Hosting")
	if c.HasProduct("Cloud Hosting") || len(c.GetProductList()) != 2 {
		t.Errorf("expected product to be removed, got %#v", c.GetProductList())
	}
}
