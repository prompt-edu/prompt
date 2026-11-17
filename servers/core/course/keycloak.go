package course

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// keycloakTimeout bounds the Keycloak provisioning of a new course, and separately the cleanup after
// it failed. The creating transaction holds a pool connection meanwhile.
const keycloakTimeout = 30 * time.Second

// ErrCourseIdentifierTaken is returned when another course's semester tag and name join to the same
// "<semester>-<name>" identifier, such as "ws24" with "ios-lab" and "ws24-ios" with "lab". The
// Keycloak groups and roles of a course are named after the identifier, so the two would share them.
var ErrCourseIdentifierTaken = errors.New("another course has the same semester and name when joined with a hyphen")

// ProvisionKeycloak creates the Keycloak groups and roles of a course that qtx inserted and has not
// committed yet, and makes the requester its Lecturer. Creating and copying a course both use it.
//
// It first takes a lock on the course's identifier that is held until the transaction ends, and
// refuses an identifier that another course has. From then on, no other course can own Keycloak
// state under this identifier, so the provisioning may reuse or reset what it finds there, and a
// failed provisioning is cleaned up before the rollback releases the identifier.
func ProvisionKeycloak(
	ctx context.Context,
	qtx *db.Queries,
	created db.Course,
	requesterID string,
	createGroupsAndRoles func(ctx context.Context, courseName, iterationName, userID string) error,
	cleanUpGroupsAndRoles func(ctx context.Context, courseName, iterationName string) error,
) error {
	identifier := permissionValidation.CourseIdentifier(created.SemesterTag.String, created.Name)
	if err := qtx.LockCourseIdentifier(ctx, identifier); err != nil {
		return fmt.Errorf("failed to lock the course identifier: %w", err)
	}
	taken, err := qtx.CheckCourseIdentifierTaken(ctx, db.CheckCourseIdentifierTakenParams{Identifier: identifier, CourseID: created.ID})
	if err != nil {
		return fmt.Errorf("failed to check the course identifier: %w", err)
	}
	if taken {
		return fmt.Errorf("%w: %s", ErrCourseIdentifierTaken, identifier)
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	if err := createGroupsAndRoles(keycloakCtx, created.Name, created.SemesterTag.String, requesterID); err != nil {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
		defer cancelCleanup()
		if cleanupErr := cleanUpGroupsAndRoles(cleanupCtx, created.Name, created.SemesterTag.String); cleanupErr != nil {
			log.Error("failed to clean up the keycloak groups and roles of course ", identifier, ": ", cleanupErr)
		}
		return fmt.Errorf("failed to create keycloak groups and roles: %w", err)
	}
	return nil
}
