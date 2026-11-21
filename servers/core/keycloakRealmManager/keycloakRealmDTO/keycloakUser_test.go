package keycloakRealmDTO

import (
	"testing"

	"github.com/Nerzal/gocloak/v14"
	"github.com/stretchr/testify/assert"
)

func TestGetKeycloakUserDTO_AllFieldsPopulated(t *testing.T) {
	user := gocloak.User{
		Username:  gocloak.StringP("jdoe"),
		Email:     gocloak.StringP("jdoe@example.com"),
		FirstName: gocloak.StringP("John"),
		LastName:  gocloak.StringP("Doe"),
	}

	dto := GetKeycloakUserDTO(user)

	assert.Equal(t, "jdoe", dto.Username)
	assert.Equal(t, "jdoe@example.com", dto.Email)
	assert.Equal(t, "John", dto.FirstName)
	assert.Equal(t, "Doe", dto.LastName)
}

func TestGetKeycloakUserDTO_NilFieldsSafe(t *testing.T) {
	// Keycloak omits firstName, lastName, or email when empty, resulting in nil pointers
	user := gocloak.User{
		Username:  gocloak.StringP("guest"),
		Email:     nil,
		FirstName: nil,
		LastName:  nil,
	}

	dto := GetKeycloakUserDTO(user)

	assert.Equal(t, "guest", dto.Username)
	assert.Equal(t, "", dto.Email)
	assert.Equal(t, "", dto.FirstName)
	assert.Equal(t, "", dto.LastName)
}

func TestGetKeycloakUserDTO_AllNilFieldsSafe(t *testing.T) {
	user := gocloak.User{}

	dto := GetKeycloakUserDTO(user)

	assert.Equal(t, "", dto.Username)
	assert.Equal(t, "", dto.Email)
	assert.Equal(t, "", dto.FirstName)
	assert.Equal(t, "", dto.LastName)
}
