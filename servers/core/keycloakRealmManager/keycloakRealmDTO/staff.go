package keycloakRealmDTO

import (
	"regexp"

	"github.com/Nerzal/gocloak/v14"
)

// StaffMember is a Keycloak user as displayed in the course staff management UI.
type StaffMember struct {
	KeycloakUserID string `json:"keycloakUserID"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	// UniversityLogin is the login phases match tutors against: the university_login attribute, or
	// a username in the TUM ID format when the attribute is missing. Otherwise it is empty.
	UniversityLogin string `json:"universityLogin"`
}

// UniversityLoginAttribute is the Keycloak user attribute that also feeds the token claim.
const UniversityLoginAttribute = "university_login"

// The production admin API returns TUM accounts with the TUM ID as username but no login attribute.
var tumIDPattern = regexp.MustCompile(`^[a-z]{2}[0-9]{2}[a-z]{3}$`)

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
	Truncated bool          `json:"truncated"`
}

// GetStaffMemberFromKeycloakUser maps a gocloak.User to a StaffMember,
// safely handling nil pointer fields by substituting empty strings.
func GetStaffMemberFromKeycloakUser(user *gocloak.User) StaffMember {
	return StaffMember{
		KeycloakUserID:  gocloak.PString(user.ID),
		Username:        gocloak.PString(user.Username),
		Email:           gocloak.PString(user.Email),
		FirstName:       gocloak.PString(user.FirstName),
		LastName:        gocloak.PString(user.LastName),
		UniversityLogin: universityLogin(user),
	}
}

func universityLogin(user *gocloak.User) string {
	if login := firstAttribute(user, UniversityLoginAttribute); login != "" {
		return login
	}
	if username := gocloak.PString(user.Username); tumIDPattern.MatchString(username) {
		return username
	}
	return ""
}

func firstAttribute(user *gocloak.User, name string) string {
	values := user.Attributes[name]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
