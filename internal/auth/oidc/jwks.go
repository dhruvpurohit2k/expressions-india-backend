// Package oidc verifies ID tokens issued by external identity providers
// (Google, Apple). It fetches and caches the provider's JSON Web Key Set
// (JWKS) and validates token signature, issuer, audience, and expiry.
package oidc

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwksRefreshInterval = 1 * time.Hour
	jwksHTTPTimeout     = 5 * time.Second
)

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksDoc struct {
	Keys []jwk `json:"keys"`
}

// jwksCache fetches and caches RSA public keys from a JWKS URL.
type jwksCache struct {
	url     string
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
	httpc   *http.Client
}

func newJWKSCache(url string) *jwksCache {
	return &jwksCache{
		url:   url,
		keys:  map[string]*rsa.PublicKey{},
		httpc: &http.Client{Timeout: jwksHTTPTimeout},
	}
}

func (c *jwksCache) get(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	stale := time.Since(c.fetched) > jwksRefreshInterval
	c.mu.RUnlock()
	if ok && !stale {
		return key, nil
	}
	if err := c.refresh(); err != nil {
		// On refresh failure, fall back to cached key if we have one.
		if ok {
			return key, nil
		}
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if k, ok := c.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("oidc: no key matches kid %q", kid)
}

func (c *jwksCache) refresh() error {
	resp, err := c.httpc.Get(c.url)
	if err != nil {
		return fmt.Errorf("oidc: fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oidc: jwks status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("oidc: read jwks: %w", err)
	}
	var doc jwksDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("oidc: parse jwks: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := rsaKey(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return errors.New("oidc: jwks contained no usable RSA keys")
	}
	c.mu.Lock()
	c.keys = keys
	c.fetched = time.Now()
	c.mu.Unlock()
	return nil
}

func rsaKey(nB64u, eB64u string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64u)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64u)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("oidc: invalid RSA exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

// IDClaims is the minimal subset of OIDC ID token claims we care about.
// Apple omits email/name on subsequent sign-ins; treat them as optional.
type IDClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"` // Google: bool, Apple: string "true"/"false"
	Name          string `json:"name"`
	jwt.RegisteredClaims
}

// EmailIsVerified normalizes the inconsistent provider encoding.
func (c *IDClaims) EmailIsVerified() bool {
	switch v := c.EmailVerified.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	}
	return false
}

// Verifier validates ID tokens signed by a single OIDC provider.
type Verifier struct {
	jwks      *jwksCache
	issuer    string
	audiences []string
}

func newVerifier(jwksURL, issuer string, audiences []string) *Verifier {
	return &Verifier{
		jwks:      newJWKSCache(jwksURL),
		issuer:    issuer,
		audiences: audiences,
	}
}

// Verify parses and validates an ID token. Returns claims on success.
func (v *Verifier) Verify(idToken string) (*IDClaims, error) {
	claims := &IDClaims{}
	tok, err := jwt.ParseWithClaims(idToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("oidc: unexpected signing method %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("oidc: missing kid header")
		}
		return v.jwks.get(kid)
	},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("oidc: token invalid: %w", err)
	}
	if !tok.Valid {
		return nil, errors.New("oidc: token invalid")
	}
	if !audMatches(claims.Audience, v.audiences) {
		return nil, errors.New("oidc: audience mismatch")
	}
	if claims.Sub == "" {
		return nil, errors.New("oidc: missing sub")
	}
	return claims, nil
}

func audMatches(tokenAud jwt.ClaimStrings, allowed []string) bool {
	for _, a := range tokenAud {
		for _, want := range allowed {
			if a == want {
				return true
			}
		}
	}
	return false
}
