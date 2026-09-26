package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectProfilePictureFileIDs(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(ctx, "../../database_dumps/profile_picture_test.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	require.NoError(t, err)
	defer cleanup()

	s := &PrivacyService{queries: *testDB.Queries}

	// Seeded in database_dumps/profile_picture_test.sql.
	adaUserID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	adaStudentID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	adaFileID := uuid.MustParse("ffffffff-0000-0000-0000-000000000001")
	instructorUserID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	instructorFileID := uuid.MustParse("ffffffff-0000-0000-0000-000000000002")
	graceStudentID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")

	cases := map[string]struct {
		subject sdk.SubjectIdentifiers
		want    []uuid.UUID
	}{
		"platform user":                   {sdk.SubjectIdentifiers{UserID: instructorUserID}, []uuid.UUID{instructorFileID}},
		"student without known account":   {sdk.SubjectIdentifiers{StudentID: adaStudentID}, []uuid.UUID{adaFileID}},
		"student with account, deduped":   {sdk.SubjectIdentifiers{UserID: adaUserID, StudentID: adaStudentID}, []uuid.UUID{adaFileID}},
		"student without profile picture": {sdk.SubjectIdentifiers{StudentID: graceStudentID}, []uuid.UUID{}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fileIDs, err := s.collectProfilePictureFileIDs(ctx, tc.subject)

			require.NoError(t, err)
			assert.ElementsMatch(t, tc.want, fileIDs)
		})
	}
}
