package service

import (
	"fmt"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

func MakeUniqueFileNameWithEnding(answerWithFile db.GetAllApplicationAnswersFileUploadWithFileRecordByCourseParticipationIDsRow) string {
  shortID := answerWithFile.ID.String()[:6]
  return fmt.Sprintf("%s-%s", shortID, answerWithFile.Filename)
}
