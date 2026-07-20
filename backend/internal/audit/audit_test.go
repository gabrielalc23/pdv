package audit

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestMetadata_New(t *testing.T) {
	m := NewMetadata()
	m.Set("reason", "test")
	m.Set("count", 42)

	if err := m.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty metadata")
	}
}

func TestMetadata_Nil(t *testing.T) {
	var m Metadata
	if err := m.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if string(data) != "{}" {
		t.Fatalf("expected {}, got %s", string(data))
	}
}

func TestMetadata_EmptyKey(t *testing.T) {
	m := NewMetadata()
	m.Set("", "value")

	if err := m.Validate(); err == nil {
		t.Fatal("expected validation error for empty key")
	}
}

func TestEventTypeConstants(t *testing.T) {
	events := []string{
		EventAuthRefreshed,
		EventAuthRefreshReused,
		EventAuthLoggedOut,
		EventAuthLoggedOutAll,
		EventSessionRevoked,
		EventAuthRegistered,
		EventAuthLoginSucceeded,
		EventAuthLoginFailed,
		EventAuthContextChanged,
		EventAuthPasswordChanged,
		EventAuthPasswordResetReq,
		EventAuthPasswordResetComp,
		EventAuthEmailVerified,
		EventOrganizationCreated,
		EventOrganizationUpdated,
		EventOrganizationArchived,
		EventStoreCreated,
		EventStoreUpdated,
		EventStoreStatusChanged,
		EventMembershipInvited,
		EventMembershipJoined,
		EventMembershipSuspended,
		EventMembershipReactivated,
		EventMembershipRemoved,
		EventRoleCreated,
		EventRoleUpdated,
		EventRoleStatusChanged,
		EventRoleBindingAdded,
		EventRoleBindingRemoved,
	}

	for _, e := range events {
		if e == "" {
			t.Fatal("event type constant must not be empty")
		}
	}
}

func TestWriter_RequiresEventType(t *testing.T) {
	w := NewWriter()
	err := w.Write(nil, nil, Event{
		EventType: "",
	})
	if err == nil {
		t.Fatal("expected error for empty event type")
	}
}

func TestMetadataMarshalRoundtrip(t *testing.T) {
	m := NewMetadata()
	m.Set("reason", "user_logged_out")
	m.Set("affected_session_count", 3)

	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty marshaled data")
	}
}

func mustUUID(s string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		panic(err)
	}
	return id
}
