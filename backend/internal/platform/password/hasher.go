package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrHashInvalid        = errors.New("password hash is malformed")
	ErrHashInvalidAlgo    = errors.New("password hash uses unsupported algorithm")
	ErrHashInvalidVersion = errors.New("password hash uses unsupported algorithm version")
	ErrHashInvalidParams  = errors.New("password hash has invalid parameters")
	ErrHashExcessive      = errors.New("password hash has excessive parameters")
)

type Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint32
	SaltLength  uint32
	KeyLength   uint32
}

func (p Params) Validate() error {
	if p.MemoryKiB < 8 || p.MemoryKiB > 1<<24 {
		return errors.New("argon2 memory must be between 8 KiB and 16 GiB")
	}
	if p.Iterations < 1 || p.Iterations > 100 {
		return errors.New("argon2 iterations must be between 1 and 100")
	}
	if p.Parallelism < 1 || p.Parallelism > 255 {
		return errors.New("argon2 parallelism must be between 1 and 255")
	}
	if p.SaltLength < 8 || p.SaltLength > 64 {
		return errors.New("argon2 salt length must be between 8 and 64 bytes")
	}
	if p.KeyLength < 16 || p.KeyLength > 64 {
		return errors.New("argon2 key length must be between 16 and 64 bytes")
	}
	return nil
}

type Hasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (match, needsRehash bool, err error)
}

type hasher struct {
	params Params
	rand   io.Reader
}

func NewHasher(params Params) (Hasher, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &hasher{params: params, rand: rand.Reader}, nil
}

func NewHasherWithReader(params Params, r io.Reader) (Hasher, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &hasher{params: params, rand: r}, nil
}

func DefaultParams() Params {
	return Params{
		MemoryKiB:   65536,
		Iterations:  3,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func (h *hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := io.ReadFull(h.rand, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.MemoryKiB,
		uint8(h.params.Parallelism),
		h.params.KeyLength,
	)

	return encodePHC(h.params, salt, hash), nil
}

func (h *hasher) Verify(password, encodedHash string) (match, needsRehash bool, err error) {
	params, salt, hash, err := decodePHC(encodedHash)
	if err != nil {
		return false, false, err
	}

	computed := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.MemoryKiB,
		uint8(params.Parallelism),
		params.KeyLength,
	)

	if subtle.ConstantTimeCompare(hash, computed) != 1 {
		return false, false, nil
	}

	needsRehash = params.MemoryKiB < h.params.MemoryKiB ||
		params.Iterations < h.params.Iterations ||
		params.KeyLength < h.params.KeyLength

	return true, needsRehash, nil
}

const argon2Version = 19
const MaxPHCLength = 256
const maxArgon2MemoryKiB uint32 = 1 << 24
const maxArgon2Iterations uint32 = 100
const maxArgon2Parallelism uint32 = 255

func encodePHC(params Params, salt, hash []byte) string {
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2Version, params.MemoryKiB, params.Iterations, params.Parallelism, saltB64, hashB64)
}

func decodePHC(encoded string) (Params, []byte, []byte, error) {
	if len(encoded) > MaxPHCLength {
		return Params{}, nil, nil, fmt.Errorf("%w: hash exceeds %d bytes", ErrHashExcessive, MaxPHCLength)
	}
	if !strings.HasPrefix(encoded, "$") {
		return Params{}, nil, nil, fmt.Errorf("%w: missing leading $", ErrHashInvalid)
	}

	parts := strings.Split(encoded[1:], "$")
	if len(parts) != 5 {
		return Params{}, nil, nil, fmt.Errorf("%w: expected 5 segments, got %d", ErrHashInvalid, len(parts))
	}

	if parts[0] != "argon2id" {
		return Params{}, nil, nil, fmt.Errorf("%w: expected argon2id, got %q", ErrHashInvalidAlgo, parts[0])
	}

	if !strings.HasPrefix(parts[1], "v=") {
		return Params{}, nil, nil, fmt.Errorf("%w: missing version prefix", ErrHashInvalid)
	}
	version, err := strconv.Atoi(parts[1][2:])
	if err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: invalid version: %w", ErrHashInvalidParams, err)
	}
	if version != argon2Version {
		return Params{}, nil, nil, fmt.Errorf("%w: got v=%d, expected v=%d", ErrHashInvalidVersion, version, argon2Version)
	}

	params, err := parseArgon2Params(parts[2])
	if err != nil {
		return Params{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: invalid salt encoding: %w", ErrHashInvalid, err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, fmt.Errorf("%w: invalid hash encoding: %w", ErrHashInvalid, err)
	}

	if len(hash) == 0 {
		return Params{}, nil, nil, fmt.Errorf("%w: empty hash", ErrHashInvalid)
	}

	params.SaltLength = uint32(len(salt))
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

func parseArgon2Params(segment string) (Params, error) {
	var p Params
	for _, part := range strings.Split(segment, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return Params{}, fmt.Errorf("%w: malformed param %q", ErrHashInvalidParams, part)
		}
		val, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return Params{}, fmt.Errorf("%w: invalid param value %q: %w", ErrHashInvalidParams, part, err)
		}
		switch kv[0] {
		case "m":
			p.MemoryKiB = uint32(val)
		case "t":
			p.Iterations = uint32(val)
		case "p":
			p.Parallelism = uint32(val)
		default:
			return Params{}, fmt.Errorf("%w: unknown parameter %q", ErrHashInvalidParams, kv[0])
		}
	}
	if p.MemoryKiB == 0 || p.Iterations == 0 || p.Parallelism == 0 {
		return Params{}, fmt.Errorf("%w: missing or zero parameter", ErrHashInvalidParams)
	}
	if p.MemoryKiB > maxArgon2MemoryKiB || p.Iterations > maxArgon2Iterations || p.Parallelism > maxArgon2Parallelism {
		return Params{}, fmt.Errorf("%w: parameter exceeds maximum", ErrHashExcessive)
	}
	return p, nil
}

var DummyHashValue = "$argon2id$v=19$m=65536,t=3,p=1$ZG1zYWx0ZG1zYWx0ZG1zYWx0ZG1zYWx0$ME2iNpx3q2S0OgXMwHjq6iGz3zXH3nKJMqK0GtCFUyI"

func IsDummyHash(h string) bool {
	return subtle.ConstantTimeCompare([]byte(h), []byte(DummyHashValue)) == 1
}
