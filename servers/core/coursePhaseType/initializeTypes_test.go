package coursePhaseType

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPresentationPhaseTypeExists(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(ctx, "../database_dumps/copy_course_test.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	require.NoError(t, err)
	defer cleanup()

	CoursePhaseTypeServiceSingleton = &CoursePhaseTypeService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	require.NoError(t, initPresentation())
	require.NoError(t, initPresentation(), "initialization must be idempotent")

	exists, err := testDB.Queries.TestPresentationPhaseTypeExists(ctx)
	require.NoError(t, err)
	assert.True(t, exists)

	phaseTypes, err := testDB.Queries.GetAllCoursePhaseTypes(ctx)
	require.NoError(t, err)

	presentationCount := 0
	for _, phaseType := range phaseTypes {
		if phaseType.Name != "Presentation" {
			continue
		}
		presentationCount++
		assert.Equal(t, "{CORE_HOST}/presentation/api", phaseType.BaseUrl)

		phaseInputs, err := testDB.Queries.GetCoursePhaseRequiredPhaseInputs(ctx, phaseType.ID)
		require.NoError(t, err)
		require.Len(t, phaseInputs, 1)
		assert.Equal(t, "teams", phaseInputs[0].DtoName)
		assert.True(t, phaseInputs[0].Optional)

		participationInputs, err := testDB.Queries.GetCoursePhaseRequiredParticipationInputs(ctx, phaseType.ID)
		require.NoError(t, err)
		require.Len(t, participationInputs, 1)
		assert.Equal(t, "teamAllocation", participationInputs[0].DtoName)
		assert.True(t, participationInputs[0].Optional)
	}
	assert.Equal(t, 1, presentationCount)

	presentationDTOs, err := GetAllCoursePhaseTypes(ctx)
	require.NoError(t, err)
	for _, phaseType := range presentationDTOs {
		if phaseType.Name == "Presentation" {
			require.Len(t, phaseType.RequiredPhaseInputDTOs, 1)
			assert.True(t, phaseType.RequiredPhaseInputDTOs[0].Optional)
			require.Len(t, phaseType.RequiredParticipationInputDTOs, 1)
			assert.True(t, phaseType.RequiredParticipationInputDTOs[0].Optional)
			return
		}
	}
	t.Fatal("Presentation phase type DTO not found")
}

func TestPresentationPhaseTypeInitializationRestoresMissingInputs(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(ctx, "../database_dumps/copy_course_test.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	require.NoError(t, err)
	defer cleanup()

	CoursePhaseTypeServiceSingleton = &CoursePhaseTypeService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	presentationPhaseID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	require.NoError(t, testDB.Queries.CreateCoursePhaseType(ctx, db.CreateCoursePhaseTypeParams{
		ID:           presentationPhaseID,
		Name:         "Presentation",
		InitialPhase: false,
		BaseUrl:      "{CORE_HOST}/presentation/api",
		Description:  pgtype.Text{String: "Existing Presentation phase type without input descriptors.", Valid: true},
	}))

	require.NoError(t, initPresentation())
	require.NoError(t, initPresentation(), "restored input descriptors must remain idempotent")

	phaseInputs, err := testDB.Queries.GetCoursePhaseRequiredPhaseInputs(ctx, presentationPhaseID)
	require.NoError(t, err)
	require.Len(t, phaseInputs, 1)
	assert.Equal(t, "teams", phaseInputs[0].DtoName)
	assert.True(t, phaseInputs[0].Optional)

	participationInputs, err := testDB.Queries.GetCoursePhaseRequiredParticipationInputs(ctx, presentationPhaseID)
	require.NoError(t, err)
	require.Len(t, participationInputs, 1)
	assert.Equal(t, "teamAllocation", participationInputs[0].DtoName)
	assert.True(t, participationInputs[0].Optional)
}
