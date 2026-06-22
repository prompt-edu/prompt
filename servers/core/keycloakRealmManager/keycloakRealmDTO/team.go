package keycloakRealmDTO

import "github.com/Nerzal/gocloak/v13"

type TeamMember struct {
	KeycloakUserID string `json:"keycloakUserID"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
}

type CourseTeam struct {
	Lecturers []TeamMember `json:"lecturers"`
	Editors   []TeamMember `json:"editors"`
}

type UserSearchResults struct {
	Results   []TeamMember `json:"results"`
	Truncated bool         `json:"truncated"`
}

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
