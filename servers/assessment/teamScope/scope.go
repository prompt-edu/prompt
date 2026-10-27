package teamScope

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const scopeContextKey = "teamScope"

var ErrOutsideScope = errors.New("access restricted to assigned team")

// Scope is the set of participants the caller may see and edit. Only course editors are
// confined; admins and lecturers get the unrestricted zero value.
type Scope struct {
	confined bool
	members  map[uuid.UUID]struct{}
}

func ConfinedTo(members []uuid.UUID) Scope {
	scope := Scope{confined: true, members: make(map[uuid.UUID]struct{}, len(members))}
	for _, member := range members {
		scope.members[member] = struct{}{}
	}
	return scope
}

func (s Scope) Allows(courseParticipationID uuid.UUID) bool {
	if !s.confined {
		return true
	}
	_, ok := s.members[courseParticipationID]
	return ok
}

func (s Scope) AllowsNobody() bool {
	return s.confined && len(s.members) == 0
}

func (s Scope) AllowsAll(courseParticipationIDs []uuid.UUID) bool {
	for _, id := range courseParticipationIDs {
		if !s.Allows(id) {
			return false
		}
	}
	return true
}

// Get returns the scope resolved by one of this package's middlewares. A route missing them
// gets a scope that allows nobody, so a wiring mistake hides data instead of leaking it.
func Get(c *gin.Context) Scope {
	if scope, ok := c.Get(scopeContextKey); ok {
		if typed, ok := scope.(Scope); ok {
			return typed
		}
	}
	log.Error("team scope middleware is not installed on route ", c.FullPath())
	return Scope{confined: true}
}

func Filter[T any](c *gin.Context, items []T, courseParticipationID func(T) uuid.UUID) []T {
	scope := Get(c)
	if !scope.confined {
		return items
	}
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if scope.Allows(courseParticipationID(item)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
