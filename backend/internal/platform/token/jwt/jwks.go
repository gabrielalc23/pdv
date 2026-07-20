package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type JWKSKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	X   string `json:"x"`
}

type JWKS struct {
	Keys []JWKSKey `json:"keys"`
}

type JWKSService struct {
	keyring *Keyring
	mu      sync.RWMutex
	cached  []byte
	etag    string
	lastMod time.Time
}

func NewJWKSService(keyring *Keyring) *JWKSService {
	s := &JWKSService{keyring: keyring}
	s.generate()
	return s
}

func (s *JWKSService) generate() {
	kids := s.keyring.SortedKIDs()
	keys := make([]JWKSKey, 0, len(kids))
	for _, kid := range kids {
		pubKey := s.keyring.PublicKeys[kid]
		keys = append(keys, JWKSKey{
			Kty: "OKP",
			Crv: "Ed25519",
			Use: "sig",
			Alg: "EdDSA",
			Kid: kid,
			X:   base64.RawURLEncoding.EncodeToString([]byte(pubKey)),
		})
	}

	jwks := JWKS{Keys: keys}
	data, err := json.Marshal(jwks)
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = data
	prefix := len(data)
	if prefix > 16 {
		prefix = 16
	}
	s.etag = fmt.Sprintf(`"%x"`, data[:prefix])
	s.lastMod = time.Now()
}

func (s *JWKSService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cached := s.cached
	etag := s.etag
	s.mu.RUnlock()

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(cached)
}
