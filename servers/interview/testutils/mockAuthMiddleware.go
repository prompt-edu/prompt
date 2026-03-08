package testutils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/sirupsen/logrus"
)

type MockIdentity struct {
	Roles                 []string
	Email                 string
	MatriculationNumber   string
	UniversityLogin       string
	CourseParticipationID uuid.UUID
	FirstName             string
	LastName              string
}

func defaultIdentity() MockIdentity {
	return MockIdentity{
		Roles:                 []string{},
		Email:                 "student@example.com",
		MatriculationNumber:   "0000000",
		UniversityLogin:       "st12345",
		CourseParticipationID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
		FirstName:             "Test",
		LastName:              "Student",
	}
}

func mergeIdentity(custom MockIdentity) MockIdentity {
	defaults := defaultIdentity()

	if custom.Email != "" {
		defaults.Email = custom.Email
	}
	if custom.MatriculationNumber != "" {
		defaults.MatriculationNumber = custom.MatriculationNumber
	}
	if custom.UniversityLogin != "" {
		defaults.UniversityLogin = custom.UniversityLogin
	}
	if custom.CourseParticipationID != uuid.Nil {
		defaults.CourseParticipationID = custom.CourseParticipationID
	}
	if custom.FirstName != "" {
		defaults.FirstName = custom.FirstName
	}
	if custom.LastName != "" {
		defaults.LastName = custom.LastName
	}
	if len(custom.Roles) > 0 {
		defaults.Roles = custom.Roles
	}

	return defaults
}

func NewMockAuthMiddleware(identity MockIdentity) func(allowedRoles ...string) gin.HandlerFunc {
	merged := mergeIdentity(identity)

	return func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			userRoles := make(map[string]bool)
			for _, role := range merged.Roles {
				userRoles[role] = true
			}

			// Create and set token user using the SDK's TokenUser type
			tokenUser := keycloakTokenVerifier.TokenUser{
				Roles:      userRoles,
				IsLecturer: false,
				IsEditor:   false,
			}
			c.Set("tokenUser", tokenUser)

			c.Set("userRoles", userRoles)
			c.Set("userEmail", merged.Email)
			c.Set("matriculationNumber", merged.MatriculationNumber)
			c.Set("universityLogin", merged.UniversityLogin)
			c.Set("courseParticipationID", merged.CourseParticipationID)
			c.Set("firstName", merged.FirstName)
			c.Set("lastName", merged.LastName)

			logrus.WithFields(logrus.Fields{
				"email": merged.Email,
				"roles": merged.Roles,
			}).Info("Using mock authentication middleware")

			c.Next()
		}
	}
}

func DefaultMockAuthMiddleware() func(allowedRoles ...string) gin.HandlerFunc {
	return NewMockAuthMiddleware(MockIdentity{})
}
