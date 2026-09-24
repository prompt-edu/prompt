package copy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

type CopyServiceTestSuite struct {
	suite.Suite
	suiteCtx    context.Context
	cleanup     func()
	queries     db.Queries
	copyService *CopyService
}

func (s *CopyServiceTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		s.T().Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup
	s.queries = *testDB.Queries
	s.copyService = NewCopyService(s.queries)
}

func (s *CopyServiceTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// configurePhase stores a fully configured certificate phase: template, release date, student
// page text, and a recorded download.
func (s *CopyServiceTestSuite) configurePhase(template, studentPageText string) uuid.UUID {
	coursePhaseID := uuid.New()
	_, err := s.queries.UpsertCoursePhaseConfig(s.suiteCtx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:   coursePhaseID,
		TemplateContent: pgtype.Text{String: template, Valid: true},
		UpdatedBy:       pgtype.Text{String: "Source Lecturer", Valid: true},
	})
	s.Require().NoError(err)
	_, err = s.queries.UpdateReleaseDate(s.suiteCtx, db.UpdateReleaseDateParams{
		CoursePhaseID: coursePhaseID,
		ReleaseDate:   pgtype.Timestamptz{Time: time.Now().Add(-24 * time.Hour), Valid: true},
		UpdatedBy:     pgtype.Text{String: "Source Lecturer", Valid: true},
	})
	s.Require().NoError(err)
	_, err = s.queries.UpsertStudentPageText(s.suiteCtx, db.UpsertStudentPageTextParams{
		CoursePhaseID:   coursePhaseID,
		StudentPageText: pgtype.Text{String: studentPageText, Valid: true},
	})
	s.Require().NoError(err)
	_, err = s.queries.RecordCertificateDownload(s.suiteCtx, db.RecordCertificateDownloadParams{
		StudentID:     uuid.New(),
		CoursePhaseID: coursePhaseID,
	})
	s.Require().NoError(err)
	return coursePhaseID
}

func (s *CopyServiceTestSuite) TestCopyPhaseCopiesTemplateAndStudentPageText() {
	source := s.configurePhase("= Source Certificate", "<p>Well done!</p>")
	target := uuid.New()

	s.Require().NoError(s.copyService.CopyPhase(s.suiteCtx, source, target))

	copied, err := s.queries.GetCoursePhaseConfig(s.suiteCtx, target)
	s.Require().NoError(err)
	s.Equal("= Source Certificate", copied.TemplateContent.String)
	s.Equal("<p>Well done!</p>", copied.StudentPageText.String)
	s.False(copied.ReleaseDate.Valid, "the release date is printed on the certificate and must not carry over")
	s.False(copied.UpdatedBy.Valid)

	hasDownloads, err := s.queries.HasDownloads(s.suiteCtx, target)
	s.Require().NoError(err)
	s.False(hasDownloads, "download records belong to the source phase's students")

	sourceConfig, err := s.queries.GetCoursePhaseConfig(s.suiteCtx, source)
	s.Require().NoError(err)
	s.Equal("= Source Certificate", sourceConfig.TemplateContent.String, "the source phase is left untouched")
	s.True(sourceConfig.ReleaseDate.Valid)
}

func (s *CopyServiceTestSuite) TestCopyPhaseReplacesTheTargetTemplate() {
	source := s.configurePhase("= Source Certificate", "<p>Source text</p>")
	target := s.configurePhase("= Target Certificate", "<p>Target text</p>")
	before, err := s.queries.GetCoursePhaseConfig(s.suiteCtx, target)
	s.Require().NoError(err)

	s.Require().NoError(s.copyService.CopyPhase(s.suiteCtx, source, target))

	after, err := s.queries.GetCoursePhaseConfig(s.suiteCtx, target)
	s.Require().NoError(err)
	s.Equal("= Source Certificate", after.TemplateContent.String)
	s.Equal("<p>Source text</p>", after.StudentPageText.String)
	s.Equal(before.ReleaseDate, after.ReleaseDate, "the target keeps its own release date")
	s.False(after.UpdatedBy.Valid, "the target's updater no longer describes the copied template")
}

func (s *CopyServiceTestSuite) TestCopyPhaseWithoutSourceConfigCopiesNothing() {
	target := uuid.New()

	s.Require().NoError(s.copyService.CopyPhase(s.suiteCtx, uuid.New(), target))

	_, err := s.queries.GetCoursePhaseConfig(s.suiteCtx, target)
	s.True(errors.Is(err, pgx.ErrNoRows))
}

func (s *CopyServiceTestSuite) TestCopyPhaseOntoItselfIsANoOp() {
	source := s.configurePhase("= Source Certificate", "<p>Well done!</p>")

	s.Require().NoError(s.copyService.CopyPhase(s.suiteCtx, source, source))

	unchanged, err := s.queries.GetCoursePhaseConfig(s.suiteCtx, source)
	s.Require().NoError(err)
	s.True(unchanged.ReleaseDate.Valid)
	s.Equal("Source Lecturer", unchanged.UpdatedBy.String)
}

func TestCopyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CopyServiceTestSuite))
}

// The probe core sends to detect copy support never reaches the database.
func TestHandlePhaseCopyProbeReturnsNoError(t *testing.T) {
	phaseID := uuid.New()
	require.NoError(t, NewCopyService(*db.New(nil)).CopyPhase(context.Background(), phaseID, phaseID))
}
