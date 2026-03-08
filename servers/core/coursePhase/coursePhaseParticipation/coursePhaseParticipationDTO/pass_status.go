package coursePhaseParticipationDTO

import db "github.com/prompt-edu/prompt/servers/core/db/sqlc"

func GetPassStatusString(passStatus db.NullPassStatus) string {
	if passStatus.Valid {
		return string(passStatus.PassStatus)
	}
	return string(db.PassStatusNotAssessed)
}

func GetPassStatusDBModel(passStatus *db.PassStatus) db.NullPassStatus {
	if passStatus == nil {
		return db.NullPassStatus{
			Valid: false,
		}
	}
	return db.NullPassStatus{
		Valid:      true,
		PassStatus: *passStatus,
	}
}
