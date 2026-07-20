package password

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Blocklist interface {
	Contains(password string) bool
}

type blocklist struct {
	passwords map[string]struct{}
}

func NewBlocklist(passwords []string) Blocklist {
	m := make(map[string]struct{}, len(passwords))
	for _, p := range passwords {
		m[p] = struct{}{}
	}
	return &blocklist{passwords: m}
}

func NewBlocklistFromFile(path string) (Blocklist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open blocklist file: %w", err)
	}
	defer f.Close()

	m := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			m[line] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read blocklist file: %w", err)
	}

	return &blocklist{passwords: m}, nil
}

func (b *blocklist) Contains(password string) bool {
	_, ok := b.passwords[password]
	return ok
}

func NewBuiltinBlocklist() Blocklist {
	return NewBlocklist(builtinBlocklist)
}

var builtinBlocklist = []string{
	"password",
	"123456789012345",
	"qwertyuiop12345",
	"senha1234567890",
	"admin1234567890",
	"letmein12345678",
	"welcome12345678",
	"monkey1234567890",
	"dragon1234567890",
	"master1234567890",
	"summer2024123456",
	"winter2024123456",
	"spring2024123456",
	"fall202412345678",
}
