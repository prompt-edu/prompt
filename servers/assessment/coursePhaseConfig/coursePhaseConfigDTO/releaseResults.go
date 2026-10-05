package coursePhaseConfigDTO

type ResultsReleasedMailReport struct {
	SuccessfulEmails    []string `json:"successfulEmails"`
	FailedEmails        []string `json:"failedEmails"`
	RequestedRecipients int      `json:"requestedRecipients"`
}

type ReleaseResultsResponse struct {
	Message    string                     `json:"message"`
	MailReport *ResultsReleasedMailReport `json:"mailReport,omitempty"`
	MailError  string                     `json:"mailError,omitempty"`
}
