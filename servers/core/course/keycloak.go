package course

import (
	"context"
	"fmt"
	"time"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// keycloakTimeout bounds the Keycloak provisioning of a new course, and separately the cleanup after
// it failed. The creating transaction holds a pool connection meanwhile.
const keycloakTimeout = 30 * time.Second

// ProvisionKeycloak creates the Keycloak groups and roles of a course that the caller's transaction
// inserted and has not committed yet, and makes the requester its Lecturer. Creating and copying a
// course both use it.
//
// The groups and roles are named after the identifier "<semester>-<name>". ValidateCourseIdentifier
// keeps "-" out of a new course's semester tag and name, so no other course has its identifier, and
// the uncommitted row holds the (name, semester_tag) unique slot until the transaction ends. Keycloak
// state under the identifier can therefore only be left over from a course that no longer exists:
// the provisioning deletes it before it creates the course's own, and a failed provisioning is
// cleaned up before the rollback releases the slot.
func ProvisionKeycloak(
	ctx context.Context,
	created db.Course,
	requesterID string,
	createGroupsAndRoles func(ctx context.Context, courseName, iterationName, userID string) error,
	cleanUpGroupsAndRoles func(ctx context.Context, courseName, iterationName string) error,
) error {
	if err := ValidateCourseIdentifier(created.Name, created.SemesterTag.String); err != nil {
		return err
	}

	keycloakCtx, cancel := context.WithTimeout(ctx, keycloakTimeout)
	defer cancel()
	if err := createGroupsAndRoles(keycloakCtx, created.Name, created.SemesterTag.String, requesterID); err != nil {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), keycloakTimeout)
		defer cancelCleanup()
		if cleanupErr := cleanUpGroupsAndRoles(cleanupCtx, created.Name, created.SemesterTag.String); cleanupErr != nil {
			log.Error("failed to clean up the keycloak groups and roles of course ", created.SemesterTag.String, "-", created.Name, ": ", cleanupErr)
		}
		return fmt.Errorf("failed to create keycloak groups and roles: %w", err)
	}
	return nil
}
