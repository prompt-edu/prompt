package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/storage"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Seeded in database_dumps/profile_picture_test.sql.
var (
	adaUserID            = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	adaStudentID         = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	adaFileID            = uuid.MustParse("ffffffff-0000-0000-0000-000000000001")
	adaPictureKey        = "profile-picture/bbbbbbbb-0000-0000-0000-000000000001/ada.jpg"
	instructorUserID     = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	instructorFileID     = uuid.MustParse("ffffffff-0000-0000-0000-000000000002")
	instructorPictureKey = "profile-picture/bbbbbbbb-0000-0000-0000-000000000002/instructor.jpg"
	graceStudentID       = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
)

// pictureStore is an in-memory stand-in for the bucket the seeded pictures live in.
type pictureStore struct {
	mu      sync.Mutex
	objects map[string][]byte
	deleted []string
}

func newPictureStore() *pictureStore {
	return &pictureStore{objects: map[string][]byte{
		adaPictureKey:        []byte("ada's picture"),
		instructorPictureKey: []byte("instructor's picture"),
	}}
}

func (p *pictureStore) wasDeleted(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, deleted := range p.deleted {
		if deleted == key {
			return true
		}
	}
	return false
}

func (p *pictureStore) adapter() *storage.MockStorageAdapter {
	return &storage.MockStorageAdapter{
		DownloadFunc: func(_ context.Context, key string) (io.ReadCloser, error) {
			p.mu.Lock()
			defer p.mu.Unlock()
			content, ok := p.objects[key]
			if !ok {
				return nil, storage.ErrObjectNotFound
			}
			return io.NopCloser(bytes.NewReader(content)), nil
		},
		DeleteFunc: func(_ context.Context, key string) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			delete(p.objects, key)
			p.deleted = append(p.deleted, key)
			return nil
		},
	}
}

// newProfilePicturePrivacyService serves the seeded pictures from a fresh database and store.
func newProfilePicturePrivacyService(t *testing.T) (*PrivacyService, *pictureStore) {
	ctx := context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(ctx, "../../database_dumps/profile_picture_test.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	require.NoError(t, err)
	t.Cleanup(cleanup)

	store := newPictureStore()
	fileStore := files.NewStorageService(*testDB.Queries, testDB.Conn, store.adapter(), 50, nil)
	return &PrivacyService{queries: *testDB.Queries, conn: testDB.Conn, applicationFiles: fileStore}, store
}

func TestCollectProfilePictureFileIDs(t *testing.T) {
	s, _ := newProfilePicturePrivacyService(t)

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
			fileIDs, err := s.collectProfilePictureFileIDs(context.Background(), tc.subject)

			require.NoError(t, err)
			assert.ElementsMatch(t, tc.want, fileIDs)
		})
	}
}

func TestExecuteCoreDeletion_DeletesProfilePicture(t *testing.T) {
	cases := map[string]sdk.SubjectIdentifiers{
		"student with account": {UserID: adaUserID, StudentID: adaStudentID},
		// An admin can delete a student who never logged in, so only the student is known
		"student without known account": {StudentID: adaStudentID},
	}

	for name, subject := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			s, store := newProfilePicturePrivacyService(t)

			require.NoError(t, s.ExecuteCoreDeletion(ctx, subject))

			assert.True(t, store.wasDeleted(adaPictureKey), "the picture must be removed from storage")
			_, err := s.queries.GetProfilePictureByUserID(ctx, adaUserID)
			assert.ErrorIs(t, err, pgx.ErrNoRows, "the profile_picture row must be gone")
			_, err = s.queries.GetFileByID(ctx, adaFileID)
			assert.ErrorIs(t, err, pgx.ErrNoRows, "the file record must be gone")

			assert.False(t, store.wasDeleted(instructorPictureKey), "other pictures must stay")
			_, err = s.queries.GetProfilePictureByUserID(ctx, instructorUserID)
			assert.NoError(t, err)
		})
	}
}

func TestAddProfilePictures_ExportContainsPicture(t *testing.T) {
	ctx := context.Background()
	s, store := newProfilePicturePrivacyService(t)
	ex, err := utils.NewExport()
	require.NoError(t, err)
	defer ex.Close()

	s.addProfilePictures(ctx, ex, sdk.SubjectIdentifiers{UserID: adaUserID, StudentID: adaStudentID})

	archive := uploadExport(t, ex)
	assert.Equal(t, store.objects[adaPictureKey], readZipEntry(t, archive, fmt.Sprintf("user/profile_pictures/%s.jpg", adaFileID)))
	assert.Nil(t, readZipEntry(t, archive, fmt.Sprintf("user/profile_pictures/%s.jpg", instructorFileID)))
}

// uploadExport finishes the export and returns the archive it would have uploaded.
func uploadExport(t *testing.T, ex *utils.Export) []byte {
	var archive []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		archive = body
	}))
	defer server.Close()

	require.NoError(t, ex.UploadTo(context.Background(), server.URL))
	return archive
}

// readZipEntry returns the content of the archive entry at path, or nil if there is none.
func readZipEntry(t *testing.T, archive []byte, path string) []byte {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)
	for _, entry := range reader.File {
		if entry.Name != path {
			continue
		}
		file, err := entry.Open()
		require.NoError(t, err)
		defer func() { _ = file.Close() }()
		content, err := io.ReadAll(file)
		require.NoError(t, err)
		return content
	}
	return nil
}
