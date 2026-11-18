package org

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type keycloakSpy struct {
	created           []string
	deleted           []string
	added             []orgMembers
	members           orgMembers
	createErr         error
	deleteErr         error
	existErr          error
	addErr            error
	deletedDespiteErr bool
	onCreate          func(slug string)
	onDelete          func(slug string)
}

func (k *keycloakSpy) create(_ context.Context, slug string) error {
	k.created = append(k.created, slug)
	if k.onCreate != nil {
		k.onCreate(slug)
	}
	return k.createErr
}

func (k *keycloakSpy) delete(_ context.Context, slug string) (orgMembers, error) {
	k.deleted = append(k.deleted, slug)
	if k.onDelete != nil {
		k.onDelete(slug)
	}
	return k.members, k.deleteErr
}

func (k *keycloakSpy) exist(_ context.Context, _ string) (bool, error) {
	return !k.deletedDespiteErr, k.existErr
}

func (k *keycloakSpy) add(_ context.Context, _ string, members orgMembers) error {
	k.added = append(k.added, members)
	return k.addErr
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
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.keycloak = &keycloakSpy{}
	directory := newFakeMemberDirectory()
	suite.service = NewOrgService(*testDB.Queries, testDB.Conn, suite.keycloak.create, suite.keycloak.delete, suite.keycloak.exist, suite.keycloak.add, directory)
}

// withMembers returns the suite's service with its own member directory holding the users.
func (suite *OrgServiceTestSuite) withMembers(users ...keycloakRealmDTO.StaffMember) (*OrgService, *fakeMemberDirectory) {
	members := newFakeMemberDirectory(users...)
	service := *suite.service
	service.members = members
	return &service, members
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

func (suite *OrgServiceTestSuite) failCommitOn(event, row, slug string) {
	_, err := suite.conn.Exec(suite.ctx, fmt.Sprintf(`
		CREATE FUNCTION fail_commit() RETURNS trigger LANGUAGE plpgsql AS $fn$
		BEGIN RAISE EXCEPTION 'commit rejected'; END $fn$;
		CREATE CONSTRAINT TRIGGER fail_commit AFTER %s ON org DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW WHEN (%s.slug = '%s') EXECUTE FUNCTION fail_commit();`, event, row, slug))
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() {
		_, err := suite.conn.Exec(suite.ctx, "DROP TRIGGER fail_commit ON org; DROP FUNCTION fail_commit();")
		require.NoError(suite.T(), err)
	})
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

func (suite *OrgServiceTestSuite) TestCreateOrgTakesTheCreationTimeAfterProvisioningKeycloak() {
	var provisionedAt time.Time
	suite.keycloak.onCreate = func(string) {
		err := suite.conn.QueryRow(suite.ctx, "SELECT clock_timestamp()").Scan(&provisionedAt)
		require.NoError(suite.T(), err)
	}

	created, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Stamped", Slug: "stamped"})
	require.NoError(suite.T(), err)

	assert.True(suite.T(), created.CreatedAt.After(provisionedAt), "created at %v, provisioned at %v", created.CreatedAt, provisionedAt)
	stored, err := suite.service.GetOrg(suite.ctx, created.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), created.CreatedAt, stored.CreatedAt)
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

func (suite *OrgServiceTestSuite) TestCreateOrgCleansUpKeycloakWhenTheCommitFails() {
	suite.failCommitOn("INSERT", "NEW", "commit-fails")

	_, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Commit Fails", Slug: "commit-fails"})

	assert.ErrorContains(suite.T(), err, "commit rejected")
	assert.Zero(suite.T(), suite.countOrgs("commit-fails"))
	assert.Equal(suite.T(), []string{"commit-fails"}, suite.keycloak.deleted)
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
	rows, err := suite.conn.Query(suite.ctx, "SELECT id FROM org ORDER BY name, slug")
	require.NoError(suite.T(), err)
	expected, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	require.NoError(suite.T(), err)
	actual := make([]uuid.UUID, 0, len(orgs))
	for _, org := range orgs {
		actual = append(actual, org.ID)
	}
	assert.Equal(suite.T(), expected, actual)
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

func (suite *OrgServiceTestSuite) TestMovingAnOrgToTheTopLevelTakesNoHierarchyLock() {
	parent := suite.createOrg("detach-parent", nil)
	org := suite.createOrg("detach-child", &parent.ID)
	tx, err := suite.conn.Begin(suite.ctx)
	require.NoError(suite.T(), err)
	defer func() { _ = tx.Rollback(suite.ctx) }()
	require.NoError(suite.T(), db.New(tx).LockOrgHierarchy(suite.ctx))

	ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel()
	detached, err := suite.service.UpdateOrgParent(ctx, org.ID, orgDTO.UpdateOrgParent{ParentOrgID: nil})

	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), detached.ParentOrgID)
}

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

	slug, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "delete-me", slug)
	assert.Equal(suite.T(), []string{"delete-me"}, suite.keycloak.deleted)
	_, err = suite.service.GetOrg(suite.ctx, org.ID)
	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
}

func (suite *OrgServiceTestSuite) TestDeleteOrgWithChildOrgsIsRejectedBeforeKeycloak() {
	parent := suite.createOrg("delete-parent", nil)
	suite.createOrg("delete-parent-child", &parent.ID)

	_, err := suite.service.DeleteOrg(suite.ctx, parent.ID)

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

	_, err = suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorIs(suite.T(), err, ErrOrgHasCourses)
	assert.Empty(suite.T(), suite.keycloak.deleted)
	assert.Equal(suite.T(), 1, suite.countOrgs("delete-with-course"))
}

func (suite *OrgServiceTestSuite) TestDeleteOrgKeepsTheOrgWhenKeycloakFails() {
	org := suite.createOrg("delete-fails", nil)
	suite.keycloak.deleteErr = errors.New("keycloak unavailable")

	_, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorIs(suite.T(), err, suite.keycloak.deleteErr)
	assert.Equal(suite.T(), 1, suite.countOrgs("delete-fails"))
	assert.Equal(suite.T(), []string{"delete-fails", "delete-fails"}, suite.keycloak.created, "the failed deletion is followed by a restore")
}

func (suite *OrgServiceTestSuite) TestDeleteOrgReportsAFailedRestore() {
	org := suite.createOrg("restore-fails", nil)
	suite.keycloak.deleteErr = errors.New("keycloak unavailable")
	suite.keycloak.createErr = errors.New("restore failed")

	_, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorIs(suite.T(), err, suite.keycloak.deleteErr)
	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Equal(suite.T(), 1, suite.countOrgs("restore-fails"))
}

func (suite *OrgServiceTestSuite) TestDeleteOrgCompletesWhenKeycloakDeletedEverythingDespiteAnError() {
	org := suite.createOrg("deleted-despite-error", nil)
	suite.keycloak.deleteErr = errors.New("response lost")
	suite.keycloak.deletedDespiteErr = true

	slug, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "deleted-despite-error", slug)
	assert.Zero(suite.T(), suite.countOrgs("deleted-despite-error"))
	assert.Equal(suite.T(), []string{"deleted-despite-error"}, suite.keycloak.created, "nothing is restored")
}

func (suite *OrgServiceTestSuite) TestDeleteOrgKeepsTheOrgWithoutRestoringWhenTheKeycloakCheckFails() {
	org := suite.createOrg("check-fails", nil)
	suite.keycloak.deleteErr = errors.New("keycloak unavailable")
	suite.keycloak.existErr = errors.New("check failed")

	_, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorIs(suite.T(), err, suite.keycloak.deleteErr)
	assert.ErrorIs(suite.T(), err, suite.keycloak.existErr)
	assert.Equal(suite.T(), 1, suite.countOrgs("check-fails"))
	assert.Equal(suite.T(), []string{"check-fails"}, suite.keycloak.created, "nothing is restored")
}

func (suite *OrgServiceTestSuite) TestDeleteOrgRestoresKeycloakAndItsMembersWhenTheCommitFails() {
	org := suite.createOrg("delete-commit-fails", nil)
	suite.failCommitOn("DELETE", "OLD", "delete-commit-fails")
	suite.keycloak.members = orgMembers{permissionValidation.OrgAdmin: {"admin-1"}, permissionValidation.OrgMember: {"member-1", "member-2"}}

	_, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorContains(suite.T(), err, "commit rejected")
	assert.NotErrorIs(suite.T(), err, ErrOrgMembersRemoved, "every member is back")
	assert.Equal(suite.T(), 1, suite.countOrgs("delete-commit-fails"))
	assert.Equal(suite.T(), []string{"delete-commit-fails"}, suite.keycloak.deleted)
	assert.Equal(suite.T(), []string{"delete-commit-fails", "delete-commit-fails"}, suite.keycloak.created, "the failed commit is followed by a restore")
	assert.Equal(suite.T(), []orgMembers{suite.keycloak.members}, suite.keycloak.added, "the members read before the deletion are re-added")
}

func (suite *OrgServiceTestSuite) TestDeleteOrgReportsMembersThatCouldNotBeReAdded() {
	org := suite.createOrg("re-add-fails", nil)
	suite.failCommitOn("DELETE", "OLD", "re-add-fails")
	suite.keycloak.members = orgMembers{permissionValidation.OrgAdmin: {"admin-1"}}
	suite.keycloak.addErr = errors.New("failed to re-add user admin-1")

	_, err := suite.service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorContains(suite.T(), err, "commit rejected")
	assert.ErrorIs(suite.T(), err, ErrOrgMembersRemoved)
	assert.ErrorIs(suite.T(), err, suite.keycloak.addErr)
	assert.Equal(suite.T(), 1, suite.countOrgs("re-add-fails"))
}

func (suite *OrgServiceTestSuite) TestDeleteOrgNotFound() {
	_, err := suite.service.DeleteOrg(suite.ctx, uuid.New())

	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
	assert.Empty(suite.T(), suite.keycloak.deleted)
}

func (suite *OrgServiceTestSuite) TestGetOwnOrgsReturnsTheTokensOrgsWithTheStrongerRole() {
	adminOrg := suite.createOrg("own-admin", nil)
	memberOrg := suite.createOrg("own-member", nil)
	suite.createOrg("own-unrelated", nil)

	ownOrgs, err := suite.service.GetOwnOrgs(suite.ctx, map[string]bool{
		"org-own-admin-Member":           true,
		"org-own-admin-Admin":            true,
		"org-own-member-Member":          true,
		"org-own-unrelated-Admin":        false,
		"org-own-unknown-Admin":          true,
		"org-own-member-cg-Admin":        true,
		"ws25-own-admin-Lecturer":        true,
		permissionValidation.PromptAdmin: true,
	}, time.Now().Add(time.Minute))

	require.NoError(suite.T(), err)
	require.Len(suite.T(), ownOrgs, 2)
	assert.Equal(suite.T(), adminOrg.ID, ownOrgs[0].ID)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, ownOrgs[0].Role)
	assert.Equal(suite.T(), memberOrg.ID, ownOrgs[1].ID)
	assert.Equal(suite.T(), "own-member", ownOrgs[1].Slug)
	assert.Equal(suite.T(), permissionValidation.OrgMember, ownOrgs[1].Role)
}

func (suite *OrgServiceTestSuite) TestGetOwnOrgsLeavesOutAnOrgCreatedAfterTheToken() {
	org := suite.createOrg("own-later", nil)

	ownOrgs, err := suite.service.GetOwnOrgs(suite.ctx, map[string]bool{"org-own-later-Admin": true}, org.CreatedAt.Add(-time.Minute))

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), ownOrgs)
}

func (suite *OrgServiceTestSuite) TestGetOwnOrgsWithoutOrgRolesIsAnEmptyList() {
	ownOrgs, err := suite.service.GetOwnOrgs(suite.ctx, map[string]bool{"ws25-ase-Lecturer": true}, time.Now())

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ownOrgs)
	assert.Empty(suite.T(), ownOrgs)
}

var caller = Caller{UserID: "caller"}

// A member changed from a stale list may have been removed since, so changing their role never adds
// them back.
func (suite *OrgServiceTestSuite) TestUpdateOrgMemberRoleNeverAddsAUserToTheOrg() {
	org := suite.createOrg("members-update-removed", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))

	_, err := service.UpdateOrgMemberRole(suite.ctx, org.ID, "ada", permissionValidation.OrgAdmin, caller)

	assert.ErrorIs(suite.T(), err, ErrNotOrgMember)
	assert.Empty(suite.T(), members.roleOf("members-update-removed", "ada"))
}

func (suite *OrgServiceTestSuite) TestMembershipOfAnUnknownOrgIsNotFound() {
	service, _ := suite.withMembers(staffMember("ada", "ad12min"))

	_, err := service.GetOrgMembers(suite.ctx, uuid.New())
	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
	_, err = service.AddOrgMember(suite.ctx, uuid.New(), "ada", permissionValidation.OrgMember, caller)
	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
	_, err = service.UpdateOrgMemberRole(suite.ctx, uuid.New(), "ada", permissionValidation.OrgMember, caller)
	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
	_, err = service.RemoveOrgMember(suite.ctx, uuid.New(), "ada", caller)
	assert.ErrorIs(suite.T(), err, ErrOrgNotFound)
}

func (suite *OrgServiceTestSuite) TestUpdateOrgMemberRoleChangesTheRoleOfExactlyThatUser() {
	org := suite.createOrg("members-role", nil)
	// max has no university login, so the table shows his username, which is ada's university login
	ada := staffMember("ada", "ad12min")
	max := keycloakRealmDTO.StaffMember{KeycloakUserID: "max", Username: "ad12min"}
	service, members := suite.withMembers(ada, max)
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-role", "max", permissionValidation.OrgMember))

	gotOrg, err := service.UpdateOrgMemberRole(suite.ctx, org.ID, "max", permissionValidation.OrgAdmin, caller)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), org.ID, gotOrg.ID)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, members.roleOf("members-role", "max"))
	assert.Empty(suite.T(), members.roleOf("members-role", "ada"))
}

func (suite *OrgServiceTestSuite) TestUpdateOrgMemberRoleRejectsTheCallersOwnMembership() {
	org := suite.createOrg("members-self-role", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-self-role", "ada", permissionValidation.OrgAdmin))

	_, err := service.UpdateOrgMemberRole(suite.ctx, org.ID, "ada", permissionValidation.OrgMember, Caller{UserID: "ada"})

	assert.ErrorIs(suite.T(), err, ErrOwnMembership)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, members.roleOf("members-self-role", "ada"))
}

func (suite *OrgServiceTestSuite) TestUpdateOrgMemberRoleOfAnUnknownUser() {
	org := suite.createOrg("members-ghost", nil)
	service, _ := suite.withMembers()

	_, err := service.UpdateOrgMemberRole(suite.ctx, org.ID, "ghost", permissionValidation.OrgMember, caller)

	assert.ErrorIs(suite.T(), err, keycloakRealmManager.ErrUserNotFound)
}

func (suite *OrgServiceTestSuite) TestRemoveOrgMemberTakesTheUserOutOfTheOrg() {
	org := suite.createOrg("members-remove", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-remove", "ada", permissionValidation.OrgMember))

	gotOrg, err := service.RemoveOrgMember(suite.ctx, org.ID, "ada", caller)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), org.ID, gotOrg.ID)
	assert.Empty(suite.T(), members.roleOf("members-remove", "ada"))
}

func (suite *OrgServiceTestSuite) TestRemoveOrgMemberRejectsTheCallersOwnMembership() {
	org := suite.createOrg("members-leave", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-leave", "ada", permissionValidation.OrgAdmin))

	_, err := service.RemoveOrgMember(suite.ctx, org.ID, "ada", Caller{UserID: "ada"})

	assert.ErrorIs(suite.T(), err, ErrOwnMembership)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, members.roleOf("members-leave", "ada"))
}

func (suite *OrgServiceTestSuite) TestAddOrgMemberAddsAUserWithoutAnOrgRole() {
	org := suite.createOrg("members-add-new", nil)
	ada := staffMember("ada", "ad12min")
	service, members := suite.withMembers(ada)

	gotOrg, err := service.AddOrgMember(suite.ctx, org.ID, "ada", permissionValidation.OrgMember, caller)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), org.ID, gotOrg.ID)
	assert.Equal(suite.T(), permissionValidation.OrgMember, members.roleOf("members-add-new", "ada"))
	listed, err := service.GetOrgMembers(suite.ctx, org.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []keycloakRealmDTO.OrgMember{{StaffMember: ada, Role: permissionValidation.OrgMember}}, listed)
}

// A user picked from a stale list may have become a member since, so adding them never changes
// their role.
func (suite *OrgServiceTestSuite) TestAddOrgMemberNeverChangesTheRoleOfAMember() {
	org := suite.createOrg("members-add-existing", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-add-existing", "ada", permissionValidation.OrgAdmin))

	_, err := service.AddOrgMember(suite.ctx, org.ID, "ada", permissionValidation.OrgMember, caller)

	assert.ErrorIs(suite.T(), err, ErrAlreadyOrgMember)
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, members.roleOf("members-add-existing", "ada"))
}

func (suite *OrgServiceTestSuite) TestAddOrgMemberOfAnUnknownUser() {
	org := suite.createOrg("members-add-ghost", nil)
	service, _ := suite.withMembers()

	_, err := service.AddOrgMember(suite.ctx, org.ID, "ghost", permissionValidation.OrgMember, caller)

	assert.ErrorIs(suite.T(), err, keycloakRealmManager.ErrUserNotFound)
}

// awaitSignal waits until a goroutine signals, and fails the test when it returns first or never
// signals, so a broken lock fails the test instead of hanging it.
func (suite *OrgServiceTestSuite) awaitSignal(signal <-chan struct{}, done <-chan error, what string) {
	select {
	case <-signal:
	case err := <-done:
		suite.T().Fatalf("%s returned before it got there: %v", what, err)
	case <-time.After(5 * time.Second):
		suite.T().Fatalf("%s never got there", what)
	}
}

// assertWaiting fails the test when the goroutine returns within a moment.
func (suite *OrgServiceTestSuite) assertWaiting(done <-chan error, what string) {
	select {
	case err := <-done:
		suite.T().Fatalf("%s did not wait: %v", what, err)
	case <-time.After(200 * time.Millisecond):
	}
}

// receive returns the result of a goroutine, and fails the test when it never returns.
func (suite *OrgServiceTestSuite) receive(done <-chan error, what string) error {
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		suite.T().Fatalf("%s never returned", what)
		return nil
	}
}

func (suite *OrgServiceTestSuite) TestMembershipChangesAreRefusedWhileTheOrgIsDeleted() {
	org := suite.createOrg("members-deleting", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	var changeErr error
	suite.keycloak.onDelete = func(string) {
		_, changeErr = service.AddOrgMember(suite.ctx, org.ID, "ada", permissionValidation.OrgMember, caller)
	}

	_, err := service.DeleteOrg(suite.ctx, org.ID)

	require.NoError(suite.T(), err)
	assert.ErrorIs(suite.T(), changeErr, ErrMembershipBusy)
	assert.Empty(suite.T(), members.roleOf("members-deleting", "ada"))
}

// A deletion re-adds the members it found when it fails, so it must not find them while a change
// is still under way in Keycloak.
func (suite *OrgServiceTestSuite) TestDeletionWaitsForAMembershipChangeInKeycloak() {
	org := suite.createOrg("members-changing", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-changing", "ada", permissionValidation.OrgMember))
	inKeycloak, release := make(chan struct{}), make(chan struct{})
	releaseChange := sync.OnceFunc(func() { close(release) })
	defer releaseChange()
	members.beforeWrite = func() {
		close(inKeycloak)
		<-release
	}

	changed := make(chan error, 1)
	go func() {
		_, err := service.RemoveOrgMember(suite.ctx, org.ID, "ada", caller)
		changed <- err
	}()
	suite.awaitSignal(inKeycloak, changed, "the membership change")
	deleted := make(chan error, 1)
	go func() {
		_, err := service.DeleteOrg(suite.ctx, org.ID)
		deleted <- err
	}()
	suite.assertWaiting(deleted, "the deletion")
	releaseChange()

	require.NoError(suite.T(), suite.receive(changed, "the membership change"))
	require.NoError(suite.T(), suite.receive(deleted, "the deletion"))
	assert.Equal(suite.T(), []string{"members-changing"}, suite.keycloak.deleted)
}

// A failed commit releases the deletion's own locks before it re-adds the members it found, so a
// change made in between would be undone by that restore.
func (suite *OrgServiceTestSuite) TestMembershipChangesAreRefusedDuringTheRestoreOfAFailedDeletion() {
	org := suite.createOrg("members-restoring", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-restoring", "ada", permissionValidation.OrgMember))
	suite.keycloak.members = orgMembers{permissionValidation.OrgMember: {"ada"}}
	suite.failCommitOn("DELETE", "OLD", "members-restoring")
	removeAda := func() error {
		_, err := service.RemoveOrgMember(suite.ctx, org.ID, "ada", caller)
		return err
	}
	var duringDeletion, duringRestore error
	suite.keycloak.onDelete = func(string) { duringDeletion = removeAda() }
	// The restore recreates the groups before it re-adds the members
	suite.keycloak.onCreate = func(string) { duringRestore = removeAda() }

	_, err := service.DeleteOrg(suite.ctx, org.ID)

	assert.ErrorContains(suite.T(), err, "commit rejected")
	assert.Len(suite.T(), suite.keycloak.added, 1, "the restore re-added the members")
	assert.ErrorIs(suite.T(), duringDeletion, ErrMembershipBusy)
	assert.ErrorIs(suite.T(), duringRestore, ErrMembershipBusy)
	assert.Equal(suite.T(), permissionValidation.OrgMember, members.roleOf("members-restoring", "ada"))
	require.NoError(suite.T(), removeAda(), "the restore released the lock")
	assert.Empty(suite.T(), members.roleOf("members-restoring", "ada"))
}

// A role change adds the new role before it removes the other one, so two overlapping changes of a
// user could leave them with neither. A change that waited for the other one would hold a pool
// connection meanwhile, so it is refused instead.
func (suite *OrgServiceTestSuite) TestAMembershipChangeIsRefusedWhileAnotherOneOfTheOrgIsUnderWay() {
	org := suite.createOrg("members-turns", nil)
	otherOrg := suite.createOrg("members-turns-other", nil)
	service, members := suite.withMembers(staffMember("ada", "ad12min"), staffMember("max", "ma12xim"))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-turns", "ada", permissionValidation.OrgMember))
	require.NoError(suite.T(), members.SetOrgMemberRole(suite.ctx, "members-turns-other", "ada", permissionValidation.OrgMember))
	writes, release := make(chan struct{}, 1), make(chan struct{})
	releaseWrites := sync.OnceFunc(func() { close(release) })
	defer releaseWrites()
	members.beforeWrite = func() {
		writes <- struct{}{}
		<-release
	}

	first := make(chan error, 1)
	go func() {
		_, err := service.UpdateOrgMemberRole(suite.ctx, org.ID, "ada", permissionValidation.OrgAdmin, caller)
		first <- err
	}()
	suite.awaitSignal(writes, first, "the first membership change")
	members.beforeWrite = nil

	_, err := service.UpdateOrgMemberRole(suite.ctx, org.ID, "max", permissionValidation.OrgAdmin, caller)
	assert.ErrorIs(suite.T(), err, ErrMembershipBusy, "a change of the same org")
	_, err = service.UpdateOrgMemberRole(suite.ctx, otherOrg.ID, "ada", permissionValidation.OrgAdmin, caller)
	assert.NoError(suite.T(), err, "a change of another org")
	releaseWrites()

	require.NoError(suite.T(), suite.receive(first, "the first membership change"))
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, members.roleOf("members-turns", "ada"))
	assert.Empty(suite.T(), members.roleOf("members-turns", "max"))
	assert.Equal(suite.T(), permissionValidation.OrgAdmin, members.roleOf("members-turns-other", "ada"))
}

// Other requests may hold every other connection of the pool, so a deletion that needs a second
// connection while it holds the lock would keep the org's membership locked until they let one go.
func (suite *OrgServiceTestSuite) TestDeleteOrgNeedsOneConnection() {
	config, err := pgxpool.ParseConfig(suite.conn.Config().ConnString())
	require.NoError(suite.T(), err)
	config.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(suite.ctx, config)
	require.NoError(suite.T(), err)
	defer pool.Close()
	service := *suite.service
	service.conn = pool
	service.queries = *db.New(pool)

	deleted := suite.createOrg("one-connection", nil)
	restored := suite.createOrg("one-connection-restored", nil)
	suite.failCommitOn("DELETE", "OLD", "one-connection-restored")
	ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel()

	slug, err := service.DeleteOrg(ctx, deleted.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "one-connection", slug)
	_, err = service.DeleteOrg(ctx, restored.ID)
	assert.ErrorContains(suite.T(), err, "commit rejected")
	assert.Equal(suite.T(), 1, suite.countOrgs("one-connection-restored"))
	// The lock is released: a membership change on another connection goes ahead
	other, _ := suite.withMembers()
	_, err = other.UpdateOrgMemberRole(ctx, restored.ID, "nobody", permissionValidation.OrgMember, caller)
	assert.ErrorIs(suite.T(), err, keycloakRealmManager.ErrUserNotFound)
}

func TestOrgServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrgServiceTestSuite))
}
