package keycloakRealmDTO

import "github.com/prompt-edu/prompt/servers/core/student/studentDTO"

type GroupMembers struct {
	Students    []studentDTO.Student `json:"students"`
	NonStudents []KeycloakUser       `json:"nonStudents"`
}
