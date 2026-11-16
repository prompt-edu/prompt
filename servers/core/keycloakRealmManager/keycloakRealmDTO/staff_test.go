package keycloakRealmDTO

import (
	"testing"

	"github.com/Nerzal/gocloak/v14"
	"github.com/stretchr/testify/assert"
)

func TestGetStaffMemberFromKeycloakUserReadsUniversityLogin(t *testing.T) {
	user := &gocloak.User{
		ID:         gocloak.StringP("4b09837e-367a-4e78-a982-176a22751fba"),
		Username:   gocloak.StringP("course-editor"),
		Email:      gocloak.StringP("editor@example.com"),
		FirstName:  gocloak.StringP("Eduard"),
		LastName:   gocloak.StringP("Editor"),
		Attributes: map[string][]string{"university_login": {"co69edt"}},
	}

	member := GetStaffMemberFromKeycloakUser(user)

	assert.Equal(t, "4b09837e-367a-4e78-a982-176a22751fba", member.KeycloakUserID)
	assert.Equal(t, "course-editor", member.Username)
	assert.Equal(t, "co69edt", member.UniversityLogin)
}

func TestGetStaffMemberFromKeycloakUserWithoutUniversityLogin(t *testing.T) {
	withoutAttributes := GetStaffMemberFromKeycloakUser(&gocloak.User{Username: gocloak.StringP("guest")})
	assert.Empty(t, withoutAttributes.UniversityLogin)

	emptyLogin := GetStaffMemberFromKeycloakUser(&gocloak.User{
		Username:   gocloak.StringP("guest"),
		Attributes: map[string][]string{"university_login": {}},
	})
	assert.Empty(t, emptyLogin.UniversityLogin)
}
