package utils

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	log "github.com/sirupsen/logrus"
)

// RespondWithDBError maps a service/database error to an HTTP response: a missing
// row (pgx.ErrNoRows / sql.ErrNoRows) becomes 404, everything else 500.
func RespondWithDBError(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, sdkUtils.ErrorResponse{Error: "resource not found"})
		return
	}
	log.Error(err)
	c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "internal server error"})
}
