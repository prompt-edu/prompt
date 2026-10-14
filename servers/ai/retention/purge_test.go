package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/ai/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	assert.Error(t, Validate(730), "metadata must outlive the longest content retention")
	assert.NoError(t, Validate(731))
}

func TestPurge(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()
	now := time.Now()

	insertCall := func(feature string, age time.Duration, completed bool) uuid.UUID {
		var callID uuid.UUID
		require.NoError(t, testDB.Conn.QueryRow(ctx, `
			INSERT INTO ai_call (course_phase_id, actor_id, actor_role, issuer, feature, provider, outcome, streamed,
			                     server_version, requested_at, completed_at)
			VALUES (gen_random_uuid(), 'actor', 'Lecturer', 'https://keycloak.test/realms/prompt', $1, 'logos.test',
			        CASE WHEN $3 THEN 'success' ELSE 'pending' END, false, 'test', $2,
			        CASE WHEN $3 THEN $2::timestamptz END)
			RETURNING id`, feature, now.Add(-age), completed).Scan(&callID))
		for _, statement := range []string{
			`INSERT INTO ai_call_content (call_id, request) VALUES ($1, '\x00')`,
			`INSERT INTO ai_call_subject (call_id, course_participation_id) VALUES ($1, gen_random_uuid())`,
			`INSERT INTO ai_call_event (call_id, actor_id, type) VALUES ($1, 'actor', 'shown')`,
		} {
			_, err := testDB.Conn.Exec(ctx, statement, callID)
			require.NoError(t, err)
		}
		return callID
	}
	day := 24 * time.Hour
	retainedAssessment := insertCall("assessment.action_item_suggestions", 200*day, true)
	expiredAssessment := insertCall("assessment.action_item_suggestions", 800*day, true)
	expiredMetadata := insertCall("assessment.action_item_suggestions", 2000*day, true)
	unknownFeature := insertCall("removed.feature", 200*day, true)
	abandoned := insertCall("assessment.action_item_suggestions", 2*time.Hour, false)

	require.NoError(t, Purge(ctx, testDB.Queries, now, 1825))

	exists := func(table string, callID uuid.UUID) bool {
		var found bool
		column := "call_id"
		if table == "ai_call" {
			column = "id"
		}
		require.NoError(t, testDB.Conn.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM "+table+" WHERE "+column+" = $1)", callID).Scan(&found))
		return found
	}
	assert.True(t, exists("ai_call_content", retainedAssessment))
	assert.True(t, exists("ai_call_subject", retainedAssessment))
	assert.False(t, exists("ai_call_content", expiredAssessment))
	assert.False(t, exists("ai_call_subject", expiredAssessment), "subjects go with the content")
	assert.True(t, exists("ai_call", expiredAssessment), "metadata outlives the content")
	assert.True(t, exists("ai_call_event", expiredAssessment))
	assert.False(t, exists("ai_call", expiredMetadata))
	assert.False(t, exists("ai_call_event", expiredMetadata))
	assert.True(t, exists("ai_call_content", unknownFeature), "an unknown feature keeps the longest retention")

	var outcome, errorCode string
	require.NoError(t, testDB.Conn.QueryRow(ctx,
		"SELECT outcome, error_code FROM ai_call WHERE id = $1 AND completed_at IS NOT NULL", abandoned).Scan(&outcome, &errorCode))
	assert.Equal(t, "error", outcome)
	assert.Equal(t, "abandoned", errorCode)
}
