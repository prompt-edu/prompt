package presentation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/testutils"
)

const uploadTTLSeconds = 60

type MaterialDBTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service *Service
	storage *testutils.FakeStorage
	now     time.Time
}

func (s *MaterialDBTestSuite) SetupSuite() {
	s.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(s.ctx, "../database_dumps/presentation_seed.sql")
	require.NoError(s.T(), err)
	s.cleanup = cleanup
	s.storage = testutils.NewFakeStorage()
	s.service = NewService(
		testDB.Queries, testDB.Conn, s.storage, "http://core.test",
		uploadTTLSeconds, 60, 50*1024*1024,
		[]string{"application/pdf"},
	)
	// Staff paths never call out to core, so no mock core service is needed here.
	s.now = time.Now()
	s.service.now = func() time.Time { return s.now }
}

func (s *MaterialDBTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *MaterialDBTestSuite) upload(
	presentationID uuid.UUID, user User, fileName, contentType string,
) (PresignMaterialResponse, error) {
	return s.service.CreateUploadIntent(s.ctx, "", individualPhaseID, presentationID, user,
		PresignMaterialRequest{FileName: fileName, ContentType: contentType, SizeBytes: 1024})
}

// The completion guard used to reuse the presign TTL, so any upload slower than the
// presigned URL's lifetime could never be completed and its object leaked.
func (s *MaterialDBTestSuite) TestUploadCompletesAfterPresignExpiry() {
	user := instructor("uploader")
	intent, err := s.upload(adaPresentationID, user, "slides.pdf", "application/pdf")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.now.Add(uploadTTLSeconds*time.Second), intent.ExpiresAt,
		"the response must advertise the presign expiry, not the reclaim deadline")

	material, err := s.service.queries.GetPresentationMaterial(s.ctx, db.GetPresentationMaterialParams{
		ID:            intent.UploadID,
		CoursePhaseID: individualPhaseID,
	})
	require.NoError(s.T(), err)
	s.storage.Put(material.StorageKey, "application/pdf", 4096)

	// The browser PUT took far longer than the presigned URL was valid for.
	s.now = s.now.Add(30 * time.Minute)

	completed, err := s.service.CompleteUpload(
		s.ctx, "", individualPhaseID, adaPresentationID, intent.UploadID, user)
	require.NoError(s.T(), err)
	assert.EqualValues(s.T(), 4096, completed.SizeBytes)
}

// A pending row is invisible in the UI, so counting it as a dependency made the slot
// impossible to unassign through any supported action.
func (s *MaterialDBTestSuite) TestPendingUploadDoesNotBlockUnassign() {
	phaseID := individualPhaseID
	slotID := uuid.MustParse("30000000-0000-0000-0000-000000000002")
	presentationID := uuid.MustParse("40000000-0000-0000-0000-000000000002")
	user := instructor("abandoner")

	intent, err := s.upload(presentationID, user, "abandoned.pdf", "application/pdf")
	require.NoError(s.T(), err)
	material, err := s.service.queries.GetPresentationMaterial(s.ctx, db.GetPresentationMaterialParams{
		ID:            intent.UploadID,
		CoursePhaseID: phaseID,
	})
	require.NoError(s.T(), err)
	s.storage.Put(material.StorageKey, "application/pdf", 1024)

	require.NoError(s.T(), s.service.UnassignTarget(s.ctx, phaseID, slotID))
	// Unassign cascades the row, so the object has to be cleaned up alongside it.
	assert.False(s.T(), s.storage.Has(material.StorageKey),
		"cascaded material rows must not leave their objects behind")
}

func (s *MaterialDBTestSuite) TestReclaimRemovesExpiredPendingUploadsOnly() {
	user := instructor("reclaim")
	expiring, err := s.upload(adaPresentationID, user, "expiring.pdf", "application/pdf")
	require.NoError(s.T(), err)
	expiringRow, err := s.service.queries.GetPresentationMaterial(s.ctx, db.GetPresentationMaterialParams{
		ID:            expiring.UploadID,
		CoursePhaseID: individualPhaseID,
	})
	require.NoError(s.T(), err)
	s.storage.Put(expiringRow.StorageKey, "application/pdf", 1024)

	// A second upload that gets completed before the reaper runs must survive.
	kept, err := s.upload(adaPresentationID, user, "kept.pdf", "application/pdf")
	require.NoError(s.T(), err)
	keptRow, err := s.service.queries.GetPresentationMaterial(s.ctx, db.GetPresentationMaterialParams{
		ID:            kept.UploadID,
		CoursePhaseID: individualPhaseID,
	})
	require.NoError(s.T(), err)
	s.storage.Put(keptRow.StorageKey, "application/pdf", 2048)
	_, err = s.service.CompleteUpload(
		s.ctx, "", individualPhaseID, adaPresentationID, kept.UploadID, user)
	require.NoError(s.T(), err)

	// Nothing has expired yet.
	reclaimed, err := s.service.ReclaimExpiredMaterials(s.ctx, 10)
	require.NoError(s.T(), err)
	assert.Zero(s.T(), reclaimed)

	// Wait past the reclaim deadline. expires_at is stored, so move real time forward by
	// asking the database to treat the row as old.
	_, err = s.service.conn.Exec(s.ctx,
		"UPDATE presentation_material SET expires_at = now() - interval '1 minute' WHERE id = $1",
		expiring.UploadID)
	require.NoError(s.T(), err)

	reclaimed, err = s.service.ReclaimExpiredMaterials(s.ctx, 10)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, reclaimed)
	assert.False(s.T(), s.storage.Has(expiringRow.StorageKey))
	assert.True(s.T(), s.storage.Has(keptRow.StorageKey), "completed uploads must not be reclaimed")
}

func (s *MaterialDBTestSuite) TestDisallowedContentTypeRejected() {
	user := instructor("mime")

	_, err := s.upload(adaPresentationID, user, "payload.html", "text/html")
	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 415, apiErr.Status)
	assert.Equal(s.T(), "material_type_not_allowed", apiErr.Code)

	// Presigning honestly and then storing something else must fail at completion, and
	// must not leave the object behind.
	intent, err := s.upload(adaPresentationID, user, "trojan.pdf", "application/pdf")
	require.NoError(s.T(), err)
	material, err := s.service.queries.GetPresentationMaterial(s.ctx, db.GetPresentationMaterialParams{
		ID:            intent.UploadID,
		CoursePhaseID: individualPhaseID,
	})
	require.NoError(s.T(), err)
	s.storage.Put(material.StorageKey, "text/html", 512)

	_, err = s.service.CompleteUpload(
		s.ctx, "", individualPhaseID, adaPresentationID, intent.UploadID, user)
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), "material_type_not_allowed", apiErr.Code)
	assert.False(s.T(), s.storage.Has(material.StorageKey))
}

// A presenter may curate their own uploads but not an instructor's rubric.
func (s *MaterialDBTestSuite) TestStudentCannotDeleteInstructorMaterial() {
	lecturer := instructor("lecturer")
	intent, err := s.upload(adaPresentationID, lecturer, "rubric.pdf", "application/pdf")
	require.NoError(s.T(), err)
	material, err := s.service.queries.GetPresentationMaterial(s.ctx, db.GetPresentationMaterialParams{
		ID:            intent.UploadID,
		CoursePhaseID: individualPhaseID,
	})
	require.NoError(s.T(), err)
	s.storage.Put(material.StorageKey, "application/pdf", 1024)
	_, err = s.service.CompleteUpload(
		s.ctx, "", individualPhaseID, adaPresentationID, intent.UploadID, lecturer)
	require.NoError(s.T(), err)

	// The presenter of this individual presentation, i.e. authorized for it but not the
	// uploader of this file.
	student := User{
		ID:                    "student-ada",
		Name:                  "Ada Lovelace",
		CourseParticipationID: uuid.MustParse("50000000-0000-0000-0000-000000000001"),
	}
	err = s.service.DeleteMaterial(
		s.ctx, "", individualPhaseID, adaPresentationID, intent.UploadID, student)

	var apiErr *APIError
	require.True(s.T(), errors.As(err, &apiErr))
	assert.Equal(s.T(), 403, apiErr.Status)
	assert.Equal(s.T(), "material_forbidden", apiErr.Code)
	assert.True(s.T(), s.storage.Has(material.StorageKey))
}

func TestMaterialDBTestSuite(t *testing.T) {
	suite.Run(t, new(MaterialDBTestSuite))
}
