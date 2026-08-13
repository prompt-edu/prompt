package mailingDTO

type MailingReport struct {
	SuccessfulEmails []string `json:"successfulEmails"`
	FailedEmails     []string `json:"failedEmails"`
}
