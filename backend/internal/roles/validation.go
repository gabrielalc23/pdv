package roles

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

var roleKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*$`)

type normalizedRoleInput struct {
	key             string
	name            string
	description     pgtype.Text
	assignmentScope database.RoleAssignmentScope
	scopes          []string
}

func normalizeRoleInput(input UpsertRoleInput) (normalizedRoleInput, error) {
	key := strings.ToLower(strings.TrimSpace(input.Key))
	if key == "" {
		return normalizedRoleInput{}, validationError("key", "is required")
	}
	if len(key) > 80 || !roleKeyPattern.MatchString(key) {
		return normalizedRoleInput{}, validationError("key", "must be at most 80 characters in lowercase snake_case")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return normalizedRoleInput{}, validationError("name", "is required")
	}
	if len(name) > 120 {
		return normalizedRoleInput{}, validationError("name", "must be at most 120 characters")
	}

	var description pgtype.Text
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		if len(value) > 255 {
			return normalizedRoleInput{}, validationError("description", "must be at most 255 characters")
		}
		if value != "" {
			description = pgtype.Text{String: value, Valid: true}
		}
	}

	assignmentScope := database.RoleAssignmentScope(strings.ToUpper(strings.TrimSpace(input.AssignmentScope)))
	if !assignmentScope.Valid() {
		return normalizedRoleInput{}, validationError("assignmentScope", "must be ORGANIZATION or STORE")
	}

	scopes := make([]string, 0, len(input.Scopes))
	seen := make(map[string]struct{}, len(input.Scopes))
	for _, raw := range input.Scopes {
		code := strings.TrimSpace(raw)
		if code == "" {
			return normalizedRoleInput{}, validationError("scopes", "scope codes must not be blank")
		}
		if _, exists := seen[code]; exists {
			continue
		}
		if len(code) > 100 {
			return normalizedRoleInput{}, validationError("scopes", "scope codes must be at most 100 characters")
		}
		seen[code] = struct{}{}
		scopes = append(scopes, code)
	}
	sort.Strings(scopes)

	return normalizedRoleInput{
		key:             key,
		name:            name,
		description:     description,
		assignmentScope: assignmentScope,
		scopes:          scopes,
	}, nil
}

func parseUUID(raw, field string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(strings.TrimSpace(raw)); err != nil || !id.Valid || zeroUUID(id) {
		return pgtype.UUID{}, validationError(field, "must be a valid UUID")
	}
	return id, nil
}

func optionalUUID(raw *string, field string) (pgtype.UUID, error) {
	if raw == nil {
		return pgtype.UUID{}, nil
	}
	return parseUUID(*raw, field)
}

func normalizeExpiration(value *time.Time, now time.Time) (pgtype.Timestamptz, error) {
	if value == nil {
		return pgtype.Timestamptz{}, nil
	}
	expiresAt := value.UTC()
	if !expiresAt.After(now) {
		return pgtype.Timestamptz{}, validationError("expiresAt", "must be in the future")
	}
	return pgtype.Timestamptz{Time: expiresAt, Valid: true}, nil
}

func zeroUUID(id pgtype.UUID) bool {
	for _, value := range id.Bytes {
		if value != 0 {
			return false
		}
	}
	return true
}
