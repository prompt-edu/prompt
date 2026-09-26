package profilePicture

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/profilePicture/profilePictureDTO"
	"github.com/prompt-edu/prompt/servers/core/storage"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Seeded in database_dumps/profile_picture_test.sql.
var (
	adaStudentID             = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	graceStudentID           = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
	alanStudentID            = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000003")
	adaParticipationID       = uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	graceParticipationID     = uuid.MustParse("cccccccc-0000-0000-0000-000000000002")
	adaUserID                = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	instructorUserID         = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	adaPictureStorageKey     = "profile-picture/bbbbbbbb-0000-0000-0000-000000000001/ada.jpg"
	instructorPictureKey     = "profile-picture/bbbbbbbb-0000-0000-0000-000000000002/instructor.jpg"
	validJPEG                = append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, []byte("jpeg body")...)
	notAJPEG                 = []byte("\x89PNG\r\n\x1a\n")
	oversizedPictureContents = append([]byte{0xFF, 0xD8, 0xFF}, bytes.Repeat([]byte{0}, maxPictureSizeBytes)...)
)

// objectStore is an in-memory stand-in for the S3 bucket behind the mock adapter.
type objectStore struct {
	mu          sync.Mutex
	objects     map[string][]byte
	contentType map[string]string
	deleted     []string
}

func newObjectStore() *objectStore {
	return &objectStore{objects: map[string][]byte{}, contentType: map[string]string{}}
}

func (o *objectStore) put(key, contentType string, content []byte) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.objects[key] = content
	o.contentType[key] = contentType
}

func (o *objectStore) wasDeleted(key string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, deleted := range o.deleted {
		if deleted == key {
			return true
		}
	}
	return false
}

func (o *objectStore) adapter() *storage.MockStorageAdapter {
	return &storage.MockStorageAdapter{
		GetMetadataFunc: func(_ context.Context, key string) (*storage.FileMetadata, error) {
			o.mu.Lock()
			defer o.mu.Unlock()
			content, ok := o.objects[key]
			if !ok {
				return nil, storage.ErrObjectNotFound
			}
			return &storage.FileMetadata{StorageKey: key, Size: int64(len(content)), ContentType: o.contentType[key]}, nil
		},
		DownloadFunc: func(_ context.Context, key string) (io.ReadCloser, error) {
			o.mu.Lock()
			defer o.mu.Unlock()
			content, ok := o.objects[key]
			if !ok {
				return nil, storage.ErrObjectNotFound
			}
			return io.NopCloser(bytes.NewReader(content)), nil
		},
		DeleteFunc: func(_ context.Context, key string) error {
			o.mu.Lock()
			defer o.mu.Unlock()
			delete(o.objects, key)
			o.deleted = append(o.deleted, key)
			return nil
		},
	}
}

type ProfilePictureServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	queries db.Queries
	conn    *pgxpool.Pool
	store   *objectStore
	service *ProfilePictureService
}

func (suite *ProfilePictureServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/profile_picture_test.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.queries = *testDB.Queries
	suite.conn = testDB.Conn
}

func (suite *ProfilePictureServiceTestSuite) SetupTest() {
	suite.store = newObjectStore()
	suite.store.put(adaPictureStorageKey, pictureContentType, validJPEG)
	suite.store.put(instructorPictureKey, pictureContentType, validJPEG)

	fileStore := files.NewStorageService(suite.queries, suite.conn, suite.store.adapter(), 50, []string{pictureContentType, "image/png"})
	suite.service = NewProfilePictureService(suite.queries, fileStore)
}

func (suite *ProfilePictureServiceTestSuite) TearDownSuite() {
	suite.cleanup()
}

// upload presigns and fakes the client's PUT, returning the storage key to complete.
func (suite *ProfilePictureServiceTestSuite) upload(userID uuid.UUID, contentType string, content []byte) string {
	presigned, err := suite.service.PresignUpload(suite.ctx, userID)
	require.NoError(suite.T(), err)
	suite.store.put(presigned.StorageKey, contentType, content)
	return presigned.StorageKey
}

// insertStudent adds a student with the given university login so tests can link uploads
// without touching the seeded students other tests rely on.
func (suite *ProfilePictureServiceTestSuite) insertStudent(universityLogin string) uuid.UUID {
	studentID := uuid.New()
	_, err := suite.conn.Exec(suite.ctx,
		`INSERT INTO student (id, first_name, last_name, email, university_login, has_university_account, gender)
		 VALUES ($1, 'Test', 'Student', $2, $3, true, 'diverse')`,
		studentID, fmt.Sprintf("%s@tum.de", universityLogin), universityLogin)
	require.NoError(suite.T(), err)
	return studentID
}

func (suite *ProfilePictureServiceTestSuite) TestPresignUpload_ScopesKeyToCaller() {
	userID := uuid.New()

	presigned, err := suite.service.PresignUpload(suite.ctx, userID)

	require.NoError(suite.T(), err)
	assert.True(suite.T(), strings.HasPrefix(presigned.StorageKey, fmt.Sprintf("profile-picture/%s/", userID)))
	assert.NotEmpty(suite.T(), presigned.UploadURL)
}

func (suite *ProfilePictureServiceTestSuite) TestCompleteUpload_SetsPictureAndLinksStudent() {
	userID := uuid.New()
	studentID := suite.insertStudent("te01sta")
	key := suite.upload(userID, pictureContentType, validJPEG)

	picture, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID, UniversityLogin: "te01sta"}, key)

	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), picture.URL, key)

	own, err := suite.service.GetOwnPicture(suite.ctx, userID)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), own.URL, key)

	urls, err := suite.service.LookupPictureURLs(suite.ctx, profilePictureDTO.LookupRequest{StudentIDs: []uuid.UUID{studentID}})
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), urls.Students[studentID], key)
}

func (suite *ProfilePictureServiceTestSuite) TestCompleteUpload_ReplacesAndDeletesPreviousPicture() {
	userID := uuid.New()
	firstKey := suite.upload(userID, pictureContentType, validJPEG)
	_, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, firstKey)
	require.NoError(suite.T(), err)

	secondKey := suite.upload(userID, pictureContentType, validJPEG)
	picture, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, secondKey)

	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), picture.URL, secondKey)
	assert.True(suite.T(), suite.store.wasDeleted(firstKey), "the replaced picture must be removed from storage")
	assert.False(suite.T(), suite.store.wasDeleted(secondKey))
}

func (suite *ProfilePictureServiceTestSuite) TestCompleteUpload_IsIdempotent() {
	userID := uuid.New()
	key := suite.upload(userID, pictureContentType, validJPEG)
	_, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, key)
	require.NoError(suite.T(), err)

	_, err = suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, key)

	require.NoError(suite.T(), err)
	assert.False(suite.T(), suite.store.wasDeleted(key), "completing twice must not delete the current picture")
}

func (suite *ProfilePictureServiceTestSuite) TestCompleteUpload_RejectsForeignStorageKey() {
	owner := uuid.New()
	key := suite.upload(owner, pictureContentType, validJPEG)

	_, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: uuid.New()}, key)

	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
}

func (suite *ProfilePictureServiceTestSuite) TestCompleteUpload_RejectsMissingUpload() {
	userID := uuid.New()
	key := fmt.Sprintf("profile-picture/%s/never-uploaded.jpg", userID)

	_, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, key)

	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *ProfilePictureServiceTestSuite) TestCompleteUpload_RejectsInvalidPictures() {
	cases := map[string]struct {
		contentType string
		content     []byte
	}{
		"not a JPEG":        {pictureContentType, notAJPEG},
		"wrong type":        {"image/png", validJPEG},
		"over size limit":   {pictureContentType, oversizedPictureContents},
		"empty upload body": {pictureContentType, []byte{}},
	}
	for name, tc := range cases {
		suite.Run(name, func() {
			userID := uuid.New()
			key := suite.upload(userID, tc.contentType, tc.content)

			_, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, key)

			assert.ErrorIs(suite.T(), err, ErrInvalidInput)
			assert.True(suite.T(), suite.store.wasDeleted(key), "a rejected upload must be removed from storage")
			_, err = suite.service.GetOwnPicture(suite.ctx, userID)
			assert.ErrorIs(suite.T(), err, ErrNotFound)
		})
	}
}

func (suite *ProfilePictureServiceTestSuite) TestDeleteOwnPicture() {
	userID := uuid.New()
	key := suite.upload(userID, pictureContentType, validJPEG)
	_, err := suite.service.CompleteUpload(suite.ctx, Uploader{UserID: userID}, key)
	require.NoError(suite.T(), err)

	require.NoError(suite.T(), suite.service.DeleteOwnPicture(suite.ctx, userID))

	assert.True(suite.T(), suite.store.wasDeleted(key))
	_, err = suite.service.GetOwnPicture(suite.ctx, userID)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.ErrorIs(suite.T(), suite.service.DeleteOwnPicture(suite.ctx, userID), ErrNotFound)
}

func (suite *ProfilePictureServiceTestSuite) TestLookupPictureURLs() {
	unknownID := uuid.New()

	urls, err := suite.service.LookupPictureURLs(suite.ctx, profilePictureDTO.LookupRequest{
		UserIDs:                []uuid.UUID{adaUserID, instructorUserID, unknownID},
		StudentIDs:             []uuid.UUID{adaStudentID, graceStudentID, alanStudentID},
		CourseParticipationIDs: []uuid.UUID{adaParticipationID, graceParticipationID},
	})

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), urls.Users, 2)
	assert.Contains(suite.T(), urls.Users[adaUserID], adaPictureStorageKey)
	assert.Contains(suite.T(), urls.Users[instructorUserID], instructorPictureKey)

	assert.Len(suite.T(), urls.Students, 1, "only students whose university login has a picture resolve")
	assert.Contains(suite.T(), urls.Students[adaStudentID], adaPictureStorageKey)

	assert.Len(suite.T(), urls.CourseParticipations, 1)
	assert.Contains(suite.T(), urls.CourseParticipations[adaParticipationID], adaPictureStorageKey)
}

func (suite *ProfilePictureServiceTestSuite) TestLookupPictureURLs_EmptyRequest() {
	urls, err := suite.service.LookupPictureURLs(suite.ctx, profilePictureDTO.LookupRequest{})

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), urls.Users)
	assert.Empty(suite.T(), urls.Students)
	assert.Empty(suite.T(), urls.CourseParticipations)
}

func (suite *ProfilePictureServiceTestSuite) TestLookupPictureURLs_RejectsTooManyIDs() {
	ids := make([]uuid.UUID, maxLookupIDs+1)
	for i := range ids {
		ids[i] = uuid.New()
	}

	_, err := suite.service.LookupPictureURLs(suite.ctx, profilePictureDTO.LookupRequest{UserIDs: ids})

	assert.ErrorIs(suite.T(), err, ErrInvalidInput)
}

func TestProfilePictureServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProfilePictureServiceTestSuite))
}
