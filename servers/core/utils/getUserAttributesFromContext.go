package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
)

func GetUserUUIDFromContext(c *gin.Context) (uuid.UUID, error) {
	return parseUUIDString(c.GetString(keycloakTokenVerifier.CtxUserID))
}

func GetUserEmailFromContext(c *gin.Context) string {
	return c.GetString(keycloakTokenVerifier.CtxUserEmail)
}

func GetUserNameFromContext(c *gin.Context) (string) {
	firstName := c.GetString(keycloakTokenVerifier.CtxFirstName)
	lastName := c.GetString(keycloakTokenVerifier.CtxLastName)
	return firstName + " " + lastName
}

func GetUniversityLoginFromContext(c*gin.Context) (string) {
  return c.GetString(keycloakTokenVerifier.CtxUniversityLogin)
}

func GetMatriculationNumberFromContext(c*gin.Context) (string) {
  return c.GetString(keycloakTokenVerifier.CtxMatriculationNumber)
}

func parseUUIDString(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}
