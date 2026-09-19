package domain

import (
	"strings"
	"time"
)

// Client represents a customer/client entity in the database.
type Client struct {
	ID             int64      `json:"id"`
	PICName        string     `json:"pic_name"`
	CompanyName    string     `json:"company_name"`
	CompanyAddress string     `json:"company_address"`
	CompanyEmail   string     `json:"company_email"`
	ProductOrdered string     `json:"product_ordered"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// GetProductList returns ordered products as a slice of trimmed strings.
func (c *Client) GetProductList() []string {
	if strings.TrimSpace(c.ProductOrdered) == "" {
		return nil
	}
	parts := strings.Split(c.ProductOrdered, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// SetProductList sets the comma-separated product_ordered field from a slice.
func (c *Client) SetProductList(products []string) {
	var clean []string
	for _, p := range products {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	c.ProductOrdered = strings.Join(clean, ", ")
}

// HasProduct checks if the client already ordered the given product name.
func (c *Client) HasProduct(productName string) bool {
	for _, p := range c.GetProductList() {
		if strings.EqualFold(p, strings.TrimSpace(productName)) {
			return true
		}
	}
	return false
}

// AddProduct appends a product name if not already present.
func (c *Client) AddProduct(productName string) {
	if !c.HasProduct(productName) {
		list := c.GetProductList()
		list = append(list, strings.TrimSpace(productName))
		c.SetProductList(list)
	}
}

// RemoveProduct removes a product name from the client's list.
func (c *Client) RemoveProduct(productName string) {
	list := c.GetProductList()
	var updated []string
	for _, p := range list {
		if !strings.EqualFold(p, strings.TrimSpace(productName)) {
			updated = append(updated, p)
		}
	}
	c.SetProductList(updated)
}

