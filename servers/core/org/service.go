package org

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

const keycloakTimeout = 30 * time.Second

const (
	uniqueViolation      = "23505"
	foreignKeyViolation  = "23503"
	slugUniqueConstraint = "unique_org_slug"
	parentOrgConstraint  = "fk_parent_org"
	courseOrgConstraint  = "fk_course_org"
)

var (
	ErrOrgNotFound       = errors.New("org not found")
	ErrDuplicateOrgSlug  = errors.New("an org with this slug already exists")
	ErrParentOrgNotFound = errors.New("parent org not found")
	ErrOrgCycle          = errors.New("an org cannot be placed under itself or one of its descendants")
	ErrOrgHasChildOrgs   = errors.New("the org still has child orgs")
	ErrOrgHasCourses     = errors.New("the org still has courses")
	ErrOrgMembersRemoved = errors.New("the org was not deleted, but not all of its members could be re-added to its Keycloak groups: the server log names them; re-add them, or delete the org again")
	ErrOwnMembership     = errors.New("you cannot change your own org membership; ask another Org Admin or a PROMPT admin")
)

// orgMembers maps an org role to the Keycloak IDs of the users in its group.
type orgMembers = map[string][]string

// OrgMemberDirectory reads and changes the membership of orgs, which lives in Keycloak only.
type OrgMemberDirectory interface {
	GetOrgMembers(ctx context.Context, slug string) ([]keycloakRealmDTO.OrgMember, error)
	FindUserByUniversityLogin(ctx context.Context, universityLogin string) (keycloakRealmDTO.StaffMember, error)
	SetOrgMemberRole(ctx context.Context, slug, userID, orgRole string) error
	RemoveOrgMember(ctx context.Context, slug, userID string) error
}

type OrgService struct {
	queries                 db.Queries
	conn                    *pgxpool.Pool
	createOrgGroupsAndRoles func(ctx context.Context, slug string) error
	deleteOrgGroupsAndRoles func(ctx context.Context, slug string) (orgMembers, error)
	orgGroupsOrRolesExist   func(ctx context.Context, slug string) (bool, error)
	addOrgMembers           func(ctx context.Context, slug string, members orgMembers) error
	members                 OrgMemberDirectory
}

func NewOrgService(queries db.Queries, conn *pgxpool.Pool,
	createOrgGroupsAndRoles func(ctx context.Context, slug string) error,
	deleteOrgGroupsAndRoles func(ctx context.Context, slug string) (orgMembers, error),
	orgGroupsOrRolesExist func(ctx context.Context, slug string) (bool, error),
	addOrgMembers func(ctx context.Context, slug string, members orgMembers) error,
	members OrgMemberDirectory) *OrgService {
	return &OrgService{
		queries:                 queries,
		conn:                    conn,
		createOrgGroupsAndRoles: createOrgGroupsAndRoles,
		deleteOrgGroupsAndRoles: deleteOrgGroupsAndRoles,
		addOrgMembers:           addOrgMembers,
		orgGroupsOrRolesExist:   orgGroupsOrRolesExist,
		members:                 members,
	}
}

func (s *OrgService) GetAllOrgs(ctx context.Context) ([]orgDTO.Org, error) {
	orgs, err := s.queries.GetAllOrgs(ctx)
	if err != nil {
		return nil, err
	}
	return orgDTO.GetOrgDTOsFromDBModels(orgs), nil
}

func (s *OrgService) GetOrg(ctx context.Context, id uuid.UUID) (orgDTO.Org, error) {
	org, err := s.queries.GetOrgByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return orgDTO.Org{}, ErrOrgNotFound
	}
	if err != nil {
		return orgDTO.Org{}, err
	}
	return orgDTO.GetOrgDTOFromDBModel(org), nil
}

// GetOwnOrgs returns the orgs whose roles the caller's token holds, each with the caller's role; an
// Admin role wins over a Member role of the same org. As in CheckOrgPermission, an org created after
// the token was issued is left out: the token's roles for its slug belong to a deleted org.
func (s *OrgService) GetOwnOrgs(ctx context.Context, userRoles map[string]bool, tokenIssuedAt time.Time) ([]orgDTO.OwnOrg, error) {
	roleBySlug := map[string]string{}
	for role, held := range userRoles {
		slug, orgRole, ok := permissionValidation.ParseOrgRole(role)
		if held && ok && roleBySlug[slug] != permissionValidation.OrgAdmin {
			roleBySlug[slug] = orgRole
		}
	}

	ownOrgs := []orgDTO.OwnOrg{}
	if len(roleBySlug) == 0 {
		return ownOrgs, nil
	}
	orgs, err := s.queries.GetOrgsBySlugs(ctx, slices.Collect(maps.Keys(roleBySlug)))
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		if permissionValidation.TokenPostdatesOrg(tokenIssuedAt, org.CreatedAt.Time) {
			ownOrgs = append(ownOrgs, orgDTO.OwnOrg{Org: orgDTO.GetOrgDTOFromDBModel(org), Role: roleBySlug[org.Slug]})
		}
	}
	return ownOrgs, nil
}

func (s *OrgService) GetOrgMembers(ctx context.Context, orgID uuid.UUID) ([]keycloakRealmDTO.OrgMember, error) {
	org, err := s.GetOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	members, err := s.members.GetOrgMembers(keycloakCtx, org.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get the members of org %s: %w", org.Slug, err)
	}
	return members, nil
}

// SetOrgMember gives the user with the university login the org role, adding them to the org if
// they are not a member yet. Nobody changes their own membership, so an Org Admin cannot lock
// themselves out of the org.
func (s *OrgService) SetOrgMember(ctx context.Context, orgID uuid.UUID, input orgDTO.SetOrgMember, callerUserID string) (orgDTO.Org, keycloakRealmDTO.OrgMember, error) {
	org, err := s.GetOrg(ctx, orgID)
	if err != nil {
		return orgDTO.Org{}, keycloakRealmDTO.OrgMember{}, err
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	user, err := s.members.FindUserByUniversityLogin(keycloakCtx, strings.TrimSpace(input.UniversityLogin))
	if err != nil {
		return orgDTO.Org{}, keycloakRealmDTO.OrgMember{}, err
	}
	if user.KeycloakUserID == callerUserID {
		return orgDTO.Org{}, keycloakRealmDTO.OrgMember{}, ErrOwnMembership
	}
	if err := s.members.SetOrgMemberRole(keycloakCtx, org.Slug, user.KeycloakUserID, input.Role); err != nil {
		return orgDTO.Org{}, keycloakRealmDTO.OrgMember{}, fmt.Errorf("failed to set the org role of user %s: %w", user.KeycloakUserID, err)
	}
	return org, keycloakRealmDTO.OrgMember{StaffMember: user, Role: input.Role}, nil
}

// RemoveOrgMember takes the user out of the org. Nobody removes themselves, see SetOrgMember.
func (s *OrgService) RemoveOrgMember(ctx context.Context, orgID uuid.UUID, userID, callerUserID string) (orgDTO.Org, error) {
	org, err := s.GetOrg(ctx, orgID)
	if err != nil {
		return orgDTO.Org{}, err
	}
	if userID == callerUserID {
		return orgDTO.Org{}, ErrOwnMembership
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	if err := s.members.RemoveOrgMember(keycloakCtx, org.Slug, userID); err != nil {
		return orgDTO.Org{}, fmt.Errorf("failed to remove user %s from org %s: %w", userID, org.Slug, err)
	}
	return org, nil
}

func (s *OrgService) CreateOrg(ctx context.Context, input orgDTO.CreateOrg) (orgDTO.Org, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return orgDTO.Org{}, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	created, err := qtx.CreateOrg(ctx, db.CreateOrgParams{
		ID:           uuid.New(),
		ParentOrgID:  toPgUUID(input.ParentOrgID),
		Name:         strings.TrimSpace(input.Name),
		Slug:         strings.TrimSpace(input.Slug),
		School:       optionalText(input.School),
		University:   optionalText(input.University),
		Website:      optionalText(input.Website),
		ContactEmail: optionalText(input.ContactEmail),
	})
	if err != nil {
		return orgDTO.Org{}, mapWriteViolation(err)
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	if err := s.createOrgGroupsAndRoles(keycloakCtx, created.Slug); err != nil {
		s.cleanUpKeycloak(ctx, created.Slug)
		return orgDTO.Org{}, fmt.Errorf("failed to create keycloak groups and roles: %w", err)
	}

	// The insert may have waited on the slug of an org whose deletion was still removing its roles
	// from Keycloak, so the creation time is taken after provisioning, not at the transaction start:
	// a token holding the old org's roles must predate it.
	stamped, err := qtx.StampOrgCreatedAt(ctx, created.ID)
	if err != nil {
		s.cleanUpKeycloak(ctx, created.Slug)
		return orgDTO.Org{}, fmt.Errorf("failed to stamp the org's creation time: %w", err)
	}
	created = stamped

	if err := tx.Commit(ctx); err != nil {
		return s.reconcileFailedCreateCommit(ctx, created, fmt.Errorf("failed to commit transaction: %w", err))
	}
	return orgDTO.GetOrgDTOFromDBModel(created), nil
}

func (s *OrgService) reconcileFailedCreateCommit(ctx context.Context, created db.Org, commitErr error) (orgDTO.Org, error) {
	checkCtx, cancel := db.GetTimeoutContext(context.WithoutCancel(ctx))
	defer cancel()
	orgID, err := s.queries.GetOrgIDBySlug(checkCtx, created.Slug)
	switch {
	case err == nil && orgID == created.ID:
		return orgDTO.GetOrgDTOFromDBModel(created), nil
	case errors.Is(err, pgx.ErrNoRows):
		s.cleanUpKeycloak(ctx, created.Slug)
	case err != nil:
		return orgDTO.Org{}, errors.Join(commitErr, fmt.Errorf("failed to check whether the org was created: %w", err))
	}
	return orgDTO.Org{}, commitErr
}

func (s *OrgService) cleanUpKeycloak(ctx context.Context, slug string) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
	defer cancel()
	if _, err := s.deleteOrgGroupsAndRoles(cleanupCtx, slug); err != nil {
		log.Error("failed to clean up keycloak groups and roles of org ", slug, ": ", err)
	}
}

func (s *OrgService) UpdateOrg(ctx context.Context, id uuid.UUID, input orgDTO.UpdateOrg) (orgDTO.Org, error) {
	updated, err := s.queries.UpdateOrg(ctx, db.UpdateOrgParams{
		ID:           id,
		Name:         strings.TrimSpace(input.Name),
		School:       optionalText(input.School),
		University:   optionalText(input.University),
		Website:      optionalText(input.Website),
		ContactEmail: optionalText(input.ContactEmail),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return orgDTO.Org{}, ErrOrgNotFound
	}
	if err != nil {
		return orgDTO.Org{}, err
	}
	return orgDTO.GetOrgDTOFromDBModel(updated), nil
}

func (s *OrgService) UpdateOrgParent(ctx context.Context, id uuid.UUID, input orgDTO.UpdateOrgParent) (orgDTO.Org, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return orgDTO.Org{}, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if input.ParentOrgID != nil {
		if *input.ParentOrgID == id {
			return orgDTO.Org{}, ErrOrgCycle
		}
		if err := qtx.LockOrgHierarchy(ctx); err != nil {
			return orgDTO.Org{}, fmt.Errorf("failed to lock the org hierarchy: %w", err)
		}
		ancestorIDs, err := qtx.GetOrgAncestorIDs(ctx, *input.ParentOrgID)
		if err != nil {
			return orgDTO.Org{}, fmt.Errorf("failed to get the ancestors of the parent org: %w", err)
		}
		if slices.Contains(ancestorIDs, id) {
			return orgDTO.Org{}, ErrOrgCycle
		}
	}

	updated, err := qtx.UpdateOrgParent(ctx, db.UpdateOrgParentParams{
		ID:          id,
		ParentOrgID: toPgUUID(input.ParentOrgID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return orgDTO.Org{}, ErrOrgNotFound
	}
	if err != nil {
		return orgDTO.Org{}, mapWriteViolation(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return orgDTO.Org{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return orgDTO.GetOrgDTOFromDBModel(updated), nil
}

func (s *OrgService) DeleteOrg(ctx context.Context, id uuid.UUID) (string, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	slug, err := qtx.DeleteOrg(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrOrgNotFound
	}
	if err != nil {
		return "", mapDeleteViolation(err)
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	members, err := s.deleteOrgGroupsAndRoles(keycloakCtx, slug)
	if err != nil {
		if err := s.reconcileFailedKeycloakDelete(ctx, slug, members, fmt.Errorf("failed to delete keycloak groups and roles: %w", err)); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.reconcileFailedDeleteCommit(ctx, id, slug, members, fmt.Errorf("failed to commit transaction: %w", err))
	}
	return slug, nil
}

func (s *OrgService) reconcileFailedKeycloakDelete(ctx context.Context, slug string, members orgMembers, deleteErr error) error {
	checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
	defer cancel()
	exist, err := s.orgGroupsOrRolesExist(checkCtx, slug)
	if err != nil {
		return errors.Join(deleteErr, fmt.Errorf("failed to check the keycloak groups and roles: %w", err))
	}
	if !exist {
		log.Warn("keycloak deleted the groups and roles of org ", slug, " despite an error: ", deleteErr)
		return nil
	}
	if err := s.restoreKeycloak(ctx, slug, members); err != nil {
		return errors.Join(deleteErr, err)
	}
	return deleteErr
}

func (s *OrgService) reconcileFailedDeleteCommit(ctx context.Context, id uuid.UUID, slug string, members orgMembers, commitErr error) (string, error) {
	checkCtx, cancel := db.GetTimeoutContext(context.WithoutCancel(ctx))
	defer cancel()
	_, err := s.queries.GetOrgByID(checkCtx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return slug, nil
	}
	if err != nil {
		return "", errors.Join(commitErr, fmt.Errorf("failed to check whether the org was deleted: %w", err))
	}
	if err := s.restoreKeycloak(ctx, slug, members); err != nil {
		return "", errors.Join(commitErr, err)
	}
	return "", commitErr
}

// restoreKeycloak recreates the org's groups and roles and puts its members back. When only the
// members fail, the groups are back, so that error carries ErrOrgMembersRemoved.
func (s *OrgService) restoreKeycloak(ctx context.Context, slug string, members orgMembers) error {
	restoreCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
	defer cancel()
	if err := s.createOrgGroupsAndRoles(restoreCtx, slug); err != nil {
		return fmt.Errorf("failed to restore keycloak groups and roles: %w", err)
	}
	if err := s.addOrgMembers(restoreCtx, slug, members); err != nil {
		return errors.Join(ErrOrgMembersRemoved, fmt.Errorf("failed to re-add the org's members: %w", err))
	}
	return nil
}

func optionalText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	return pgtype.Text{String: trimmed, Valid: trimmed != ""}
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func mapWriteViolation(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch {
	case pgErr.Code == uniqueViolation && pgErr.ConstraintName == slugUniqueConstraint:
		return ErrDuplicateOrgSlug
	case pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == parentOrgConstraint:
		return ErrParentOrgNotFound
	}
	return err
}

func mapDeleteViolation(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != foreignKeyViolation {
		return err
	}
	switch pgErr.ConstraintName {
	case parentOrgConstraint:
		return ErrOrgHasChildOrgs
	case courseOrgConstraint:
		return ErrOrgHasCourses
	}
	return err
}
