package jwt

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Keyring struct {
	ActiveKID  string
	PrivateKey ed25519.PrivateKey
	PublicKeys map[string]ed25519.PublicKey
}

func LoadKeyring(activeKID, privateKeyPath, publicKeysDir string) (*Keyring, error) {
	if activeKID == "" {
		return nil, fmt.Errorf("%w: active key id is required", ErrKeyInvalid)
	}
	if privateKeyPath == "" {
		return nil, fmt.Errorf("%w: private key path is required", ErrKeyInvalid)
	}

	privPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key %q: %w", privateKeyPath, err)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("%w: invalid PEM block in private key", ErrKeyInvalid)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse private key: %w", ErrKeyInvalid, err)
	}

	edPriv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: private key is not Ed25519", ErrKeyInvalid)
	}

	if !edPriv.Public().(ed25519.PublicKey).Equal(edPriv.Public()) {
		return nil, fmt.Errorf("%w: corrupted key pair", ErrKeyInvalid)
	}

	publicKeys := make(map[string]ed25519.PublicKey)
	activePub := edPriv.Public().(ed25519.PublicKey)

	if publicKeysDir != "" {
		entries, err := os.ReadDir(publicKeysDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read public keys dir %q: %w", publicKeysDir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) != ".pem" {
				continue
			}
			if strings.HasSuffix(entry.Name(), ".priv.pem") {
				continue
			}

			kid := strings.TrimSuffix(entry.Name(), ".pem")
			pubPath := filepath.Join(publicKeysDir, entry.Name())
			pubKey, err := loadPublicKeyPEM(pubPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load public key %q: %w", entry.Name(), err)
			}

			if _, exists := publicKeys[kid]; exists {
				return nil, fmt.Errorf("%w: duplicate key id %q", ErrKeyDuplicate, kid)
			}
			publicKeys[kid] = pubKey
		}
	}

	if _, exists := publicKeys[activeKID]; !exists {
		publicKeys[activeKID] = activePub
	}

	if stored, exists := publicKeys[activeKID]; exists {
		if !stored.Equal(activePub) {
			return nil, fmt.Errorf("%w: active key %q public key mismatch", ErrKeyInvalid, activeKID)
		}
	}

	publicKeys[activeKID] = activePub

	return &Keyring{
		ActiveKID:  activeKID,
		PrivateKey: edPriv,
		PublicKeys: publicKeys,
	}, nil
}

func NewEphemeralKeyring(activeKID string) (*Keyring, error) {
	if activeKID == "" {
		activeKID = "dev-key"
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	return &Keyring{
		ActiveKID:  activeKID,
		PrivateKey: priv,
		PublicKeys: map[string]ed25519.PublicKey{activeKID: pub},
	}, nil
}

func (k *Keyring) PublicKey(kid string) (ed25519.PublicKey, bool) {
	key, ok := k.PublicKeys[kid]
	return key, ok
}

func (k *Keyring) SortedKIDs() []string {
	kids := make([]string, 0, len(k.PublicKeys))
	for kid := range k.PublicKeys {
		kids = append(kids, kid)
	}
	sort.Strings(kids)
	return kids
}

func (k *Keyring) PublicKeyBytes(kid string) ([]byte, error) {
	key, ok := k.PublicKeys[kid]
	if !ok {
		return nil, fmt.Errorf("%w: key %q not found", ErrKeyNotFound, kid)
	}
	return key, nil
}

func loadPublicKeyPEM(path string) (ed25519.PublicKey, error) {
	pubPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pubPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block in %q", path)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	edPub, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key in %q is not Ed25519", path)
	}

	return edPub, nil
}
