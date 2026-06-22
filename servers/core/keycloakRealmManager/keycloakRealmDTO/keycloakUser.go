package keycloakRealmDTO

import "github.com/Nerzal/gocloak/v14"

type KeycloakUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func GetKeycloakUserDTO(user gocloak.User) KeycloakUser {
	return KeycloakUser{
		Username:  *user.Username,
		Email:     *user.Email,
		FirstName: *user.FirstName,
		LastName:  *user.LastName,
	}
}
