package utils

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// ParseJWTClaims décode les claims d'un JWT (sans validation de signature)
func ParseJWTClaims(token string) map[string]interface{} {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return map[string]interface{}{}
	}

	payload, err := decodeSegment(parts[1])
	if err != nil {
		return map[string]interface{}{}
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return map[string]interface{}{}
	}
	return claims
}

// decodeSegment décode un segment base64url sans padding
func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}
