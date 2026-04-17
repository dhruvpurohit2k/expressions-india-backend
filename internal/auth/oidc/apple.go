package oidc

import (
	"errors"
	"os"
	"strings"
	"sync"
)

const (
	appleJWKSURL = "https://appleid.apple.com/auth/keys"
	appleIssuer  = "https://appleid.apple.com"
)

var (
	appleOnce     sync.Once
	appleVerifier *Verifier
	appleInitErr  error
)

// AppleVerifier returns a process-wide verifier configured from
// APPLE_OIDC_AUDIENCES (comma-separated list of bundle IDs / Service IDs —
// typically the iOS app's bundle ID, plus the Service ID if you offer
// Sign in with Apple on the web).
func AppleVerifier() (*Verifier, error) {
	appleOnce.Do(func() {
		raw := os.Getenv("APPLE_OIDC_AUDIENCES")
		if raw == "" {
			appleInitErr = errors.New("APPLE_OIDC_AUDIENCES not set")
			return
		}
		var auds []string
		for _, a := range strings.Split(raw, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				auds = append(auds, a)
			}
		}
		if len(auds) == 0 {
			appleInitErr = errors.New("APPLE_OIDC_AUDIENCES has no valid entries")
			return
		}
		appleVerifier = newVerifier(appleJWKSURL, appleIssuer, auds)
	})
	return appleVerifier, appleInitErr
}
