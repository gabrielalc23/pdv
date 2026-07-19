package invitations

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/auth"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/mailer"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

const invitationPath = "/accept-invitation"

type Config struct {
	TokenHashKey       []byte
	PublicURL          string
	InvitationTTL      time.Duration
	OwnerInvitationTTL time.Duration
	AccessTokenTTL     time.Duration
}

type Service struct {
	store       Store
	tx          TxProvider
	hasher      password.Hasher
	policy      password.Policy
	blocklist   password.Blocklist
	signer      *jwt.Signer
	clock       clock.Clock
	mailer      mailer.Mailer
	invalidator CacheInvalidator
	codec       TokenCodec
	publicURL   *url.URL
	cfg         Config
}

type AcceptResult struct {
	Auth       *auth.AuthResult
	Membership AcceptedResponse
}

func NewService(store Store, tx TxProvider, hasher password.Hasher, policy password.Policy, blocklist password.Blocklist, signer *jwt.Signer, clk clock.Clock, mail mailer.Mailer, invalidator CacheInvalidator, cfg Config) (*Service, error) {
	if store == nil || tx == nil || hasher == nil || signer == nil || clk == nil || mail == nil || invalidator == nil {
		return nil, fmt.Errorf("invitation service dependencies are required")
	}
	codec, err := NewTokenCodec(cfg.TokenHashKey)
	if err != nil {
		return nil, err
	}
	publicURL, err := parsePublicURL(cfg.PublicURL)
	if err != nil {
		return nil, err
	}
	if cfg.InvitationTTL == 0 {
		cfg.InvitationTTL = 7 * 24 * time.Hour
	}
	if cfg.OwnerInvitationTTL == 0 {
		cfg.OwnerInvitationTTL = cfg.InvitationTTL
	}
	if cfg.InvitationTTL <= 0 || cfg.OwnerInvitationTTL <= 0 || cfg.OwnerInvitationTTL > cfg.InvitationTTL || cfg.AccessTokenTTL <= 0 {
		return nil, fmt.Errorf("invalid invitation service TTL configuration")
	}
	return &Service{store: store, tx: tx, hasher: hasher, policy: policy, blocklist: blocklist, signer: signer, clock: clk, mailer: mail, invalidator: invalidator, codec: codec, publicURL: publicURL, cfg: cfg}, nil
}

func (s *Service) List(ctx context.Context, actor authcontext.Principal, input ListInput) (ListResponse, error) {
	if err := requireActor(actor, authz.ScopeInvitationsRead); err != nil {
		return ListResponse{}, err
	}
	page, pageSize := input.Page, input.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return ListResponse{}, validationError("pagination", "Paginação inválida.")
	}
	status, err := optionalInvitationStatus(input.Status)
	if err != nil {
		return ListResponse{}, err
	}
	createdFrom, err := parseOptionalTime("createdFrom", input.CreatedFrom)
	if err != nil {
		return ListResponse{}, err
	}
	createdTo, err := parseOptionalTime("createdTo", input.CreatedTo)
	if err != nil {
		return ListResponse{}, err
	}
	email := pgtype.Text{String: normalizeEmail(input.Email), Valid: strings.TrimSpace(input.Email) != ""}
	params := database.ListInvitationsParams{OrganizationID: actor.OrganizationID, Status: status, Email: email, CreatedFrom: createdFrom, CreatedTo: createdTo, PageOffset: int32((page - 1) * pageSize), PageSize: int32(pageSize)}
	rows, err := s.store.ListInvitations(ctx, params)
	if err != nil {
		return ListResponse{}, dependency("list invitations", err)
	}
	total, err := s.store.CountInvitations(ctx, database.CountInvitationsParams{OrganizationID: actor.OrganizationID, Status: status, Email: email, CreatedFrom: createdFrom, CreatedTo: createdTo})
	if err != nil {
		return ListResponse{}, dependency("count invitations", err)
	}
	data := make([]InvitationResponse, 0, len(rows))
	for _, row := range rows {
		assignments, err := s.assignmentResponses(ctx, actor.OrganizationID, row.ID)
		if err != nil {
			return ListResponse{}, err
		}
		data = append(data, InvitationResponse{ID: row.ID.String(), Email: row.Email, Status: statusString(row.Status), ExpiresAt: formatTime(row.ExpiresAt.Time), Assignments: assignments, CreatedAt: formatTime(row.CreatedAt.Time), UpdatedAt: formatTime(row.UpdatedAt.Time)})
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return ListResponse{Data: data, Pagination: PaginationResponse{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}}, nil
}

func (s *Service) Create(ctx context.Context, actor authcontext.Principal, input CreateInput, meta requestmeta.RequestMetadata) (InvitationResponse, error) {
	if err := requireActor(actor, authz.ScopeMembersInvite); err != nil {
		return InvitationResponse{}, err
	}
	if err := validateCreate(&input); err != nil {
		return InvitationResponse{}, err
	}
	normalized := normalizeEmail(input.Email)
	secret, secretHash, err := s.codec.Prepare()
	if err != nil {
		return InvitationResponse{}, dependency("generate token", err)
	}
	var invitation database.OrganizationInvitation
	var assignments []AssignmentResponse
	var organizationName, inviterName, rawToken string
	err = s.tx.WithTx(ctx, func(tx TxStore) error {
		organization, err := tx.GetOrganization(ctx, actor.OrganizationID)
		if err != nil {
			return dependency("get organization", err)
		}
		inviter, err := tx.GetMembership(ctx, actor.OrganizationID, actor.MembershipID)
		if err != nil || inviter.UserID != actor.UserID || inviter.Status != database.MembershipStatusACTIVE {
			return ErrInsufficientScope
		}
		validated, payload, ttl, err := validateAssignments(ctx, tx, actor, input.Assignments, s.cfg)
		if err != nil {
			return err
		}
		if err := tx.ExpirePendingInvitationsForEmail(ctx, actor.OrganizationID, normalized); err != nil {
			return dependency("expire stale invitations", err)
		}
		invitation, err = tx.CreateInvitation(ctx, database.CreateInvitationParams{OrganizationID: actor.OrganizationID, Email: input.Email, EmailNormalized: normalized, SecretHash: secretHash, ExpiresAt: pgtype.Timestamptz{Time: s.clock.Now().Add(ttl), Valid: true}, InvitedByMembershipID: actor.MembershipID})
		if err != nil {
			return mapPersistenceError(err)
		}
		if _, err := tx.CreateInvitationRoleBindings(ctx, database.CreateInvitationRoleBindingsParams{OrganizationID: actor.OrganizationID, InvitationID: invitation.ID, Assignments: payload}); err != nil {
			return mapPersistenceError(err)
		}
		rawToken, err = s.codec.Format(invitation.ID, secret)
		if err != nil {
			return dependency("format token", err)
		}
		assignments = validated
		organizationName, inviterName = organization.Name, inviter.DisplayName
		return tx.WriteAudit(ctx, invitationAudit(actor, invitation.ID, audit.EventMembershipInvited, meta, map[string]any{"assignment_count": len(assignments), "resend": false}))
	})
	if err != nil {
		return InvitationResponse{}, err
	}
	s.send(ctx, invitation.Email, organizationName, inviterName, rawToken)
	return mapInvitation(invitation, assignments), nil
}

func (s *Service) Resend(ctx context.Context, actor authcontext.Principal, rawID string, meta requestmeta.RequestMetadata) (InvitationResponse, error) {
	if err := requireActor(actor, authz.ScopeInvitationsManage); err != nil {
		return InvitationResponse{}, err
	}
	id, err := parseUUID("id", rawID)
	if err != nil {
		return InvitationResponse{}, err
	}
	var invitation database.OrganizationInvitation
	var assignments []AssignmentResponse
	var organizationName, inviterName, rawToken string
	err = s.tx.WithTx(ctx, func(tx TxStore) error {
		locked, err := tx.GetInvitation(ctx, id)
		if err != nil || locked.OrganizationID != actor.OrganizationID {
			return ErrInvitationNotFound
		}
		if err := invitationStateError(locked, s.clock.Now(), true); err != nil {
			return err
		}
		var hash []byte
		rawToken, hash, err = s.codec.Generate(locked.ID)
		if err != nil {
			return dependency("rotate token", err)
		}
		assignments, err = assignmentResponses(ctx, tx, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		ttl := s.cfg.InvitationTTL
		for _, assignment := range assignments {
			if assignment.Role.Key == "owner" {
				if !actor.Scopes.Has(authz.ScopeOrganizationOwners) {
					return ErrInsufficientScope
				}
				ttl = s.cfg.OwnerInvitationTTL
				break
			}
		}
		invitation, err = tx.RotateInvitationSecret(ctx, database.RotateInvitationSecretParams{SecretHash: hash, ExpiresAt: pgtype.Timestamptz{Time: s.clock.Now().Add(ttl), Valid: true}, OrganizationID: actor.OrganizationID, InvitationID: id})
		if err != nil {
			return mapPersistenceError(err)
		}
		organization, err := tx.GetOrganization(ctx, actor.OrganizationID)
		if err != nil {
			return dependency("get organization", err)
		}
		inviter, err := tx.GetMembership(ctx, actor.OrganizationID, locked.InvitedByMembershipID)
		if err != nil {
			return dependency("get inviter", err)
		}
		organizationName, inviterName = organization.Name, inviter.DisplayName
		return tx.WriteAudit(ctx, invitationAudit(actor, id, audit.EventMembershipInvited, meta, map[string]any{"assignment_count": len(assignments), "resend": true}))
	})
	if err != nil {
		return InvitationResponse{}, err
	}
	s.send(ctx, invitation.Email, organizationName, inviterName, rawToken)
	return mapInvitation(invitation, assignments), nil
}

func (s *Service) Revoke(ctx context.Context, actor authcontext.Principal, rawID string, meta requestmeta.RequestMetadata) error {
	if err := requireActor(actor, authz.ScopeInvitationsManage); err != nil {
		return err
	}
	id, err := parseUUID("id", rawID)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(tx TxStore) error {
		locked, err := tx.GetInvitation(ctx, id)
		if err != nil || locked.OrganizationID != actor.OrganizationID {
			return ErrInvitationNotFound
		}
		if err := invitationStateError(locked, s.clock.Now(), false); err != nil {
			return err
		}
		if _, err := tx.RevokeInvitation(ctx, database.RevokeInvitationParams{OrganizationID: actor.OrganizationID, InvitationID: id}); err != nil {
			return mapPersistenceError(err)
		}
		return tx.WriteAudit(ctx, invitationAudit(actor, id, "membership.invitation_revoked", meta, nil))
	})
}

func (s *Service) Inspect(ctx context.Context, rawToken string) (InspectResponse, error) {
	parsed, err := s.codec.Parse(rawToken)
	if err != nil {
		return InspectResponse{}, ErrInvitationNotFound
	}
	invitation, err := s.store.GetInvitation(ctx, parsed.Selector)
	if err != nil || !s.codec.Verify(parsed.Secret, invitation.SecretHash) {
		return InspectResponse{}, ErrInvitationNotFound
	}
	if err := invitationStateError(invitation, s.clock.Now(), false); err != nil {
		return InspectResponse{}, err
	}
	organization, err := s.store.GetOrganization(ctx, invitation.OrganizationID)
	if err != nil {
		return InspectResponse{}, dependency("get organization", err)
	}
	assignments, err := s.assignmentResponses(ctx, invitation.OrganizationID, invitation.ID)
	if err != nil {
		return InspectResponse{}, err
	}
	_, userErr := s.store.GetUserByNormalizedEmail(ctx, invitation.EmailNormalized)
	if userErr != nil && !errors.Is(userErr, pgx.ErrNoRows) {
		return InspectResponse{}, dependency("inspect user", userErr)
	}
	return InspectResponse{Organization: OrganizationResponse{Name: organization.Name, Slug: organization.Slug}, EmailMasked: maskEmail(invitation.Email), ExpiresAt: formatTime(invitation.ExpiresAt.Time), Assignments: assignments, ExistingUser: userErr == nil}, nil
}

func (s *Service) Accept(ctx context.Context, principal *authcontext.Principal, input AcceptInput, meta requestmeta.RequestMetadata) (AcceptResult, error) {
	parsed, err := s.codec.Parse(input.Token)
	if err != nil {
		return AcceptResult{}, ErrInvitationNotFound
	}
	var passwordHash string
	if principal == nil {
		if err := validateAnonymousAccept(&input); err != nil {
			return AcceptResult{}, err
		}
		preflight, err := s.store.GetInvitation(ctx, parsed.Selector)
		if err != nil || !s.codec.Verify(parsed.Secret, preflight.SecretHash) {
			return AcceptResult{}, ErrInvitationNotFound
		}
		if err := invitationStateError(preflight, s.clock.Now(), false); err != nil {
			return AcceptResult{}, err
		}
		if err := s.policy.Validate(input.Password, preflight.EmailNormalized, s.blocklist); err != nil {
			if errors.Is(err, password.ErrPasswordCommon) {
				return AcceptResult{}, ErrCommonPassword
			}
			return AcceptResult{}, ErrWeakPassword
		}
		passwordHash, err = s.hasher.Hash(input.Password)
		if err != nil {
			return AcceptResult{}, dependency("hash password", err)
		}
	} else {
		input.ClientID = principal.ClientID
	}
	var result AcceptResult
	var sessionIDs []pgtype.UUID
	var acceptedMembershipID pgtype.UUID
	err = s.tx.WithTx(ctx, func(tx TxStore) error {
		invitation, err := tx.GetInvitation(ctx, parsed.Selector)
		if err != nil || !s.codec.Verify(parsed.Secret, invitation.SecretHash) {
			return ErrInvitationNotFound
		}
		if err := invitationStateError(invitation, s.clock.Now(), false); err != nil {
			return err
		}
		organization, err := tx.GetOrganization(ctx, invitation.OrganizationID)
		if err != nil || organization.Status != database.OrganizationStatusACTIVE {
			return ErrInvitationNotFound
		}
		var user database.User
		newUser := principal == nil
		if principal != nil {
			user, err = tx.GetUserByID(ctx, principal.UserID)
			if err != nil || user.EmailNormalized != invitation.EmailNormalized {
				return ErrEmailMismatch
			}
		} else {
			if _, err := tx.GetUserByNormalizedEmail(ctx, invitation.EmailNormalized); err == nil {
				return ErrAuthenticationRequired
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return dependency("check existing user", err)
			}
			created, err := tx.CreateUser(ctx, database.CreateUserWithPasswordParams{Email: invitation.Email, EmailNormalized: invitation.EmailNormalized, DisplayName: input.DisplayName, PasswordHash: passwordHash})
			if err != nil {
				return mapAnonymousCreateError(err)
			}
			if err := tx.VerifyUserEmail(ctx, created.ID); err != nil {
				return dependency("verify invited email", err)
			}
			user = database.User{ID: created.ID, Email: created.Email, EmailNormalized: created.EmailNormalized, DisplayName: created.DisplayName, Status: created.Status, EmailVerifiedAt: pgtype.Timestamptz{Time: s.clock.Now(), Valid: true}, PasswordVersion: created.PasswordVersion, CreatedAt: created.CreatedAt, UpdatedAt: created.UpdatedAt}
		}
		if user.Status == database.UserStatusSUSPENDED {
			return ErrUserSuspended
		}
		if user.Status == database.UserStatusDISABLED {
			return ErrUserDisabled
		}
		membership, reactivated, err := ensureMembership(ctx, tx, invitation, user.ID)
		if err != nil {
			return err
		}
		acceptedMembershipID = membership.ID
		if reactivated {
			if err := tx.DeleteMembershipBindings(ctx, invitation.OrganizationID, membership.ID); err != nil {
				return dependency("replace suspended membership bindings", err)
			}
		}
		bindings, err := tx.CopyBindings(ctx, database.CreateMembershipBindingsFromInvitationParams{MembershipID: membership.ID, OrganizationID: invitation.OrganizationID, InvitationID: invitation.ID})
		if err != nil {
			return dependency("copy invitation bindings", err)
		}
		version, err := tx.IncrementMembershipVersion(ctx, invitation.OrganizationID, membership.ID)
		if err != nil {
			return dependency("increment membership authorization version", err)
		}
		membership.AuthorizationVersion = version
		if !newUser {
			sessionIDs, err = tx.ListSessionIDs(ctx, invitation.OrganizationID, membership.ID)
			if err != nil {
				return dependency("list affected sessions", err)
			}
		}
		if _, err := tx.AcceptInvitation(ctx, database.AcceptInvitationParams{AcceptedByMembershipID: membership.ID, OrganizationID: invitation.OrganizationID, InvitationID: invitation.ID}); err != nil {
			return mapPersistenceError(err)
		}
		eventType := audit.EventMembershipJoined
		if reactivated {
			eventType = audit.EventMembershipReactivated
		}
		actorSession := pgtype.UUID{}
		if principal != nil {
			actorSession = principal.SessionID
		}
		eventActor := authcontext.Principal{UserID: user.ID, SessionID: actorSession, OrganizationID: invitation.OrganizationID, MembershipID: membership.ID}
		if err := tx.WriteAudit(ctx, invitationAudit(eventActor, invitation.ID, eventType, meta, map[string]any{"membership_id": membership.ID.String(), "binding_count": len(bindings), "authorization_version": version})); err != nil {
			return dependency("write acceptance audit", err)
		}
		result.Membership = AcceptedResponse{Status: "ACCEPTED", MembershipID: membership.ID.String()}
		scopes, err := tx.ResolveScopes(ctx, database.ResolveEffectiveScopesParams{ContextKind: database.AuthContextKindORGANIZATION, OrganizationID: invitation.OrganizationID, MembershipID: membership.ID, UserID: user.ID})
		if err != nil {
			return dependency("resolve invitation scopes", err)
		}
		createdSession, err := tx.CreateSession(ctx, sessions.CreateSessionInput{UserID: user.ID, ClientID: input.ClientID, DeviceName: nullableText(input.DeviceName), UserAgent: nullableText(meta.UserAgent), IPAddress: parseIP(meta.ClientIP), ContextKind: sessions.ContextOrganization, OrganizationID: invitation.OrganizationID, MembershipID: membership.ID})
		if err != nil {
			return dependency("create invitation session", err)
		}
		authResult, err := s.issue(user, organization, membership, scopes, createdSession)
		if err != nil {
			return dependency("issue access token", err)
		}
		result.Auth = &authResult
		return nil
	})
	if err != nil {
		return AcceptResult{}, err
	}
	for _, id := range sessionIDs {
		s.invalidator.InvalidateSession(ctx, id)
	}
	s.invalidator.InvalidateMembershipAuthorizationVersion(ctx, acceptedMembershipID)
	return result, nil
}

func ensureMembership(ctx context.Context, tx TxStore, invitation database.OrganizationInvitation, userID pgtype.UUID) (database.OrganizationMembership, bool, error) {
	membership, err := tx.GetLatestMembership(ctx, invitation.OrganizationID, userID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && membership.Status == database.MembershipStatusREMOVED) {
		inviter, inviterErr := tx.GetMembership(ctx, invitation.OrganizationID, invitation.InvitedByMembershipID)
		if inviterErr != nil {
			return database.OrganizationMembership{}, false, dependency("get invitation creator", inviterErr)
		}
		created, createErr := tx.CreateMembership(ctx, database.CreateMembershipParams{OrganizationID: invitation.OrganizationID, UserID: userID, CreatedByUserID: inviter.UserID})
		if createErr != nil {
			return database.OrganizationMembership{}, false, mapPersistenceError(createErr)
		}
		return created, false, nil
	}
	if err != nil {
		return database.OrganizationMembership{}, false, dependency("get membership", err)
	}
	locked, err := tx.LockMembership(ctx, invitation.OrganizationID, membership.ID)
	if err != nil {
		return database.OrganizationMembership{}, false, dependency("lock membership", err)
	}
	if locked.Status == database.MembershipStatusSUSPENDED {
		updated, err := tx.UpdateMembershipStatus(ctx, database.UpdateMembershipStatusParams{Status: database.MembershipStatusACTIVE, OrganizationID: invitation.OrganizationID, MembershipID: locked.ID})
		if err != nil {
			return database.OrganizationMembership{}, false, dependency("reactivate membership", err)
		}
		return updated, true, nil
	}
	return locked, false, nil
}

func (s *Service) issue(user database.User, organization database.Organization, membership database.OrganizationMembership, scopes database.ResolveEffectiveScopesRow, created sessions.CreateSessionResult) (auth.AuthResult, error) {
	jti, err := randomUUID()
	if err != nil {
		return auth.AuthResult{}, err
	}
	oav, mav := organization.AuthorizationVersion, membership.AuthorizationVersion
	token, err := s.signer.Sign(jwt.SignerClaims{Subject: user.ID.String(), JTI: jti.String(), SessionID: created.Session.ID.String(), ClientID: created.Session.ClientID, Ctx: jwt.ContextOrganization, OrgID: organization.ID.String(), MembershipID: membership.ID.String(), Roles: scopes.RoleKeys, Scopes: scopes.ScopeCodes, OAV: &oav, MAV: &mav, PV: user.PasswordVersion, AuthTime: created.Session.CreatedAt, AMR: []string{"pwd", "invitation"}})
	if err != nil {
		return auth.AuthResult{}, err
	}
	membershipID := membership.ID.String()
	response := auth.AuthResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(s.cfg.AccessTokenTTL / time.Second), User: auth.UserResponse{ID: user.ID.String(), Email: user.Email, DisplayName: user.DisplayName, EmailVerified: true}, Session: auth.SessionResponse{ID: created.Session.ID.String(), ClientID: created.Session.ClientID, CreatedAt: formatTime(created.Session.CreatedAt), IdleExpiresAt: formatTime(created.Session.IdleExpiresAt), AbsoluteExpiresAt: formatTime(created.Session.AbsoluteExpiresAt)}, Context: auth.ContextResponse{Kind: "organization", MembershipID: &membershipID, Organization: &auth.OrganizationResponse{ID: organization.ID.String(), Name: organization.Name, Slug: organization.Slug}, Roles: scopes.RoleKeys, Scopes: scopes.ScopeCodes}}
	return auth.AuthResult{Response: response, RawRefreshToken: created.RawRefreshToken, RefreshExpires: created.Session.AbsoluteExpiresAt}, nil
}

func validateAssignments(ctx context.Context, tx TxStore, actor authcontext.Principal, inputs []AssignmentInput, cfg Config) ([]AssignmentResponse, []byte, time.Duration, error) {
	type assignmentPayload struct {
		RoleID  string  `json:"role_id"`
		StoreID *string `json:"store_id"`
	}
	seen := make(map[string]struct{}, len(inputs))
	responses := make([]AssignmentResponse, 0, len(inputs))
	payload := make([]assignmentPayload, 0, len(inputs))
	ttl := cfg.InvitationTTL
	granted, err := tx.ResolveActorGrantScopes(ctx, actor.OrganizationID, actor.MembershipID)
	if err != nil {
		return nil, nil, 0, dependency("resolve inviter scopes", err)
	}
	grantSet := make(map[string]struct{}, len(granted))
	for _, scope := range granted {
		grantSet[scope] = struct{}{}
	}
	for index, input := range inputs {
		roleID, err := parseUUID(fmt.Sprintf("assignments[%d].roleId", index), input.RoleID)
		if err != nil {
			return nil, nil, 0, err
		}
		storeID, err := optionalUUID(fmt.Sprintf("assignments[%d].storeId", index), input.StoreID)
		if err != nil {
			return nil, nil, 0, err
		}
		key := roleID.String() + ":" + storeID.String()
		if _, exists := seen[key]; exists {
			return nil, nil, 0, validationError("assignments", "Atribuição duplicada.")
		}
		seen[key] = struct{}{}
		role, err := tx.GetRole(ctx, actor.OrganizationID, roleID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, 0, ErrRoleNotFound
		}
		if err != nil {
			return nil, nil, 0, dependency("get role", err)
		}
		if !role.IsActive {
			return nil, nil, 0, ErrRoleInactive
		}
		roleScopes, err := tx.ListScopeCodesForRole(ctx, actor.OrganizationID, roleID)
		if err != nil {
			return nil, nil, 0, dependency("list role scopes", err)
		}
		for _, scope := range roleScopes {
			if _, allowed := grantSet[scope]; !allowed {
				return nil, nil, 0, ErrInsufficientScope
			}
		}
		if role.Key == "owner" && role.IsSystem {
			if !actor.Scopes.Has(authz.ScopeOrganizationOwners) {
				return nil, nil, 0, ErrInsufficientScope
			}
			ttl = cfg.OwnerInvitationTTL
		}
		if err := validateRoleStore(role, storeID, fmt.Sprintf("assignments[%d].storeId", index)); err != nil {
			return nil, nil, 0, err
		}
		response := AssignmentResponse{Role: RoleResponse{ID: role.ID.String(), Key: role.Key, Name: role.Name}}
		var storeText *string
		if storeID.Valid {
			store, err := tx.GetStore(ctx, actor.OrganizationID, storeID)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, 0, ErrStoreNotFound
			}
			if err != nil {
				return nil, nil, 0, dependency("get store", err)
			}
			if store.Status != database.StoreStatusACTIVE {
				return nil, nil, 0, ErrStoreInactive
			}
			value := storeID.String()
			storeText = &value
			response.Store = &StoreResponse{ID: value, Code: store.Code, Name: store.Name}
		}
		responses = append(responses, response)
		payload = append(payload, assignmentPayload{RoleID: roleID.String(), StoreID: storeText})
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, 0, dependency("encode assignments", err)
	}
	return responses, encoded, ttl, nil
}

func validateRoleStore(role database.Role, storeID pgtype.UUID, field string) error {
	if role.AssignmentScope == database.RoleAssignmentScopeSTORE && !storeID.Valid {
		return validationError(field, "Loja obrigatória para role de loja.")
	}
	if role.AssignmentScope == database.RoleAssignmentScopeORGANIZATION && storeID.Valid {
		return validationError(field, "Loja proibida para role de organização.")
	}
	return nil
}

func (s *Service) assignmentResponses(ctx context.Context, organizationID, invitationID pgtype.UUID) ([]AssignmentResponse, error) {
	return assignmentResponses(ctx, s.store, organizationID, invitationID)
}

func assignmentResponses(ctx context.Context, store Store, organizationID, invitationID pgtype.UUID) ([]AssignmentResponse, error) {
	bindings, err := store.ListInvitationRoleBindings(ctx, organizationID, invitationID)
	if err != nil {
		return nil, dependency("list invitation assignments", err)
	}
	result := make([]AssignmentResponse, 0, len(bindings))
	for _, binding := range bindings {
		role, err := store.GetRole(ctx, organizationID, binding.RoleID)
		if err != nil {
			return nil, dependency("get assignment role", err)
		}
		item := AssignmentResponse{Role: RoleResponse{ID: role.ID.String(), Key: role.Key, Name: role.Name}}
		if binding.StoreID.Valid {
			item.Store = &StoreResponse{ID: binding.StoreID.String(), Code: binding.StoreCode.String, Name: binding.StoreName.String}
		}
		result = append(result, item)
	}
	return result, nil
}

func requireActor(actor authcontext.Principal, scope authcontext.Scope) error {
	if !validUUID(actor.UserID) || !validUUID(actor.SessionID) || !validUUID(actor.OrganizationID) || !validUUID(actor.MembershipID) {
		return validationError("context", "Contexto de organização obrigatório.")
	}
	if !actor.Scopes.Has(scope) {
		return ErrInsufficientScope
	}
	return nil
}

func invitationStateError(invitation database.OrganizationInvitation, now time.Time, resend bool) error {
	switch invitation.Status {
	case database.InvitationStatusACCEPTED:
		return ErrInvitationAccepted
	case database.InvitationStatusREVOKED:
		return ErrInvitationRevoked
	case database.InvitationStatusEXPIRED:
		if resend {
			return ErrInvitationExpired
		}
		return ErrInvitationExpired
	}
	if !now.Before(invitation.ExpiresAt.Time) && !resend {
		return ErrInvitationExpired
	}
	return nil
}

func optionalInvitationStatus(value string) (database.NullInvitationStatus, error) {
	if strings.TrimSpace(value) == "" {
		return database.NullInvitationStatus{}, nil
	}
	status := database.InvitationStatus(strings.ToUpper(strings.TrimSpace(value)))
	switch status {
	case database.InvitationStatusPENDING, database.InvitationStatusACCEPTED, database.InvitationStatusREVOKED, database.InvitationStatusEXPIRED:
		return database.NullInvitationStatus{InvitationStatus: status, Valid: true}, nil
	default:
		return database.NullInvitationStatus{}, validationError("status", "Status inválido.")
	}
}

func statusString(value any) string {
	switch typed := value.(type) {
	case database.InvitationStatus:
		return string(typed)
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func mapInvitation(invitation database.OrganizationInvitation, assignments []AssignmentResponse) InvitationResponse {
	return InvitationResponse{ID: invitation.ID.String(), Email: invitation.Email, Status: string(invitation.Status), ExpiresAt: formatTime(invitation.ExpiresAt.Time), Assignments: assignments, CreatedAt: formatTime(invitation.CreatedAt.Time), UpdatedAt: formatTime(invitation.UpdatedAt.Time)}
}

func mapPersistenceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvitationNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "idx_organization_invitations_pending_email":
			return ErrInvitationPending
		case "invitation_role_bindings_owner_inviter_check", "organization_invitations_owner_inviter_check":
			return ErrInsufficientScope
		case "invitation_role_bindings_role_fkey":
			return ErrRoleNotFound
		case "invitation_role_bindings_store_fkey":
			return ErrStoreNotFound
		}
	}
	return dependency("persistence operation", err)
}

func mapAnonymousCreateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "users_email_normalized_unique" {
		return ErrAuthenticationRequired
	}
	return mapPersistenceError(err)
}

func invitationAudit(actor authcontext.Principal, entityID pgtype.UUID, eventType string, meta requestmeta.RequestMetadata, metadata map[string]any) audit.Event {
	if metadata == nil {
		metadata = map[string]any{}
	}
	encoded, _ := json.Marshal(metadata)
	var ip *netip.Addr
	if parsed, err := netip.ParseAddr(meta.ClientIP); err == nil {
		ip = &parsed
	}
	return audit.Event{OrganizationID: actor.OrganizationID, ActorUserID: actor.UserID, ActorMembershipID: actor.MembershipID, SessionID: actor.SessionID, EventType: eventType, EntityType: pgtype.Text{String: "organization_invitation", Valid: true}, EntityID: entityID, RequestID: nullableText(meta.RequestID), IPAddress: ip, UserAgent: nullableText(meta.UserAgent), Outcome: database.AuditOutcomeSUCCESS, Metadata: encoded}
}

func (s *Service) send(ctx context.Context, to, organization, inviter, token string) {
	mailCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	if err := s.mailer.SendInvitation(mailCtx, to, organization, inviter, s.invitationLink(token)); err != nil {
		slog.Error("invitation email delivery failed", "error", err)
	}
}

func (s *Service) invitationLink(token string) string {
	link := *s.publicURL
	link.Path = strings.TrimRight(link.Path, "/") + invitationPath
	link.RawPath = ""
	link.Fragment = url.Values{"token": []string{token}}.Encode()
	return link.String()
}

func parsePublicURL(value string) (*url.URL, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return nil, fmt.Errorf("APP_PUBLIC_URL must be a non-empty absolute HTTP(S) URL")
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.Opaque != "" {
		return nil, fmt.Errorf("APP_PUBLIC_URL must be an absolute HTTP(S) URL")
	}
	parsed.RawQuery, parsed.Fragment, parsed.RawFragment = "", "", ""
	parsed.ForceQuery = false
	return parsed, nil
}

func maskEmail(email string) string {
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" {
		return "***"
	}
	runes := []rune(local)
	return string(runes[0]) + "***@" + domain
}

func randomUUID() (pgtype.UUID, error) {
	var id pgtype.UUID
	if _, err := rand.Read(id.Bytes[:]); err != nil {
		return id, err
	}
	id.Bytes[6] = (id.Bytes[6] & 0x0f) | 0x40
	id.Bytes[8] = (id.Bytes[8] & 0x3f) | 0x80
	id.Valid = true
	return id, nil
}

func nullableText(value string) pgtype.Text { return pgtype.Text{String: value, Valid: value != ""} }
func parseIP(value string) *netip.Addr {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return nil
	}
	return &ip
}
func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339) }
func dependency(operation string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrDependencyUnavailable, operation, err)
}
