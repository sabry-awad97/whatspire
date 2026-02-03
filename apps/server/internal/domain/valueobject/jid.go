package valueobject

import "strings"

// JID represents a WhatsApp Jabber ID
type JID string

// CleanJID removes the device ID from a WhatsApp JID while keeping the domain
// Example: "201021347532:98@s.whatsapp.net" -> "201021347532@s.whatsapp.net"
func CleanJID(jid string) string {
	if jid == "" {
		return ""
	}

	// Split by @ to separate user part and domain
	parts := strings.Split(jid, "@")
	if len(parts) != 2 {
		return jid // Return as-is if format is unexpected
	}

	userPart := parts[0]
	domain := parts[1]

	// Remove device ID (everything after :)
	if idx := strings.Index(userPart, ":"); idx != -1 {
		userPart = userPart[:idx]
	}

	return userPart + "@" + domain
}

// ExtractPhone extracts the phone number from a WhatsApp JID
// Example: "201021347532:98@s.whatsapp.net" -> "201021347532"
func ExtractPhone(jid string) string {
	if jid == "" {
		return ""
	}

	// Split by @ to get user part
	userPart := strings.Split(jid, "@")[0]

	// Remove device ID if present
	if idx := strings.Index(userPart, ":"); idx != -1 {
		return userPart[:idx]
	}

	return userPart
}

// IsGroupJID checks if the JID is a group JID
func IsGroupJID(jid string) bool {
	return strings.Contains(jid, "@g.us")
}

// IsBroadcastJID checks if the JID is a broadcast list JID
func IsBroadcastJID(jid string) bool {
	return strings.Contains(jid, "@broadcast")
}

// IsValidJID validates if the JID has a valid format
func IsValidJID(jid string) bool {
	if jid == "" {
		return false
	}

	// Must contain @
	if !strings.Contains(jid, "@") {
		return false
	}

	// Check for valid domains
	validDomains := []string{"s.whatsapp.net", "g.us", "broadcast"}
	for _, domain := range validDomains {
		if strings.Contains(jid, "@"+domain) {
			return true
		}
	}

	return false
}
