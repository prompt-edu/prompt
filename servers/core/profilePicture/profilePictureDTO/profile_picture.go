package profilePictureDTO

import (
	"time"

	"github.com/google/uuid"
)

// PresignedUpload is the target the client uploads the cropped picture to before completing it.
type PresignedUpload struct {
	UploadURL  string `json:"uploadUrl"`
	StorageKey string `json:"storageKey"`
}

// CompleteUpload makes an uploaded picture the caller's profile picture.
type CompleteUpload struct {
	StorageKey string `json:"storageKey" binding:"required"`
}

// ProfilePicture is the caller's own profile picture.
type ProfilePicture struct {
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LookupRequest asks for the pictures of several people at once. Every id kind is optional.
type LookupRequest struct {
	UserIDs                []uuid.UUID `json:"userIds"`
	StudentIDs             []uuid.UUID `json:"studentIds"`
	CourseParticipationIDs []uuid.UUID `json:"courseParticipationIds"`
}

// ProfilePictureURLs maps each requested id that has a picture to a presigned download URL.
// Ids without a picture are omitted.
type ProfilePictureURLs struct {
	Users                map[uuid.UUID]string `json:"users"`
	Students             map[uuid.UUID]string `json:"students"`
	CourseParticipations map[uuid.UUID]string `json:"courseParticipations"`
}
