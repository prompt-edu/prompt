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

func TestGetStaffMemberFromKeycloakUserFallsBackToTUMIDUsername(t *testing.T) {
	member := GetStaffMemberFromKeycloakUser(&gocloak.User{Username: gocloak.StringP("ge95may")})
	assert.Equal(t, "ge95may", member.UniversityLogin)

	attributeWins := GetStaffMemberFromKeycloakUser(&gocloak.User{
		Username:   gocloak.StringP("ab12cde"),
		Attributes: map[string][]string{"university_login": {"co69edt"}},
	})
	assert.Equal(t, "co69edt", attributeWins.UniversityLogin)

	for _, username := range []string{"course-editor", "GE95MAY", "ge95mayx", "g195may"} {
		assert.Empty(t, GetStaffMemberFromKeycloakUser(&gocloak.User{Username: gocloak.StringP(username)}).UniversityLogin, username)
	}
}
