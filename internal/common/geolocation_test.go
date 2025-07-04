package common

import (
	"testing"
)

func TestGetLocationFromIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "IP locale",
			ip:       "127.0.0.1",
			expected: "Local",
		},
		{
			name:     "localhost",
			ip:       "localhost",
			expected: "Local",
		},
		{
			name:     "IPv6 local",
			ip:       "::1",
			expected: "Local",
		},
		{
			name:     "IP Google DNS",
			ip:       "8.8.8.8",
			expected: "8.8.8.8", // devrait retourner une localisation réelle ou l'IP en fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLocationFromIP(tt.ip)

			// Pour les IPs locales, on vérifie que c'est "Local"
			if tt.expected == "Local" {
				if result != "Local" {
					t.Errorf("GetLocationFromIP() = %v, want %v", result, tt.expected)
				}
			} else {
				// Pour les autres IPs, on vérifie juste que ça ne retourne pas une chaîne vide
				if result == "" {
					t.Errorf("GetLocationFromIP() returned empty string for IP %v", tt.ip)
				}
			}
		})
	}
}
