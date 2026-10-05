package org

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// keycloakSpy stands in for the Keycloak provisioning hooks: it records every call and
// lets a test make the hooks fail.
type keycloakSpy struct {
	created   []string
	deleted   []string
	createErr error
	deleteErr error
	onDelete  func(slug string)
}

func (k *keycloakSpy) create(_ context.Context, slug string) error {
	k.created = append(k.created, slug)
	return k.createErr
}

func (k *keycloakSpy) delete(_ context.Context, slug string) error {
	k.deleted = append(k.deleted, slug)
	if k.onDelete != nil {
		k.onDelete(slug)
	}
	return k.deleteErr
}

type OrgServiceTestSuite struct {
	suite.Suite
	ctx      context.Context
	cleanup  func()
	conn     *pgxpool.Pool
	keycloak *keycloakSpy
	service  *OrgService
}

func (suite *OrgServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := setupMigratedTestDB(suite.ctx)
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.keycloak = &keycloakSpy{}
	suite.service = NewOrgService(*testDB.Queries, testDB.Conn, suite.keycloak.create, suite.keycloak.delete)
}

func (suite *OrgServiceTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *OrgServiceTestSuite) SetupTest() {
	*suite.keycloak = keycloakSpy{}
}

func (suite *OrgServiceTestSuite) createOrg(slug string, parentOrgID *uuid.UUID) orgDTO.Org {
	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{
		ParentOrgID: parentOrgID,
		Name:        "Org " + slug,
		Slug:        slug,
	})
	require.NoError(suite.T(), err)
	return created
}

func (suite *OrgServiceTestSuite) countOrgs(slug string) int {
	var count int
	err := suite.conn.QueryRow(suite.ctx, "SELECT COUNT(*) FROM org WHERE slug = $1", slug).Scan(&count)
	require.NoError(suite.T(), err)
	return count
}

func (suite *OrgServiceTestSuite) parentOf(orgID uuid.UUID) *uuid.UUID {
	org, err := suite.service.GetOrg(suite.ctx, orgID)
	require.NoError(suite.T(), err)
	return org.ParentOrgID
}

// insertBlocksOnSlug reports whether inserting the slug from another connection has to
// wait for a concurrent, uncommitted transaction that holds it.
func (suite *OrgServiceTestSuite) insertBlocksOnSlug(slug string) bool {
	tx, err := suite.conn.Begin(suite.ctx)
	require.NoError(suite.T(), err)
	defer func() { _ = tx.Rollback(suite.ctx) }()

	_, err = tx.Exec(suite.ctx, "SET LOCAL lock_timeout = '200ms'")
	require.NoError(suite.T(), err)
	_, err = tx.Exec(suite.ctx, "INSERT INTO org (name, slug) VALUES ('Probe', $1)", slug)

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

func (suite *OrgServiceTestSuite) TestCreateOrgStoresTrimmedInputAndProvisionsKeycloak() {
	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{
		Name:         "  Applied Software Engineering  ",
		Slug:         " ase ",
		School:       " CIT ",
		University:   "Technical University of Munich",
		Website:      "https://aet.cit.tum.de",
		ContactEmail: "office@example.com",
	})
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Applied Software Engineering", created.Name)
	assert.Equal(suite.T(), "ase", created.Slug)
	assert.Equal(suite.T(), "CIT", created.School.String)
	assert.Equal(suite.T(), "Technical University of Munich", created.University.String)
	assert.Equal(suite.T(), "https://aet.cit.tum.de", created.Website.String)
	assert.Equal(suite.T(), "office@example.com", created.ContactEmail.String)
	assert.Nil(suite.T(), created.ParentOrgID)
	assert.False(suite.T(), created.CreatedAt.IsZero())
	assert.Equal(suite.T(), []string{"ase"}, suite.keycloak.created)
	assert.Empty(suite.T(), suite.keycloak.deleted)

	stored, err := suite.service.GetOrg(suite.ctx, created.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), created, stored)
}

func (suite *OrgServiceTestSuite) TestCreateOrgStoresEmptyOptionalFieldsAsNull() {
	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{
		Name:    "Minimal Org",
		Slug:    "minimal",
		School:  "   ",
		Website: "",
	})
	require.NoError(suite.T(), err)

	assert.False(suite.T(), created.School.Valid)
	assert.False(suite.T(), created.University.Valid)
	assert.False(suite.T(), created.Website.Valid)
	assert.False(suite.T(), created.ContactEmail.Valid)
}

func (suite *OrgServiceTestSuite) TestCreateOrgUnderParent() {
	parent := suite.createOrg("parent-chair", nil)

	child := suite.createOrg("parent-chair-lab", &parent.ID)

	require.NotNil(suite.T(), child.ParentOrgID)
	assert.Equal(suite.T(), parent.ID, *child.ParentOrgID)
}

func (suite *OrgServiceTestSuite) TestCreateOrgRejectsDuplicateSlug() {
	suite.createOrg("duplicate", nil)

	_, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Another Org", Slug: "duplicate"})

	assert.ErrorIs(suite.T(), err, ErrDuplicateOrgSlug)
	assert.Equal(suite.T(), []string{"duplicate"}, suite.keycloak.created, "Keycloak must not be touched for the duplicate")
	assert.Equal(suite.T(), 1, suite.countOrgs("duplicate"))
}

func (suite *OrgServiceTestSuite) TestCreateOrgRejectsUnknownParent() {
	unknown := uuid.New()

	_, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{ParentOrgID: &unknown, Name: "Orphan", Slug: "orphan"})

	assert.ErrorIs(suite.T(), err, ErrParentOrgNotFound)
	assert.Empty(suite.T(), suite.keycloak.created)
	assert.Zero(suite.T(), suite.countOrgs("orphan"))
}

func (suite *OrgServiceTestSuite) TestCreateOrgCleansUpKeycloakWhileTheSlugIsStillHeld() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	slugHeld := false
	suite.keycloak.onDelete = func(slug string) {
		slugHeld = suite.insertBlocksOnSlug(slug)
	}

	_, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Held Org", Slug: "held"})

	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Equal(suite.T(), []string{"held"}, suite.keycloak.deleted)
	assert.True(suite.T(), slugHeld, "the cleanup must run before the rollback releases the slug")
	assert.Zero(suite.T(), suite.countOrgs("held"))
}

func (suite *OrgServiceTestSuite) TestCreateOrgReportsTheProvisioningErrorWhenTheCleanupFails() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	suite.keycloak.deleteErr = errors.New("cleanup failed")

	_, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Broken Org", Slug: "broken"})

	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Zero(suite.T(), suite.countOrgs("broken"))
}

func (suite *OrgServiceTestSuite) TestCreateOrgSucceedsOnRetryAfterKeycloakFailure() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	_, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Retried Org", Slug: "retried"})
	require.Error(suite.T(), err)

	suite.keycloak.createErr = nil
	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Retried Org", Slug: "retried"})

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "retried", created.Slug)
	assert.Equal(suite.T(), []string{"retried", "retried"}, suite.keycloak.created)
}

func (suite *OrgServiceTestSuite) TestDatabaseRejectsInvalidSlugs() {
	for _, slug := range []string{"a", "ASE", "ase--lab", "-ase", "cg", "ase-cg", "cg-lab", "ase-cg-lab"} {
		_, err := suite.conn.Exec(suite.ctx, "INSERT INTO org (name, slug) VALUES ('Invalid', $1)", slug)

		var pgErr *pgconn.PgError
		if assert.ErrorAs(suite.T(), err, &pgErr, slug) {
			assert.Equal(suite.T(), "check_org_slug_format", pgErr.ConstraintName, slug)
		}
	}
}

func (suite *OrgServiceTestSuite) TestDatabaseRejectsAnOrgAsItsOwnParent() {
	org := suite.createOrg("own-parent", nil)

	_, err := suite.conn.Exec(suite.ctx, "UPDATE org SET parent_org_id = id WHERE id = $1", org.ID)

	var pgErr *pgconn.PgError
	if assert.ErrorAs(suite.T(), err, &pgErr) {
		assert.Equal(suite.T(), "check_org_not_own_parent", pgErr.ConstraintName)
	}
}

func (suite *OrgServiceTestSuite) TestGetOrgNotFound() {
	_, err := suite.service.GetOrg(suite.ctx, uuid.New())

	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
}

func (suite *OrgServiceTestSuite) TestGetAllOrgsIsSortedByName() {
	suite.createOrg("zeta", nil)
	suite.createOrg("alpha", nil)

	orgs, err := suite.service.GetAllOrgs(suite.ctx)

	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(orgs), 2)
	for i := 1; i < len(orgs); i++ {
		assert.LessOrEqual(suite.T(), orgs[i-1].Name, orgs[i].Name)
	}
}

func (suite *OrgServiceTestSuite) TestUpdateOrgReplacesNameAndOptionalFields() {
	parent := suite.createOrg("update-parent", nil)
	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{
		ParentOrgID:  &parent.ID,
		Name:         "Before",
		Slug:         "update-me",
		School:       "School",
		University:   "University",
		Website:      "https://before.example.com",
		ContactEmail: "before@example.com",
	})
	require.NoError(suite.T(), err)

	updated, err := suite.service.UpdateOrg(suite.ctx, created.ID, orgDTO.UpdateOrg{
		Name:    " After ",
		Website: "https://after.example.com",
	})

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "After", updated.Name)
	assert.Equal(suite.T(), "https://after.example.com", updated.Website.String)
	assert.False(suite.T(), updated.School.Valid)
	assert.False(suite.T(), updated.University.Valid)
	assert.False(suite.T(), updated.ContactEmail.Valid)
	assert.Equal(suite.T(), "update-me", updated.Slug, "the slug is immutable")
	assert.Equal(suite.T(), &parent.ID, updated.ParentOrgID, "the parent is changed through UpdateOrgParent only")
}

func (suite *OrgServiceTestSuite) TestUpdateOrgNotFound() {
	_, err := suite.service.UpdateOrg(suite.ctx, uuid.New(), orgDTO.UpdateOrg{Name: "Nobody"})

	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
}

func (suite *OrgServiceTestSuite) TestUpdateOrgParentMovesAndDetaches() {
	parent := suite.createOrg("move-parent", nil)
	org := suite.createOrg("move-child", nil)

	moved, err := suite.service.UpdateOrgParent(suite.ctx, org.ID, orgDTO.UpdateOrgParent{ParentOrgID: &parent.ID})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), &parent.ID, moved.ParentOrgID)

	detached, err := suite.service.UpdateOrgParent(suite.ctx, org.ID, orgDTO.UpdateOrgParent{ParentOrgID: nil})
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), detached.ParentOrgID)
}

func (suite *OrgServiceTestSuite) TestUpdateOrgParentRejectsCycles() {
	root := suite.createOrg("cycle-root", nil)
	middle := suite.createOrg("cycle-middle", &root.ID)
	leaf := suite.createOrg("cycle-leaf", &middle.ID)

	for name, parentID := range map[string]uuid.UUID{"itself": root.ID, "its child": middle.ID, "a deeper descendant": leaf.ID} {
		_, err := suite.service.UpdateOrgParent(suite.ctx, root.ID, orgDTO.UpdateOrgParent{ParentOrgID: &parentID})
		assert.ErrorIs(suite.T(), err, ErrOrgCycle, name)
	}
	assert.Nil(suite.T(), suite.parentOf(root.ID))
}

func (suite *OrgServiceTestSuite) TestUpdateOrgParentRejectsUnknownParent() {
	org := suite.createOrg("unknown-parent", nil)
	unknown := uuid.New()

	_, err := suite.service.UpdateOrgParent(suite.ctx, org.ID, orgDTO.UpdateOrgParent{ParentOrgID: &unknown})

	assert.ErrorIs(suite.T(), err, ErrParentOrgNotFound)
}

func (suite *OrgServiceTestSuite) TestUpdateOrgParentNotFound() {
	_, err := suite.service.UpdateOrgParent(suite.ctx, uuid.New(), orgDTO.UpdateOrgParent{ParentOrgID: nil})

	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
}

// Without the hierarchy lock both updates could pass their cycle check before either
// commits, leaving the two orgs as each other's parent.
func (suite *OrgServiceTestSuite) TestConcurrentParentUpdatesCannotCreateACycle() {
	for i := range 10 {
		first := suite.createOrg(fmt.Sprintf("race-first-%d", i), nil)
		second := suite.createOrg(fmt.Sprintf("race-second-%d", i), nil)

		start := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup
		wg.Go(func() {
			<-start
			_, errs[0] = suite.service.UpdateOrgParent(suite.ctx, first.ID, orgDTO.UpdateOrgParent{ParentOrgID: &second.ID})
		})
		wg.Go(func() {
			<-start
			_, errs[1] = suite.service.UpdateOrgParent(suite.ctx, second.ID, orgDTO.UpdateOrgParent{ParentOrgID: &first.ID})
		})
		close(start)
		wg.Wait()

		succeeded := 0
		for _, err := range errs {
			if err == nil {
				succeeded++
			} else {
				assert.ErrorIs(suite.T(), err, ErrOrgCycle)
			}
		}
		assert.Equal(suite.T(), 1, succeeded, "exactly one of the two moves may succeed")
	}
}

func (suite *OrgServiceTestSuite) TestDeleteOrgRemovesTheOrgAndItsKeycloakArtifacts() {
	org := suite.createOrg("delete-me", nil)

	err := suite.service.DeleteOrg(suite.ctx, org.ID)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"delete-me"}, suite.keycloak.deleted)
	_, err = suite.service.GetOrg(suite.ctx, org.ID)
	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
}

func (suite *OrgServiceTestSuite) TestDeleteOrgWithChildOrgsIsRejectedBeforeKeycloak() {
	parent := suite.createOrg("delete-parent", nil)
	suite.createOrg("delete-parent-child", &parent.ID)

	err := suite.service.DeleteOrg(suite.ctx, parent.ID)

	assert.ErrorIs(suite.T(), err, ErrOrgHasChildOrgs)
	assert.Empty(suite.T(), suite.keycloak.deleted)
	assert.Equal(suite.T(), 1, suite.countOrgs("delete-parent"))
}

func (suite *OrgServiceTestSuite) TestDeleteOrgWithCoursesIsRejectedBeforeKeycloak() {
	org := suite.createOrg("delete-with-course", nil)
	_, err := suite.conn.Exec(suite.ctx, `
		INSERT INTO course (id, name, semester_tag, course_type, start_date, end_date, org_id)
		VALUES ($1, 'Org Course', 'ws25', 'lecture', '2025-10-01', '2026-02-01', $2)`,
		uuid.New(), org.ID)
	require.NoError(suite.T(), err)

	err = suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorIs(suite.T(), err, ErrOrgHasCourses)
	assert.Empty(suite.T(), suite.keycloak.deleted)
	assert.Equal(suite.T(), 1, suite.countOrgs("delete-with-course"))
}

func (suite *OrgServiceTestSuite) TestDeleteOrgKeepsTheOrgWhenKeycloakFails() {
	org := suite.createOrg("delete-fails", nil)
	suite.keycloak.deleteErr = errors.New("keycloak unavailable")

	err := suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorIs(suite.T(), err, suite.keycloak.deleteErr)
	assert.Equal(suite.T(), 1, suite.countOrgs("delete-fails"))
	assert.Equal(suite.T(), []string{"delete-fails", "delete-fails"}, suite.keycloak.created, "the failed deletion is followed by a restore")
}

func (suite *OrgServiceTestSuite) TestDeleteOrgNotFound() {
	err := suite.service.DeleteOrg(suite.ctx, uuid.New())

	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
	assert.Empty(suite.T(), suite.keycloak.deleted)
}

func TestOrgServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrgServiceTestSuite))
}
