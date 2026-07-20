package invitations

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func TestServiceCreateRejectsInvalidInputBeforeTransaction(t *testing.T) {
	provider := &countingTxProvider{}
	service := &Service{tx: provider}
	actor := authcontext.Principal{UserID: testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216456"), SessionID: testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216457"), OrganizationID: testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216458"), MembershipID: testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216459"), Scopes: authcontext.ScopeSet{authz.ScopeMembersInvite: {}}}

	_, err := service.Create(context.Background(), actor, CreateInput{Email: "not-an-email"}, structRequestMeta())
	if err == nil {
		t.Fatal("expected validation error")
	}
	if provider.calls != 0 {
		t.Fatalf("transaction calls = %d, want 0", provider.calls)
	}
}

func TestPendingInvitationConstraintIsTranslated(t *testing.T) {
	err := mapPersistenceError(&pgconn.PgError{Code: "23505", ConstraintName: "idx_organization_invitations_pending_email"})
	if !errors.Is(err, ErrInvitationPending) {
		t.Fatalf("error = %v", err)
	}
}

func TestServicePublicTokenFailuresAreNonEnumerating(t *testing.T) {
	codec, err := newTokenCodec(make([]byte, 32), nilReader{})
	if err != nil {
		t.Fatal(err)
	}
	service := &Service{codec: codec, store: &countingStore{}}
	if _, err := service.Inspect(context.Background(), "invalid"); !errors.Is(err, ErrInvitationNotFound) {
		t.Fatalf("Inspect error = %v", err)
	}
	if _, err := service.Accept(context.Background(), nil, AcceptInput{Token: "invalid"}, structRequestMeta()); !errors.Is(err, ErrInvitationNotFound) {
		t.Fatalf("Accept error = %v", err)
	}
	if service.store.(*countingStore).calls != 0 {
		t.Fatal("malformed token reached persistence")
	}
}

type nilReader struct{}

func (nilReader) Read([]byte) (int, error) { return 0, errors.New("unused") }

type countingTxProvider struct{ calls int }

func (p *countingTxProvider) WithTx(context.Context, func(TxStore) error) error {
	p.calls++
	return errors.New("unexpected transaction")
}

type countingStore struct{ calls int }

func (s *countingStore) called() { s.calls++ }
func (s *countingStore) ListInvitations(context.Context, database.ListInvitationsParams) ([]database.ListInvitationsRow, error) {
	s.called()
	return nil, errors.New("unexpected")
}
func (s *countingStore) CountInvitations(context.Context, database.CountInvitationsParams) (int64, error) {
	s.called()
	return 0, errors.New("unexpected")
}
func (s *countingStore) GetInvitation(context.Context, pgtype.UUID) (database.OrganizationInvitation, error) {
	s.called()
	return database.OrganizationInvitation{}, errors.New("unexpected")
}
func (s *countingStore) ListInvitationRoleBindings(context.Context, pgtype.UUID, pgtype.UUID) ([]database.ListInvitationRoleBindingsRow, error) {
	s.called()
	return nil, errors.New("unexpected")
}
func (s *countingStore) GetOrganization(context.Context, pgtype.UUID) (database.Organization, error) {
	s.called()
	return database.Organization{}, errors.New("unexpected")
}
func (s *countingStore) GetRole(context.Context, pgtype.UUID, pgtype.UUID) (database.Role, error) {
	s.called()
	return database.Role{}, errors.New("unexpected")
}
func (s *countingStore) ListScopeCodesForRole(context.Context, pgtype.UUID, pgtype.UUID) ([]string, error) {
	s.called()
	return nil, errors.New("unexpected")
}
func (s *countingStore) GetUserByNormalizedEmail(context.Context, string) (database.User, error) {
	s.called()
	return database.User{}, errors.New("unexpected")
}
func (s *countingStore) GetUserByID(context.Context, pgtype.UUID) (database.User, error) {
	s.called()
	return database.User{}, errors.New("unexpected")
}

func structRequestMeta() requestmeta.RequestMetadata { return requestmeta.RequestMetadata{} }
