package domain

import (
	"fmt"
	"strings"
)

// UFWAction represents a firewall rule action.
type UFWAction string

const (
	UFWActionAllow UFWAction = "allow"
	UFWActionDeny  UFWAction = "deny"
)

// ParseUFWAction parses a string ("on" / "off" or "allow" / "deny") to UFWAction.
func ParseUFWAction(s string) (UFWAction, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "on", "allow", "enable":
		return UFWActionAllow, nil
	case "off", "deny", "disable":
		return UFWActionDeny, nil
	default:
		return "", fmt.Errorf("invalid ufw action: %q (expected 'on' or 'off')", s)
	}
}
