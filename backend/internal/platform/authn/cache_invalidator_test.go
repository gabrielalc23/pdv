package authn

import (
	"bufio"
	"context"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	platformvalkey "github.com/gabrielalc23/pdv/internal/platform/valkey"
)

func TestCacheInvalidatorInvalidatesAuthorizationVersionKeys(t *testing.T) {
	server := startCacheInvalidatorValkey(t)
	client, err := platformvalkey.NewClient(platformvalkey.Config{Addr: server.address()})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(client.Close)

	invalidator := NewCacheInvalidator(NewSessionCache(client, time.Minute))
	organizationID := repeatedUUID(0x11)
	membershipID := repeatedUUID(0x22)

	invalidator.InvalidateOrganizationAuthorizationVersion(context.Background(), organizationID)
	invalidator.InvalidateMembershipAuthorizationVersion(context.Background(), membershipID)

	want := [][]string{
		{"DEL", "auth:org-version:11111111-1111-1111-1111-111111111111"},
		{"DEL", "auth:membership-version:22222222-2222-2222-2222-222222222222"},
	}
	if got := server.operations(); !reflect.DeepEqual(got, want) {
		t.Fatalf("operations = %#v, want %#v", got, want)
	}
}

func TestCacheInvalidatorAuthorizationVersionMethodsAreNilSafe(t *testing.T) {
	var invalidator *CacheInvalidator
	invalidator.InvalidateOrganizationAuthorizationVersion(context.Background(), pgtype.UUID{})
	invalidator.InvalidateMembershipAuthorizationVersion(context.Background(), pgtype.UUID{})

	invalidator = NewCacheInvalidator(nil)
	invalidator.InvalidateOrganizationAuthorizationVersion(context.Background(), pgtype.UUID{})
	invalidator.InvalidateMembershipAuthorizationVersion(context.Background(), pgtype.UUID{})
}

func repeatedUUID(value byte) pgtype.UUID {
	var bytes [16]byte
	for index := range bytes {
		bytes[index] = value
	}
	return pgtype.UUID{Bytes: bytes, Valid: true}
}

type cacheInvalidatorValkey struct {
	listener net.Listener
	mu       sync.Mutex
	commands [][]string
}

func startCacheInvalidatorValkey(t *testing.T) *cacheInvalidatorValkey {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &cacheInvalidatorValkey{listener: listener}
	t.Cleanup(func() { _ = listener.Close() })
	go server.serve()
	return server
}

func (s *cacheInvalidatorValkey) address() string {
	return s.listener.Addr().String()
}

func (s *cacheInvalidatorValkey) operations() [][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	operations := make([][]string, len(s.commands))
	for index := range s.commands {
		operations[index] = append([]string(nil), s.commands[index]...)
	}
	return operations
}

func (s *cacheInvalidatorValkey) serve() {
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(connection)
	}
}

func (s *cacheInvalidatorValkey) handle(connection net.Conn) {
	defer connection.Close()
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)
	for {
		command, err := readRESPCommand(reader)
		if err != nil {
			return
		}
		switch strings.ToUpper(command[0]) {
		case "HELLO":
			_, _ = writer.WriteString("%2\r\n+proto\r\n:3\r\n+version\r\n+7.2.0\r\n")
		case "CLIENT":
			_, _ = writer.WriteString("+OK\r\n")
		case "CLUSTER":
			_, _ = writer.WriteString("-ERR This instance has cluster support disabled\r\n")
		case "DEL":
			s.mu.Lock()
			s.commands = append(s.commands, append([]string(nil), command...))
			s.mu.Unlock()
			_, _ = writer.WriteString(":1\r\n")
		default:
			s.mu.Lock()
			s.commands = append(s.commands, append([]string(nil), command...))
			s.mu.Unlock()
			_, _ = writer.WriteString("+OK\r\n")
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}

func readRESPCommand(reader *bufio.Reader) ([]string, error) {
	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(header, "*")))
	if err != nil {
		return nil, err
	}
	command := make([]string, count)
	for index := range command {
		bulkHeader, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(bulkHeader, "$")))
		if err != nil {
			return nil, err
		}
		value := make([]byte, length+2)
		if _, err := io.ReadFull(reader, value); err != nil {
			return nil, err
		}
		command[index] = string(value[:length])
	}
	return command, nil
}
