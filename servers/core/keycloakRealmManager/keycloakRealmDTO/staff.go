package keycloakRealmDTO

import "github.com/Nerzal/gocloak/v14"

// StaffMember is a Keycloak user as displayed in the course staff management UI.
type StaffMember struct {
	KeycloakUserID string `json:"keycloakUserID"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
}

// CourseStaff is the Lecturer / Editor membership of a single course.
type CourseStaff struct {
	Lecturers []StaffMember `json:"lecturers"`
	Editors   []StaffMember `json:"editors"`
}

// UserSearchResults is the response of the realm-wide Keycloak user search.
// Truncated is true when more matches exist than were returned and the caller
// should narrow their query.
type UserSearchResults struct {
	Results   []StaffMember `json:"results"`
	Truncated bool         `json:"truncated"`
}

// GetStaffMemberFromKeycloakUser maps a gocloak.User to a StaffMember,
// safely handling nil pointer fields by substituting empty strings.
func GetStaffMemberFromKeycloakUser(user *gocloak.User) StaffMember {
	return StaffMember{
		KeycloakUserID: gocloak.PString(user.ID),
		Username:       gocloak.PString(user.Username),
		Email:          gocloak.PString(user.Email),
		FirstName:      gocloak.PString(user.FirstName),
		LastName:       gocloak.PString(user.LastName),
	}
}
