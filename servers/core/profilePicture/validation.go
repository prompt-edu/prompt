package profilePicture

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/profilePicture/profilePictureDTO"
)

// maxLookupIDs bounds a single lookup; it covers the largest participant tables with room to spare.
const maxLookupIDs = 1000

func validateLookupRequest(req profilePictureDTO.LookupRequest) error {
	total := len(req.UserIDs) + len(req.StudentIDs) + len(req.CourseParticipationIDs)
	if total > maxLookupIDs {
		return fmt.Errorf("lookup requests at most %d ids, got %d: %w", maxLookupIDs, total, ErrInvalidInput)
	}
	return nil
}

// validateStorageKeyOwner rejects keys that were not presigned for the caller, so nobody can
// claim another user's upload as their picture.
func validateStorageKeyOwner(storageKey string, userID uuid.UUID) error {
	if !strings.HasPrefix(storageKey, ownerStorageKeyPrefix(userID)+"/") {
		return fmt.Errorf("storage key does not belong to the caller: %w", ErrInvalidInput)
	}
	return nil
}

func ownerStorageKeyPrefix(userID uuid.UUID) string {
	return fmt.Sprintf("%s/%s", storageKeyPrefix, userID.String())
}
