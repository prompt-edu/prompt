package profilePicture

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/profilePicture/profilePictureDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func authenticateAs(userID string, universityLogin string) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(keycloakTokenVerifier.CtxUserID, userID)
			c.Set(keycloakTokenVerifier.CtxUniversityLogin, universityLogin)
			c.Set(keycloakTokenVerifier.CtxUserEmail, "someone@tum.de")
			c.Next()
		}
	}
}

func rejectUserWithoutRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User has no roles"})
	}
}

func (suite *ProfilePictureServiceTestSuite) newRouter(authMiddleware, applicationMiddleware func() gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/api"), suite.service, authMiddleware, applicationMiddleware)
	return router
}

func (suite *ProfilePictureServiceTestSuite) routerForUser(userID string, universityLogin string) *gin.Engine {
	identify := authenticateAs(userID, universityLogin)
	return suite.newRouter(identify, identify)
}

func serveJSON(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var payload bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&payload).Encode(body)
	}
	req := httptest.NewRequest(method, path, &payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func (suite *ProfilePictureServiceTestSuite) TestRouter_UploadReadLookupDelete() {
	userID := uuid.New()
	router := suite.routerForUser(userID.String(), "")

	w := serveJSON(router, http.MethodPost, "/api/profile-pictures/me/presign", nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	var presigned profilePictureDTO.PresignedUpload
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &presigned))
	suite.store.put(presigned.StorageKey, pictureContentType, validJPEG)

	w = serveJSON(router, http.MethodPost, "/api/profile-pictures/me/complete", profilePictureDTO.CompleteUpload{StorageKey: presigned.StorageKey})
	require.Equal(suite.T(), http.StatusOK, w.Code, w.Body.String())
	storedKey := suite.storedKey(userID)

	w = serveJSON(router, http.MethodGet, "/api/profile-pictures/me", nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	var own profilePictureDTO.ProfilePicture
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &own))
	assert.Contains(suite.T(), own.URL, storedKey)

	w = serveJSON(router, http.MethodPost, "/api/profile-pictures/lookup", profilePictureDTO.LookupRequest{UserIDs: []uuid.UUID{userID}})
	require.Equal(suite.T(), http.StatusOK, w.Code)
	var urls profilePictureDTO.ProfilePictureURLs
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &urls))
	assert.Contains(suite.T(), urls.Users[userID], storedKey)

	w = serveJSON(router, http.MethodDelete, "/api/profile-pictures/me", nil)
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	w = serveJSON(router, http.MethodGet, "/api/profile-pictures/me", nil)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *ProfilePictureServiceTestSuite) TestRouter_FirstTimeApplicantManagesOwnPicture() {
	userID := uuid.New()
	router := suite.newRouter(rejectUserWithoutRoles, authenticateAs(userID.String(), "ap02pli"))

	w := serveJSON(router, http.MethodPost, "/api/profile-pictures/me/presign", nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	var presigned profilePictureDTO.PresignedUpload
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &presigned))
	suite.store.put(presigned.StorageKey, pictureContentType, validJPEG)

	w = serveJSON(router, http.MethodPost, "/api/profile-pictures/me/complete", profilePictureDTO.CompleteUpload{StorageKey: presigned.StorageKey})
	require.Equal(suite.T(), http.StatusOK, w.Code, w.Body.String())

	w = serveJSON(router, http.MethodGet, "/api/profile-pictures/me", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	w = serveJSON(router, http.MethodPost, "/api/profile-pictures/lookup", profilePictureDTO.LookupRequest{UserIDs: []uuid.UUID{userID}})
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code, "looking up pictures still needs a role")

	w = serveJSON(router, http.MethodDelete, "/api/profile-pictures/me", nil)
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *ProfilePictureServiceTestSuite) TestRouter_CompleteRejectsInvalidRequests() {
	router := suite.routerForUser(uuid.New().String(), "")
	foreignKey := suite.upload(uuid.New(), pictureContentType, validJPEG)

	w := serveJSON(router, http.MethodPost, "/api/profile-pictures/me/complete", map[string]string{})
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code, "a missing storage key is rejected")

	w = serveJSON(router, http.MethodPost, "/api/profile-pictures/me/complete", profilePictureDTO.CompleteUpload{StorageKey: foreignKey})
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code, "another user's upload is rejected")
}

func (suite *ProfilePictureServiceTestSuite) TestRouter_RequiresUserID() {
	router := suite.routerForUser("", "")

	w := serveJSON(router, http.MethodGet, "/api/profile-pictures/me", nil)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}
