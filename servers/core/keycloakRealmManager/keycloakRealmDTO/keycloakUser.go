package keycloakRealmDTO

import "github.com/Nerzal/gocloak/v14"

type KeycloakUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func GetKeycloakUserDTO(user gocloak.User) KeycloakUser {
	return KeycloakUser{
		Username:  stringValue(user.Username),
		Email:     stringValue(user.Email),
		FirstName: stringValue(user.FirstName),
		LastName:  stringValue(user.LastName),
	}
}
