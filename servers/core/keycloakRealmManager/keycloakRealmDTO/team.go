package keycloakRealmDTO

import "github.com/Nerzal/gocloak/v14"

// TeamMember is a Keycloak user as displayed in the course team management UI.
type TeamMember struct {
	KeycloakUserID string `json:"keycloakUserID"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
}

// CourseTeam is the Lecturer / Editor membership of a single course.
type CourseTeam struct {
	Lecturers []TeamMember `json:"lecturers"`
	Editors   []TeamMember `json:"editors"`
}

// UserSearchResults is the response of the realm-wide Keycloak user search.
// Truncated is true when more matches exist than were returned and the caller
// should narrow their query.
type UserSearchResults struct {
	Results   []TeamMember `json:"results"`
	Truncated bool         `json:"truncated"`
}

// GetTeamMemberFromKeycloakUser maps a gocloak.User to a TeamMember,
// safely handling nil pointer fields by substituting empty strings.
func GetTeamMemberFromKeycloakUser(user *gocloak.User) TeamMember {
	return TeamMember{
		KeycloakUserID: stringOrEmpty(user.ID),
		Username:       stringOrEmpty(user.Username),
		Email:          stringOrEmpty(user.Email),
		FirstName:      stringOrEmpty(user.FirstName),
		LastName:       stringOrEmpty(user.LastName),
	}
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
