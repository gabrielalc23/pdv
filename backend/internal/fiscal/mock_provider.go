package fiscal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type MockProvider struct {
	Fail      bool
	FailError error
	Now       func() time.Time
}

func (p *MockProvider) Authorize(_ context.Context, input AuthorizationInput) (AuthorizationResult, error) {
	if p.Fail {
		if p.FailError != nil {
			return AuthorizationResult{}, p.FailError
		}
		return AuthorizationResult{}, fmt.Errorf("mock fiscal authorization failed")
	}

	now := time.Now().UTC()
	if p.Now != nil {
		now = p.Now().UTC()
	}

	sum := sha256.Sum256([]byte(input.SaleID + "|" + input.SaleTotal + "|" + fmt.Sprint(input.SaleNumber)))
	accessKey := hex.EncodeToString(sum[:])
	if len(accessKey) > 44 {
		accessKey = accessKey[:44]
	}

	return AuthorizationResult{
		Provider:          "mock",
		AccessKey:         accessKey,
		Protocol:          "MOCK-" + fmt.Sprint(input.SaleNumber),
		XML:               "<fiscal sale=\"" + input.SaleID + "\" total=\"" + input.SaleTotal + "\" />",
		ExternalReference: "sale-" + input.SaleID,
		AuthorizedAt:      now,
	}, nil
}
