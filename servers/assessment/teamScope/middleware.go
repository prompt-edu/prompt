package teamScope

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	log "github.com/sirupsen/logrus"
)

// Resolve stores the caller's scope for handlers that filter lists with Filter.
func Resolve() gin.HandlerFunc {
	return func(c *gin.Context) {
		ensureScope(c)
	}
}

// RequireParticipant rejects requests whose :courseParticipationID is outside the caller's scope.
func RequireParticipant() gin.HandlerFunc {
	return func(c *gin.Context) {
		scope, ok := ensureScope(c)
		if !ok {
			return
		}
		courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !scope.Allows(courseParticipationID) {
			abortOutsideScope(c)
		}
	}
}

// RequireBodyParticipants rejects writes where the courseParticipationID or any of the
// courseParticipationIDs is outside the caller's scope. Both fields are checked because a
// handler binds whichever one its DTO declares.
func RequireBodyParticipants() gin.HandlerFunc {
	return func(c *gin.Context) {
		scope, ok := ensureScope(c)
		if !ok || !scope.confined {
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		var payload struct {
			CourseParticipationID  uuid.UUID   `json:"courseParticipationID"`
			CourseParticipationIDs []uuid.UUID `json:"courseParticipationIDs"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		targets := payload.CourseParticipationIDs
		if payload.CourseParticipationID != uuid.Nil || len(targets) == 0 {
			targets = append(targets, payload.CourseParticipationID)
		}
		if !scope.AllowsAll(targets) {
			abortOutsideScope(c)
		}
	}
}

func abortOutsideScope(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrOutsideScope.Error()})
}

func ensureScope(c *gin.Context) (Scope, bool) {
	if scope, ok := c.Get(scopeContextKey); ok {
		if typed, ok := scope.(Scope); ok {
			return typed, true
		}
	}

	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authenticated user not found in context"})
		return Scope{}, false
	}

	scope := Scope{}
	if tokenUser.IsEditor && !tokenUser.IsLecturer && !tokenUser.Roles[promptSDK.PromptAdmin] {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return Scope{}, false
		}
		scope, err = resolveEditorScope(c.GetHeader("Authorization"), coursePhaseID)
		if err != nil {
			log.WithError(err).Error("could not resolve the editor's team scope")
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "could not resolve the caller's team"})
			return Scope{}, false
		}
	}

	c.Set(scopeContextKey, scope)
	return scope, true
}
