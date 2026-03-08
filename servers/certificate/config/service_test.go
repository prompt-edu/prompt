package config

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
	"github.com/ls1intum/prompt2/servers/certificate/testutils"
)

type ConfigServiceTestSuite struct {
	suite.Suite
	suiteCtx      context.Context
	cleanup       func()
	configService *ConfigService
}

func (s *ConfigServiceTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql")
	if err != nil {
		s.T().Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup
	s.configService = &ConfigService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	ConfigServiceSingleton = s.configService
}

func (s *ConfigServiceTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *ConfigServiceTestSuite) TestGetCoursePhaseConfig_Existing() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")

	cfg, err := GetCoursePhaseConfig(s.suiteCtx, coursePhaseID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), coursePhaseID, cfg.CoursePhaseID)
	assert.True(s.T(), cfg.HasTemplate)
	assert.NotNil(s.T(), cfg.TemplateContent)
	assert.Contains(s.T(), *cfg.TemplateContent, "Certificate of Completion")
}

func (s *ConfigServiceTestSuite) TestGetCoursePhaseConfig_ExistingWithoutTemplate() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")

	cfg, err := GetCoursePhaseConfig(s.suiteCtx, coursePhaseID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), coursePhaseID, cfg.CoursePhaseID)
	assert.False(s.T(), cfg.HasTemplate)
	assert.Nil(s.T(), cfg.TemplateContent)
}

func (s *ConfigServiceTestSuite) TestGetCoursePhaseConfig_AutoCreate() {
	newID := uuid.New()

	cfg, err := GetCoursePhaseConfig(s.suiteCtx, newID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newID, cfg.CoursePhaseID)
	assert.False(s.T(), cfg.HasTemplate)
	assert.Nil(s.T(), cfg.TemplateContent)
}

func (s *ConfigServiceTestSuite) TestUpdateCoursePhaseConfig() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000002")
	template := "= Updated Template\nHello World"

	cfg, err := UpdateCoursePhaseConfig(s.suiteCtx, coursePhaseID, template, "Test User")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), coursePhaseID, cfg.CoursePhaseID)
	assert.True(s.T(), cfg.HasTemplate)
	assert.NotNil(s.T(), cfg.TemplateContent)
	assert.Equal(s.T(), template, *cfg.TemplateContent)
	assert.NotNil(s.T(), cfg.UpdatedBy)
	assert.Equal(s.T(), "Test User", *cfg.UpdatedBy)
}

func (s *ConfigServiceTestSuite) TestUpdateCoursePhaseConfig_Upsert() {
	newID := uuid.New()
	template := "= New Template"

	cfg, err := UpdateCoursePhaseConfig(s.suiteCtx, newID, template, "Another User")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newID, cfg.CoursePhaseID)
	assert.True(s.T(), cfg.HasTemplate)
	assert.Equal(s.T(), template, *cfg.TemplateContent)
}

func (s *ConfigServiceTestSuite) TestGetTemplateContent_WithTemplate() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")

	content, err := GetTemplateContent(s.suiteCtx, coursePhaseID)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), content)
	assert.Contains(s.T(), content, "Certificate of Completion")
}

func (s *ConfigServiceTestSuite) TestGetTemplateContent_WithoutTemplate() {
	// First reset the config to have no template
	resetID := uuid.New()
	_, err := s.configService.queries.CreateCoursePhaseConfig(s.suiteCtx, resetID)
	assert.NoError(s.T(), err)

	content, err := GetTemplateContent(s.suiteCtx, resetID)
	assert.Error(s.T(), err)
	assert.Empty(s.T(), content)
	assert.Contains(s.T(), err.Error(), "no template configured")
}

func (s *ConfigServiceTestSuite) TestGetTemplateContent_NonExistent() {
	nonExistentID := uuid.New()

	content, err := GetTemplateContent(s.suiteCtx, nonExistentID)
	assert.Error(s.T(), err)
	assert.Empty(s.T(), content)
}

func (s *ConfigServiceTestSuite) TestRecordCertificateDownload() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	studentID := uuid.New()

	download, err := s.configService.queries.RecordCertificateDownload(s.suiteCtx, db.RecordCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int32(1), download.DownloadCount)

	// Second download increments count
	download2, err := s.configService.queries.RecordCertificateDownload(s.suiteCtx, db.RecordCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int32(2), download2.DownloadCount)
}

func (s *ConfigServiceTestSuite) TestGetCertificateDownload() {
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")

	download, err := s.configService.queries.GetCertificateDownload(s.suiteCtx, db.GetCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), studentID, download.StudentID)
	assert.Equal(s.T(), int32(3), download.DownloadCount)
}

func (s *ConfigServiceTestSuite) TestListCertificateDownloads() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")

	downloads, err := s.configService.queries.ListCertificateDownloadsByCoursePhase(s.suiteCtx, coursePhaseID)
	assert.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(downloads), 1, "Should have at least 1 download record from seed data")
}

func (s *ConfigServiceTestSuite) TestListCertificateDownloads_Empty() {
	emptyID := uuid.New()

	downloads, err := s.configService.queries.ListCertificateDownloadsByCoursePhase(s.suiteCtx, emptyID)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), downloads)
}

func (s *ConfigServiceTestSuite) TestDeleteCertificateDownload() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000002")

	// Verify it exists
	_, err := s.configService.queries.GetCertificateDownload(s.suiteCtx, db.GetCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)

	// Delete it
	err = s.configService.queries.DeleteCertificateDownload(s.suiteCtx, db.DeleteCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.NoError(s.T(), err)

	// Verify it's gone
	_, err = s.configService.queries.GetCertificateDownload(s.suiteCtx, db.GetCertificateDownloadParams{
		StudentID:     studentID,
		CoursePhaseID: coursePhaseID,
	})
	assert.Error(s.T(), err)
}

func (s *ConfigServiceTestSuite) TestUpsertCoursePhaseConfig() {
	newID := uuid.New()
	template := "= Upsert Test"

	cfg, err := s.configService.queries.UpsertCoursePhaseConfig(s.suiteCtx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:   newID,
		TemplateContent: pgtype.Text{String: template, Valid: true},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newID, cfg.CoursePhaseID)
	assert.Equal(s.T(), template, cfg.TemplateContent.String)

	// Upsert again to update
	updatedTemplate := "= Updated Upsert"
	cfg2, err := s.configService.queries.UpsertCoursePhaseConfig(s.suiteCtx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:   newID,
		TemplateContent: pgtype.Text{String: updatedTemplate, Valid: true},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), updatedTemplate, cfg2.TemplateContent.String)
}

func TestConfigServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigServiceTestSuite))
}
