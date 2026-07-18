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
		EventType: eventType,
		Outcome:   outcome,
		Metadata:  metaBytes,
		IpAddress: ipAddr,
		UserAgent: userAgent,
		RequestID: requestID,
	})
	return err
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
