package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestUpdateMeMapsInvalidDisplayNamesBeforePersistence(t *testing.T) {
	t.Parallel()

	service := &Service{}
	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "blank", value: " \t "},
		{name: "control character", value: "Valid\u0000Name"},
		{name: "over rune limit", value: strings.Repeat("界", 151)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := service.UpdateMe(context.Background(), pgtype.UUID{}, UpdateMeRequest{DisplayName: test.value})
			var validation *ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("UpdateMe() error = %#v, want ValidationError", err)
			}
			if validation.Field != "displayName" || validation.Message != "Nome de exibição inválido." {
				t.Fatalf("UpdateMe() validation = %+v", validation)
			}
		})
	}
}
