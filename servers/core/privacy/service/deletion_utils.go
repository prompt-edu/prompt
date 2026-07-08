package service

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

func subrequestStatusFromError(err error) db.PrivacyDeletionSubrequestStatus {
	if err != nil {
		return db.PrivacyDeletionSubrequestStatusFailed
	}
	return db.PrivacyDeletionSubrequestStatusSucceeded
}

func UniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
