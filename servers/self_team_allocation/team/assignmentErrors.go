package teams

import (
	"errors"
	"net/http"
)

// assignTeamErrorStatus maps AssignTeam errors to the HTTP status of the join route.
func assignTeamErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrTeamFull):
		return http.StatusBadRequest
	case errors.Is(err, ErrTeamNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
