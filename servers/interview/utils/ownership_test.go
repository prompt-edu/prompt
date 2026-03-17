package utils

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetUserCourseParticipationID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testUUID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	c, _ := gin.CreateTestContext(nil)
	c.Set("courseParticipationID", testUUID)

	result, err := GetUserCourseParticipationID(c)
	require.NoError(t, err)
	require.Equal(t, testUUID, result)
}

func TestValidateStudentOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUUID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	c, _ := gin.CreateTestContext(nil)
	c.Set("courseParticipationID", userUUID)

	status, err := ValidateStudentOwnership(c, userUUID)
	require.NoError(t, err)
	require.Equal(t, 200, status)
}
