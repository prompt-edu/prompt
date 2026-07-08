package utils

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

// RespondWithDBError maps a service/database error to an HTTP response: a missing
// row (pgx.ErrNoRows / sql.ErrNoRows) becomes 404, everything else 500.
func RespondWithDBError(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "resource not found"})
		return
	}
	log.Error(err)
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
}
