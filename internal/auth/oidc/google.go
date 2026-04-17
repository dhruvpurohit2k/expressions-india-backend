package oidc

import (
	"errors"
	"os"
	"strings"
	"sync"
)

const (
	googleJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
	googleIssuer  = "https://accounts.google.com"
)

var (
	googleOnce     sync.Once
	googleVerifier *Verifier
	googleInitErr  error
)

// GoogleVerifier returns a process-wide verifier configured from
// GOOGLE_OIDC_AUDIENCES (comma-separated list of OAuth client IDs — typically
// one per platform: web, iOS, Android).
func GoogleVerifier() (*Verifier, error) {
	googleOnce.Do(func() {
		raw := os.Getenv("GOOGLE_OIDC_AUDIENCES")
		if raw == "" {
			googleInitErr = errors.New("GOOGLE_OIDC_AUDIENCES not set")
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
			googleInitErr = errors.New("GOOGLE_OIDC_AUDIENCES has no valid entries")
			return
		}
		googleVerifier = newVerifier(googleJWKSURL, googleIssuer, auds)
	})
	return googleVerifier, googleInitErr
}
