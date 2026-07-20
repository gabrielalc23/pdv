package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func writeAuditEvent(ctx context.Context, q Querier, eventType string, outcome database.AuditOutcome, metadata map[string]any, reqMeta requestmeta.RequestMetadata) error {
	return writeAuditEventWithSubject(ctx, q, eventType, outcome, metadata, reqMeta, auditSubject{})
}

type auditSubject struct {
	UserID         pgtype.UUID
	MembershipID   pgtype.UUID
	OrganizationID pgtype.UUID
	StoreID        pgtype.UUID
	SessionID      pgtype.UUID
}

func writeAuditEventWithSubject(ctx context.Context, q Querier, eventType string, outcome database.AuditOutcome, metadata map[string]any, reqMeta requestmeta.RequestMetadata, subject auditSubject) error {
	metaBytes, err := marshalMetadata(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	var ipAddr *netip.Addr
	if reqMeta.ClientIP != "" {
		if addr, err := netip.ParseAddr(reqMeta.ClientIP); err == nil {
			ipAddr = &addr
		}
	}

	userAgent := pgtype.Text{Valid: false}
	if reqMeta.UserAgent != "" {
		userAgent = pgtype.Text{String: reqMeta.UserAgent, Valid: true}
	}

	requestID := pgtype.Text{Valid: false}
	if reqMeta.RequestID != "" {
		requestID = pgtype.Text{String: reqMeta.RequestID, Valid: true}
	}

	_, err = q.CreateAuditEvent(ctx, database.CreateAuditEventParams{
		OrganizationID:    subject.OrganizationID,
		StoreID:           subject.StoreID,
		ActorUserID:       subject.UserID,
		ActorMembershipID: subject.MembershipID,
		SessionID:         subject.SessionID,
		EventType:         eventType,
		Outcome:           outcome,
		Metadata:          metaBytes,
		IpAddress:         ipAddr,
		UserAgent:         userAgent,
		RequestID:         requestID,
	})
	if err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}

func (s *Service) CreateSession(ctx context.Context, input CreateSessionInput) (CreateSessionResult, error) {
	var result CreateSessionResult
	err := s.provider.WithTx(ctx, func(q Querier) error {
		var err error
		result, err = s.CreateSessionInTx(ctx, q, input)
		return err
	})
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("create session transaction: %w", err)
	}
	return result, nil
}

// CreateSessionInTx lets a higher-level authentication flow include session and
// refresh creation in the same transaction as its other security writes.
func (s *Service) CreateSessionInTx(ctx context.Context, q Querier, input CreateSessionInput) (CreateSessionResult, error) {
	if err := validateClientID(input.ClientID); err != nil {
		return CreateSessionResult{}, fmt.Errorf("validate session client: %w", err)
	}
	if err := validateContextCoherence(input.ContextKind, input.OrganizationID, input.MembershipID, input.StoreID); err != nil {
		return CreateSessionResult{}, fmt.Errorf("validate session context: %w", err)
	}

	now := s.clock.Now()
	absoluteExpiresAt := now.Add(s.cfg.SessionAbsoluteTTL)
	idleExpiresAt := now.Add(s.cfg.RefreshIdleTTL)
	if idleExpiresAt.After(absoluteExpiresAt) {
		idleExpiresAt = absoluteExpiresAt
	}

	row, err := q.CreateAuthSession(ctx, database.CreateAuthSessionParams{
		UserID:                input.UserID,
		ClientID:              input.ClientID,
		DeviceName:            input.DeviceName,
		UserAgent:             input.UserAgent,
		IpAddress:             input.IPAddress,
		ContextKind:           database.AuthContextKind(input.ContextKind),
		CurrentOrganizationID: input.OrganizationID,
		CurrentMembershipID:   input.MembershipID,
		CurrentStoreID:        input.StoreID,
		IdleExpiresAt:         pgtype.Timestamptz{Time: idleExpiresAt, Valid: true},
		AbsoluteExpiresAt:     pgtype.Timestamptz{Time: absoluteExpiresAt, Valid: true},
	})
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("create auth session: %w", err)
	}

	tokenID, err := newRandomUUID()
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("create refresh token selector: %w", err)
	}
	rawToken, secretHash, err := s.codec.Generate(tokenID)
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("generate initial refresh token: %w", err)
	}

	refreshToken, err := q.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		ID:         tokenID,
		SessionID:  row.ID,
		SecretHash: secretHash,
		ExpiresAt:  pgtype.Timestamptz{Time: absoluteExpiresAt, Valid: true},
	})
	if err != nil {
		return CreateSessionResult{}, fmt.Errorf("create initial refresh token: %w", err)
	}
	if refreshToken.ID != tokenID {
		return CreateSessionResult{}, fmt.Errorf("create initial refresh token: persisted selector does not match generated selector")
	}

	return CreateSessionResult{Session: sessionFromDBRow(row), RawRefreshToken: rawToken}, nil
}

func marshalMetadata(meta map[string]any) ([]byte, error) {
	if meta == nil {
		return []byte("{}"), nil
	}
	jsonMeta := make(map[string]any, len(meta))
	for k, v := range meta {
		jsonMeta[k] = v
	}
	data, err := json.Marshal(jsonMeta)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}
	return data, nil
}
