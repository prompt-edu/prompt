package keycloakRealmDTO

// OrgMember is a user in the Admin or Member group of an org.
type OrgMember struct {
	StaffMember
	// Role is the user's org role, Admin or Member.
	Role string `json:"role" enums:"Admin,Member"`
}
