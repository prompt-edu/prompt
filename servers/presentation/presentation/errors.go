package presentation

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type APIError struct {
	Status  int
	Code    string
	Message string
	Details any
	Err     error
}

func (e *APIError) Error() string { return e.Message }
func (e *APIError) Unwrap() error { return e.Err }

func apiError(status int, code, message string, err error) *APIError {
	return &APIError{Status: status, Code: code, Message: message, Err: err}
}

func writeError(c *gin.Context, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		body := gin.H{"error": apiErr.Message, "message": apiErr.Message, "code": apiErr.Code}
		if apiErr.Details != nil {
			if apiErr.Code == "feedback_conflict" {
				body["answer"] = apiErr.Details
			} else {
				body["details"] = apiErr.Details
			}
		}
		c.JSON(apiErr.Status, body)
		return
	}
	log.WithError(err).Error("Unexpected presentation API error")
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "code": "internal_error"})
}
