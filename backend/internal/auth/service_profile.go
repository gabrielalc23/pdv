package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	usersmodule "github.com/gabrielalc23/pdv/internal/users"
)

func (s *Service) UpdateMe(ctx context.Context, userID pgtype.UUID, input UpdateMeRequest) (UserResponse, error) {
	displayName, err := usersmodule.NormalizeDisplayName(input.DisplayName)
	if err != nil {
		if errors.Is(err, usersmodule.ErrInvalidDisplayName) {
			return UserResponse{}, validationError("displayName", "Nome de exibição inválido.")
		}
		return UserResponse{}, err
	}
	row, err := s.store.Queries.UpdateUserDisplayName(ctx, database.UpdateUserDisplayNameParams{DisplayName: displayName, UserID: userID})
	if err != nil {
		return UserResponse{}, fmt.Errorf("%w: update display name: %w", ErrDependencyUnavailable, err)
	}
	return UserResponse{ID: uuidString(row.ID), Email: row.Email, DisplayName: row.DisplayName, EmailVerified: row.EmailVerifiedAt.Valid}, nil
}
