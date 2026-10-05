package org

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/org/orgDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// chainDepth is far beyond the realistic three to four levels, so that the recursive
// ancestor walk, which every cycle check runs, stays fast on pathological hierarchies.
const chainDepth = 10_000

// hierarchyTimeout guards against a hanging query on the deep chain; it is generous so
// slow CI runners do not turn it into a flaky test.
const hierarchyTimeout = 30 * time.Second

type OrgHierarchyTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	conn    *pgxpool.Pool
	queries *db.Queries
	service *OrgService
	rootID  uuid.UUID
	leafID  uuid.UUID
}

func (suite *OrgHierarchyTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := setupMigratedTestDB(suite.ctx)
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.queries = testDB.Queries
	noKeycloak := func(context.Context, string) error { return nil }
	suite.service = NewOrgService(*testDB.Queries, testDB.Conn, noKeycloak, noKeycloak)

	// chain-1 is the root and chain-<chainDepth> the leaf; each level is the parent of
	// the next. Foreign keys are checked at the end of the statement, so the order in
	// which the rows are inserted does not matter.
	_, err = suite.conn.Exec(suite.ctx, `
		INSERT INTO org (id, parent_org_id, name, slug)
		SELECT md5('chain-' || i)::uuid,
		       CASE WHEN i = 1 THEN NULL ELSE md5('chain-' || (i - 1))::uuid END,
		       'Chain ' || i,
		       'chain-' || i
		FROM generate_series(1, $1::int) AS i`, chainDepth)
	if err != nil {
		log.Fatalf("Failed to seed the org chain: %v", err)
	}
	err = suite.conn.QueryRow(suite.ctx, "SELECT id FROM org WHERE slug = 'chain-1'").Scan(&suite.rootID)
	if err != nil {
		log.Fatalf("Failed to read the chain root: %v", err)
	}
	err = suite.conn.QueryRow(suite.ctx, "SELECT id FROM org WHERE slug = $1", "chain-10000").Scan(&suite.leafID)
	if err != nil {
		log.Fatalf("Failed to read the chain leaf: %v", err)
	}
}

func (suite *OrgHierarchyTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *OrgHierarchyTestSuite) TestAncestorsOfTheDeepestOrg() {
	ctx, cancel := context.WithTimeout(suite.ctx, hierarchyTimeout)
	defer cancel()

	ancestorIDs, err := suite.queries.GetOrgAncestorIDs(ctx, suite.leafID)

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), ancestorIDs, chainDepth-1)
	assert.Contains(suite.T(), ancestorIDs, suite.rootID)
	assert.NotContains(suite.T(), ancestorIDs, suite.leafID)
}

func (suite *OrgHierarchyTestSuite) TestAncestorsOfTheRoot() {
	ancestorIDs, err := suite.queries.GetOrgAncestorIDs(suite.ctx, suite.rootID)

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), ancestorIDs)
}

func (suite *OrgHierarchyTestSuite) TestMovingTheRootUnderTheDeepestOrgIsACycle() {
	ctx, cancel := context.WithTimeout(suite.ctx, hierarchyTimeout)
	defer cancel()

	_, err := suite.service.UpdateOrgParent(ctx, suite.rootID, orgDTO.UpdateOrgParent{ParentOrgID: &suite.leafID})

	assert.ErrorIs(suite.T(), err, ErrOrgCycle)
}

func (suite *OrgHierarchyTestSuite) TestAnOrgCanBeMovedBelowTheDeepestOrg() {
	ctx, cancel := context.WithTimeout(suite.ctx, hierarchyTimeout)
	defer cancel()
	org, err := suite.service.CreateOrg(ctx, orgDTO.CreateOrg{Name: "Below the chain", Slug: "below-chain"})
	require.NoError(suite.T(), err)

	moved, err := suite.service.UpdateOrgParent(ctx, org.ID, orgDTO.UpdateOrgParent{ParentOrgID: &suite.leafID})

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), &suite.leafID, moved.ParentOrgID)
	ancestorIDs, err := suite.queries.GetOrgAncestorIDs(ctx, org.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), ancestorIDs, chainDepth)
}

// The service never creates a cycle, but the ancestor walk must still terminate if the
// data contains one, since it runs inside every cycle check.
func (suite *OrgHierarchyTestSuite) TestAncestorWalkTerminatesOnCorruptedCycle() {
	first, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{Name: "Corrupted first", Slug: "corrupted-first"})
	require.NoError(suite.T(), err)
	second, err := suite.service.CreateOrg(suite.ctx, orgDTO.CreateOrg{ParentOrgID: &first.ID, Name: "Corrupted second", Slug: "corrupted-second"})
	require.NoError(suite.T(), err)
	_, err = suite.conn.Exec(suite.ctx, "UPDATE org SET parent_org_id = $1 WHERE id = $2", second.ID, first.ID)
	require.NoError(suite.T(), err)

	ctx, cancel := context.WithTimeout(suite.ctx, hierarchyTimeout)
	defer cancel()
	ancestorIDs, err := suite.queries.GetOrgAncestorIDs(ctx, first.ID)

	require.NoError(suite.T(), err)
	assert.ElementsMatch(suite.T(), []uuid.UUID{second.ID, first.ID}, ancestorIDs)
}

func TestOrgHierarchyTestSuite(t *testing.T) {
	suite.Run(t, new(OrgHierarchyTestSuite))
}
