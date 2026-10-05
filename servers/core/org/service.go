package org

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	log "github.com/sirupsen/logrus"
)

// keycloakTimeout bounds each Keycloak provisioning or cleanup call, which runs while
// the org's transaction is open.
const keycloakTimeout = 30 * time.Second

// Postgres error codes and the constraint names of migration 0032_org.
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
)

type OrgService struct {
	queries db.Queries
	conn    *pgxpool.Pool
	// use dependency injection for keycloak to allow mocking
	createOrgGroupsAndRoles func(ctx context.Context, slug string) error
	deleteOrgGroupsAndRoles func(ctx context.Context, slug string) error
}

func NewOrgService(queries db.Queries, conn *pgxpool.Pool,
	createOrgGroupsAndRoles func(ctx context.Context, slug string) error,
	deleteOrgGroupsAndRoles func(ctx context.Context, slug string) error) *OrgService {
	return &OrgService{
		queries:                 queries,
		conn:                    conn,
		createOrgGroupsAndRoles: createOrgGroupsAndRoles,
		deleteOrgGroupsAndRoles: deleteOrgGroupsAndRoles,
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

// CreateOrg stores the org and provisions its Keycloak groups and roles in one
// transaction, so a Keycloak failure leaves no org behind.
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

	if err := tx.Commit(ctx); err != nil {
		return orgDTO.Org{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return orgDTO.GetOrgDTOFromDBModel(created), nil
}

// cleanUpKeycloak removes what a failed provisioning left behind. It runs before the
// deferred rollback: the uncommitted row still holds the slug, so no concurrent creation
// of the same slug can be provisioning it meanwhile. It gets a fresh timeout because the
// provisioning may have failed by running out of its own. A failure is only logged,
// since provisioning is idempotent and a retry reuses any leftovers.
func (s *OrgService) cleanUpKeycloak(ctx context.Context, slug string) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
	defer cancel()
	if err := s.deleteOrgGroupsAndRoles(cleanupCtx, slug); err != nil {
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

// UpdateOrgParent moves the org under another org, or to the top level. Hierarchy
// changes are serialized, so the cycle check cannot be raced by a concurrent change.
func (s *OrgService) UpdateOrgParent(ctx context.Context, id uuid.UUID, input orgDTO.UpdateOrgParent) (orgDTO.Org, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return orgDTO.Org{}, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.LockOrgHierarchy(ctx); err != nil {
		return orgDTO.Org{}, fmt.Errorf("failed to lock the org hierarchy: %w", err)
	}

	if input.ParentOrgID != nil {
		if *input.ParentOrgID == id {
			return orgDTO.Org{}, ErrOrgCycle
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

// DeleteOrg removes the org and then its Keycloak groups and roles. The foreign keys
// reject the delete while child orgs or courses reference the org, before Keycloak is
// touched; a Keycloak failure rolls the delete back and restores what the deletion had
// already removed. If the commit fails after Keycloak succeeded, the org remains without
// groups and roles until the delete is retried, which is safe because both steps
// tolerate what is already gone.
func (s *OrgService) DeleteOrg(ctx context.Context, id uuid.UUID) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	slug, err := qtx.DeleteOrg(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrgNotFound
	}
	if err != nil {
		return mapDeleteViolation(err)
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	if err := s.deleteOrgGroupsAndRoles(keycloakCtx, slug); err != nil {
		s.restoreKeycloak(ctx, slug)
		return fmt.Errorf("failed to delete keycloak groups and roles: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// restoreKeycloak re-provisions what a failed deletion removed, since the rollback keeps
// the org. The deletion removes the group last, so the group and its memberships survive
// a partial failure and only the roles and their mappings need to come back. It gets a
// fresh timeout for the same reason as cleanUpKeycloak, and a failure is only logged.
func (s *OrgService) restoreKeycloak(ctx context.Context, slug string) {
	restoreCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
	defer cancel()
	if err := s.createOrgGroupsAndRoles(restoreCtx, slug); err != nil {
		log.Error("failed to restore keycloak groups and roles of org ", slug, ": ", err)
	}
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

// mapWriteViolation maps constraint violations of inserting or updating an org.
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

// mapDeleteViolation maps the foreign keys that still reference an org being deleted.
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
