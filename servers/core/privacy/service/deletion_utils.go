package service

import (
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

func subrequestStatusFromError(err error) db.PrivacyDeletionSubrequestStatus {
	if err != nil {
		return db.PrivacyDeletionSubrequestStatusFailed
	}
	return db.PrivacyDeletionSubrequestStatusSucceeded
}
