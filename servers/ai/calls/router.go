package calls

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

func RegisterRoutes(coursePhaseAPI *gin.RouterGroup, service *Service) {
	callAPI := coursePhaseAPI.Group("/calls")
	adminOnly := promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin)
	callAPI.GET("", adminOnly, service.listCalls)
	callAPI.GET("/:callID", adminOnly, service.getCall)
	callAPI.POST("/:callID/events",
		promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor),
		service.addEvent)
}

func (s *Service) listCalls(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return
	}
	limit, cursor, err := parsePaging(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: err.Error()})
		return
	}
	page, err := s.List(c.Request.Context(), coursePhaseID, cursor, limit)
	if err != nil {
		log.WithError(err).Error("Could not list AI calls")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not list calls"})
		return
	}
	c.JSON(http.StatusOK, page)
}

func (s *Service) getCall(c *gin.Context) {
	coursePhaseID, callID, ok := parseIDs(c)
	if !ok {
		return
	}
	user, _ := keycloakTokenVerifier.GetTokenUser(c)
	detail, err := s.Get(c.Request.Context(), coursePhaseID, callID, user.ID)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, sdkUtils.ErrorResponse{Error: "call not found"})
		return
	}
	if err != nil {
		log.WithError(err).Error("Could not load AI call")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not load the call"})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (s *Service) addEvent(c *gin.Context) {
	coursePhaseID, callID, ok := parseIDs(c)
	if !ok {
		return
	}
	var request EventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "type must be shown, accepted, edited or rejected"})
		return
	}
	user, _ := keycloakTokenVerifier.GetTokenUser(c)
	event, err := s.AddEvent(c.Request.Context(), coursePhaseID, callID, user.ID, user.Roles[promptSDK.PromptAdmin], request)
	switch {
	case errors.Is(err, ErrInvalidEvent):
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: err.Error()})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, sdkUtils.ErrorResponse{Error: "call not found"})
	case err != nil:
		log.WithError(err).Error("Could not record AI call event")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not record the event"})
	default:
		c.JSON(http.StatusCreated, event)
	}
}

func parseIDs(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return uuid.Nil, uuid.Nil, false
	}
	callID, err := uuid.Parse(c.Param("callID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid call id"})
		return uuid.Nil, uuid.Nil, false
	}
	return coursePhaseID, callID, true
}

func parsePaging(c *gin.Context) (int32, *Cursor, error) {
	limit := defaultPageSize
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxPageSize {
			return 0, nil, errors.New("limit must be between 1 and 200")
		}
		limit = parsed
	}
	rawTime, rawID := c.Query("cursorRequestedAt"), c.Query("cursorId")
	if rawTime == "" && rawID == "" {
		return int32(limit), nil, nil
	}
	requestedAt, timeErr := time.Parse(time.RFC3339Nano, rawTime)
	id, idErr := uuid.Parse(rawID)
	if timeErr != nil || idErr != nil {
		return 0, nil, errors.New("cursorRequestedAt and cursorId must be given together")
	}
	return int32(limit), &Cursor{RequestedAt: requestedAt, ID: id}, nil
}
