package mailingDTO

import (
	"encoding/json"
	"errors"
	"net/mail"

	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CourseMailingSettings struct {
	ReplyTo mail.Address
	CC      []mail.Address
	BCC     []mail.Address
}

type MailItem struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetCourseMailingSettingsFromDBModel(dbModel db.GetCourseMailingSettingsForCoursePhaseIDRow) (CourseMailingSettings, error) {
	var ccNames []MailItem
	var bccNames []MailItem
	err := json.Unmarshal([]byte(dbModel.CcAddresses), &ccNames)
	if err != nil {
		return CourseMailingSettings{}, errors.New("failed to unmarshal cc addresses")
	}

	err = json.Unmarshal([]byte(dbModel.BccAddresses), &bccNames)
	if err != nil {
		return CourseMailingSettings{}, errors.New("failed to unmarshal bcc addresses")
	}

	// map to mail.Address
	var ccAddresses []mail.Address
	for _, cc := range ccNames {
		ccAddresses = append(ccAddresses, mail.Address{
			Name:    cc.Name,
			Address: cc.Email,
		})
	}

	var bccAddresses []mail.Address
	for _, bcc := range bccNames {
		bccAddresses = append(bccAddresses, mail.Address{
			Name:    bcc.Name,
			Address: bcc.Email,
		})
	}

	return CourseMailingSettings{
		ReplyTo: mail.Address{
			Name:    dbModel.ReplyToName,
			Address: dbModel.ReplyToEmail,
		},
		CC:  ccAddresses,
		BCC: bccAddresses,
	}, nil
}
