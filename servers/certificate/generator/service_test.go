package generator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/certificate/config/configDTO"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/prompt-edu/prompt/servers/certificate/participants"
	"github.com/stretchr/testify/assert"
)

type fakeTemplateProvider struct {
	template    string
	releaseDate *time.Time
}

func (f fakeTemplateProvider) GetTemplateContent(_ context.Context, _ uuid.UUID) (string, error) {
	return f.template, nil
}

func (f fakeTemplateProvider) GetCoursePhaseConfig(_ context.Context, coursePhaseID uuid.UUID) (configDTO.CoursePhaseConfig, error) {
	return configDTO.CoursePhaseConfig{CoursePhaseID: coursePhaseID, ReleaseDate: f.releaseDate}, nil
}

type fakeCourseInfoProvider struct {
	courseName string
	teamName   string
}

func (f fakeCourseInfoProvider) GetCoursePhaseWithCourse(_ context.Context, _ string, coursePhaseID uuid.UUID) (*participants.CoursePhaseWithCourse, error) {
	return &participants.CoursePhaseWithCourse{ID: coursePhaseID, Course: participants.Course{Name: f.courseName}}, nil
}

func (f fakeCourseInfoProvider) GetStudentTeamName(_ context.Context, _ string, _, _ uuid.UUID) (string, error) {
	return f.teamName, nil
}

// captureCompile stands in for the typst binary and reports back what the template would have read.
func captureCompile(data *CertificateData) func(context.Context, string, string) ([]byte, error) {
	return func(_ context.Context, tempDir, _ string) ([]byte, error) {
		content, err := os.ReadFile(filepath.Join(tempDir, "data.json"))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, data); err != nil {
			return nil, err
		}
		return []byte("%PDF-fake"), nil
	}
}

func newTestService(config templateProvider, courses courseInfoProvider, now time.Time, compile func(context.Context, string, string) ([]byte, error)) *GeneratorService {
	return &GeneratorService{
		queries: db.Queries{},
		config:  config,
		courses: courses,
		now:     func() time.Time { return now },
		compile: compile,
	}
}

func TestGenerateCertificate_UsesReleaseDate(t *testing.T) {
	releaseDate := time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC)
	var written CertificateData
	service := newTestService(
		fakeTemplateProvider{template: "= Certificate", releaseDate: &releaseDate},
		fakeCourseInfoProvider{courseName: "iPraktikum", teamName: "Hermes"},
		time.Date(2026, time.August, 27, 9, 0, 0, 0, time.UTC),
		captureCompile(&written),
	)

	pdf, err := service.GenerateCertificate(context.Background(), "Bearer token", uuid.New(), &promptTypes.Student{Person: promptTypes.Person{FirstName: "Jane", LastName: "Doe"}})

	assert.NoError(t, err)
	assert.Equal(t, []byte("%PDF-fake"), pdf)
	assert.Equal(t, CertificateData{
		StudentName: "Jane Doe",
		CourseName:  "iPraktikum",
		TeamName:    "Hermes",
		Date:        "March 15, 2026",
	}, written)
}

func TestGenerateCertificate_FallsBackToToday(t *testing.T) {
	var written CertificateData
	service := newTestService(
		fakeTemplateProvider{template: "= Certificate"},
		fakeCourseInfoProvider{courseName: "iPraktikum"},
		time.Date(2026, time.August, 27, 9, 0, 0, 0, time.UTC),
		captureCompile(&written),
	)

	_, err := service.GenerateCertificate(context.Background(), "Bearer token", uuid.New(), &promptTypes.Student{Person: promptTypes.Person{FirstName: "Jane", LastName: "Doe"}})

	assert.NoError(t, err)
	assert.Equal(t, "August 27, 2026", written.Date)
	assert.Empty(t, written.TeamName)
}

func TestGeneratePreviewCertificate_UsesMockData(t *testing.T) {
	var written CertificateData
	service := newTestService(
		fakeTemplateProvider{template: "= Certificate"},
		fakeCourseInfoProvider{},
		time.Date(2026, time.August, 27, 9, 0, 0, 0, time.UTC),
		captureCompile(&written),
	)

	pdf, err := service.GeneratePreviewCertificate(context.Background(), uuid.New())

	assert.NoError(t, err)
	assert.Equal(t, []byte("%PDF-fake"), pdf)
	assert.Equal(t, "Jane Doe", written.StudentName)
	assert.Equal(t, "August 27, 2026", written.Date)
}
