package mailingDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type MailingInfo struct {
	CourseName      string
	CourseStartDate pgtype.Date
	CourseEndDate   pgtype.Date
	MailSubject     string
	MailContent     string
}

func GetMailingInfoFromFailedMailingInformation(infos db.GetFailedMailingInformationRow) MailingInfo {
	return MailingInfo{
		CourseName:      infos.CourseName,
		CourseStartDate: infos.CourseStartDate,
		CourseEndDate:   infos.CourseEndDate,
		MailSubject:     infos.MailSubject,
		MailContent:     infos.MailContent,
	}
}

func GetMailingInfoFromPassedMailingInformation(infos db.GetPassedMailingInformationRow) MailingInfo {
	return MailingInfo{
		CourseName:      infos.CourseName,
		CourseStartDate: infos.CourseStartDate,
		CourseEndDate:   infos.CourseEndDate,
		MailSubject:     infos.MailSubject,
		MailContent:     infos.MailContent,
	}
}
