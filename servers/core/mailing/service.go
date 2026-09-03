package mailing

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	log "github.com/sirupsen/logrus"
)

type MailingService struct {
	smtpHost       string
	smtpPort       string
	smtpUsername   string
	smtpPassword   string
	senderEmail    mail.Address
	clientURL      string
	queries        db.Queries
	sendMail       func(mailingDTO.CourseMailingSettings, string, string, string) error
	now            func() time.Time
	sendManualMail func(context.Context, uuid.UUID, mailingDTO.SendManualMailRequest) (mailingDTO.ManualMailReport, error)
}

func NewMailingService(queries db.Queries, smtpHost, smtpPort, smtpUsername, smtpPassword, senderName, senderEmail, clientURL string) *MailingService {
	service := &MailingService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		senderEmail:  mail.Address{Name: senderName, Address: senderEmail},
		clientURL:    clientURL,
		queries:      queries,
	}
	service.sendMail = service.SendCourseMail
	service.now = func() time.Time { return time.Now().UTC() }
	service.sendManualMail = service.SendManualMailToParticipants
	return service
}

func (s *MailingService) SendApplicationConfirmationMail(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) (bool, error) {
	isApplicationPhase, err := s.queries.CheckIfCoursePhaseIsApplicationPhase(ctx, coursePhaseID)
	if err != nil {
		return false, fmt.Errorf("failed to verify if course phase %s is an application phase: %v", coursePhaseID, err)
	}
	if !isApplicationPhase {
		return false, fmt.Errorf("course phase %s is not an application phase, cannot send confirmation mail", coursePhaseID)
	}

	mailingInfo, err := s.queries.GetConfirmationMailingInformation(ctx, db.GetConfirmationMailingInformationParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})

	if err != nil {
		log.Error("failed to get mailing information: ", err)
		return false, fmt.Errorf("failed to retrieve confirmation mailing information for course participation %s in phase %s: %v", courseParticipationID, coursePhaseID, err)
	}

	if !mailingInfo.SendConfirmationMail {
		log.Debug("not sending because SendConfirmationMail is disabled")
		return false, nil
	}

	if mailingInfo.ConfirmationMailContent == "" {
		log.Error("mailing template is not correctly configured")
		return false, fmt.Errorf("confirmation mail template is empty for course phase %s", coursePhaseID)
	}

	courseMailingSettings, err := s.getSenderInformation(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to get sender information")
		return false, fmt.Errorf("failed to get sender information for course phase %s: %v", coursePhaseID, err)
	}

	log.Info("Sending confirmation mail to ", mailingInfo.Email.String)

	applicationURL := fmt.Sprintf("%s/apply/%s", s.clientURL, coursePhaseID.String())
	placeholderValues := getApplicationConfirmationPlaceholderValues(mailingInfo, applicationURL)
	finalMessage := replacePlaceholders(mailingInfo.ConfirmationMailContent, placeholderValues)

	// replace values in subject
	finalSubject := replacePlaceholders(mailingInfo.ConfirmationMailSubject, placeholderValues)

	err = s.SendCourseMail(courseMailingSettings, mailingInfo.Email.String, finalSubject, finalMessage)
	if err != nil {
		log.Error("failed to send confirmation mail: ", err)
		return false, fmt.Errorf("failed to send confirmation mail to %s: %v", mailingInfo.Email.String, err)
	}

	return true, nil
}

func (s *MailingService) SendStatusMailManualTrigger(ctx context.Context, coursePhaseID uuid.UUID, status db.PassStatus, recipientCourseParticipationIDs []uuid.UUID) (mailingDTO.MailingReport, error) {
	response := mailingDTO.MailingReport{
		SuccessfulEmails: make([]string, 0),
		FailedEmails:     make([]string, 0),
	}
	mailingInfo := mailingDTO.MailingInfo{}

	// 1.) get mailing info for course phase
	switch status {
	case db.PassStatusPassed:
		infos, err := s.queries.GetPassedMailingInformation(ctx, coursePhaseID)
		if err != nil {
			log.Error("failed to get mailing information: ", err)
			return response, fmt.Errorf("failed to retrieve passed status mailing information for course phase %s: %v", coursePhaseID, err)
		}
		mailingInfo = mailingDTO.GetMailingInfoFromPassedMailingInformation(infos)

	case db.PassStatusFailed:
		infos, err := s.queries.GetFailedMailingInformation(ctx, coursePhaseID)
		if err != nil {
			log.Error("failed to get mailing information: ", err)
			return response, fmt.Errorf("failed to retrieve failed status mailing information for course phase %s: %v", coursePhaseID, err)
		}
		mailingInfo = mailingDTO.GetMailingInfoFromFailedMailingInformation(infos)

	default:
		log.Error("invalid status")
		return response, fmt.Errorf("invalid pass status '%s': expected 'passed' or 'failed'", status)

	}

	// Get the course mailing settings
	courseMailingSettings, err := s.getSenderInformation(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to get sender information")
		return response, fmt.Errorf("failed to get course mailing settings: %v", err)
	}

	// 2.) Check if mailing is configured -> return if not
	if mailingInfo.MailSubject == "" || mailingInfo.MailContent == "" {
		log.Error("mailing template is not correctly configured")
		return response, fmt.Errorf("mailing template incomplete: subject ('%s') or content ('%s') is empty", mailingInfo.MailSubject, mailingInfo.MailContent)
	}

	// 3.) Claim the participants with the given status that have not been mailed for it yet, narrowed
	// to the recipient list if one was given. Claiming before sending trades a mail lost to a crash
	// for a duplicated one.
	var recipientIDs []uuid.UUID
	if recipientCourseParticipationIDs != nil {
		recipientIDs = deduplicateUUIDList(recipientCourseParticipationIDs)
	}

	participants, err := s.queries.ClaimStatusMailRecipients(ctx, db.ClaimStatusMailRecipientsParams{
		CoursePhaseID:          coursePhaseID,
		Status:                 string(status),
		SentAt:                 s.now().Format(time.RFC3339),
		CourseParticipationIds: recipientIDs,
	})
	if err != nil {
		log.Error("failed to claim status mail recipients: ", err)
		return response, fmt.Errorf("failed to retrieve participant information for course phase %s with status %s: %v", coursePhaseID, status, err)
	}

	// 4.) Send to every claimed participant, releasing the claim where the send failed so the next
	// trigger picks the participant up again.
	for _, participant := range participants {
		placeholderMap := getStatusEmailPlaceholderValues(mailingInfo.CourseName, mailingInfo.CourseStartDate, mailingInfo.CourseEndDate, db.GetParticipantMailingInformationByIDsRow(participant))
		// replace values in subject
		finalSubject := replacePlaceholders(mailingInfo.MailSubject, placeholderMap)

		// replace values in content
		finalMessage := replacePlaceholders(mailingInfo.MailContent, placeholderMap)

		err = s.sendMail(courseMailingSettings, participant.Email.String, finalSubject, finalMessage)
		if err != nil {
			log.Error("failed to send status mail to participant: ", err)
			response.FailedEmails = append(response.FailedEmails, participant.Email.String)
			s.releaseStatusMailClaim(ctx, coursePhaseID, participant.CourseParticipationID, status)
			continue
		}

		log.Debug("Successfully sent status mail to: ", participant.Email.String)
		response.SuccessfulEmails = append(response.SuccessfulEmails, participant.Email.String)
	}

	return response, nil
}

func (s *MailingService) releaseStatusMailClaim(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, status db.PassStatus) {
	if err := s.queries.ReleaseStatusMailClaim(ctx, db.ReleaseStatusMailClaimParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		Status:                string(status),
	}); err != nil {
		log.Error("failed to release status mail claim for participant, it will not be retried: ", err)
	}
}

// SendMail sends a transactional mail with no Reply-To, CC, or BCC. For mails
// tied to a course phase, use SendCourseMail.
func (s *MailingService) SendMail(recipientAddress, subject, htmlBody string) error {
	if err := s.validateMailInputs(recipientAddress, subject, htmlBody); err != nil {
		return err
	}

	to := mail.Address{Address: recipientAddress}

	var message strings.Builder
	s.buildBaseMailHeader(&message, to.String(), subject)
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	return s.dispatchSMTP([]string{recipientAddress}, message.String())
}

func (s *MailingService) SendCourseMail(courseMailingSettings mailingDTO.CourseMailingSettings, recipientAddress, subject, htmlBody string) error {
	if err := s.validateMailInputs(recipientAddress, subject, htmlBody); err != nil {
		return err
	}

	to := mail.Address{Address: recipientAddress}

	var message strings.Builder
	s.buildBaseMailHeader(&message, to.String(), subject)
	fmt.Fprintf(&message, "Reply-To: %s\r\n", courseMailingSettings.ReplyTo.String())
	if len(courseMailingSettings.CC) > 0 {
		var ccString string
		for _, cc := range courseMailingSettings.CC {
			ccString += cc.String() + ","
		}
		fmt.Fprintf(&message, "CC: %s\r\n", ccString)
	}
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	rcpts := []string{recipientAddress}
	for _, cc := range courseMailingSettings.CC {
		rcpts = append(rcpts, cc.Address)
	}
	for _, bcc := range courseMailingSettings.BCC {
		rcpts = append(rcpts, bcc.Address)
	}

	return s.dispatchSMTP(rcpts, message.String())
}

func (s *MailingService) validateMailInputs(recipientAddress, subject, htmlBody string) error {
	if s == nil {
		return errors.New("mailing service is not initialized")
	}
	if s.senderEmail.Address == "" {
		return errors.New("mailing is not correctly configured: sender email address is empty")
	}
	if recipientAddress == "" {
		return errors.New("mailing is not correctly configured: recipient address is empty")
	}
	if subject == "" {
		return errors.New("mailing is not correctly configured: subject is empty")
	}
	if htmlBody == "" {
		return errors.New("mailing is not correctly configured: HTML body is empty")
	}
	return nil
}

func (s *MailingService) dispatchSMTP(recipients []string, message string) error {
	addr := net.JoinHostPort(s.smtpHost, s.smtpPort)
	log.Debug("Connecting to SMTP server: ", addr)

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Error("failed to connect to SMTP server: ", err.Error())
		return fmt.Errorf("failed to connect to SMTP server %s: %v", addr, err)
	}

	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		_ = conn.Close()
		log.Error("failed to set connection deadline: ", err)
		return fmt.Errorf("failed to set SMTP connection timeout: %v", err)
	}

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		_ = conn.Close()
		log.Error("failed to create SMTP client: ", err.Error())
		return fmt.Errorf("failed to create SMTP client for %s: %v", s.smtpHost, err)
	}
	defer func() { _ = client.Close() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: s.smtpHost,
			MinVersion: tls.VersionTLS12,
		}
		if err = client.StartTLS(config); err != nil {
			log.Error("failed to start TLS: ", err)
			return fmt.Errorf("failed to establish TLS connection with %s: %v", s.smtpHost, err)
		}
	}

	if s.smtpUsername != "" && s.smtpPassword != "" {
		auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
		if err := client.Auth(auth); err != nil {
			log.Error("failed to authenticate with SMTP server: ", err)
			return fmt.Errorf("SMTP authentication failed for user '%s' on server %s: %v", s.smtpUsername, s.smtpHost, err)
		}
	}

	if err := client.Mail(s.senderEmail.Address); err != nil {
		log.Error("failed to set sender: ", err)
		return fmt.Errorf("SMTP server rejected sender address '%s': %v", s.senderEmail.Address, err)
	}

	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			log.Error("failed to set recipient: ", err)
			return fmt.Errorf("SMTP server rejected recipient address '%s': %v", rcpt, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		log.Error("failed to send data: ", err)
		return fmt.Errorf("SMTP server failed to accept email data: %v", err)
	}
	if _, err = writer.Write([]byte(message)); err != nil {
		log.Error("failed to write message: ", err)
		return fmt.Errorf("failed to write email content to SMTP server: %v", err)
	}
	if err := writer.Close(); err != nil {
		log.Error("failed to close writer: ", err)
		return fmt.Errorf("failed to finalize email transmission: %v", err)
	}

	return client.Quit()
}

func (s *MailingService) getSenderInformation(ctx context.Context, coursePhaseID uuid.UUID) (mailingDTO.CourseMailingSettings, error) {
	courseMailing, err := s.queries.GetCourseMailingSettingsForCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to get course mailing settings: ", err)
		return mailingDTO.CourseMailingSettings{}, fmt.Errorf("failed to retrieve course mailing settings for course phase %s: %v", coursePhaseID, err)
	}

	if courseMailing.ReplyToEmail == "" || courseMailing.ReplyToName == "" {
		log.Error("reply to email or name is not set")
		return mailingDTO.CourseMailingSettings{}, fmt.Errorf("course mailing configuration incomplete: reply-to email ('%s') or name ('%s') is empty", courseMailing.ReplyToEmail, courseMailing.ReplyToName)
	}

	courseMailingSettings, err := mailingDTO.GetCourseMailingSettingsFromDBModel(courseMailing)
	if err != nil {
		log.Error("failed to get course mailing settings: ", err)
		return mailingDTO.CourseMailingSettings{}, fmt.Errorf("failed to parse course mailing settings from database: %v", err)
	}

	return courseMailingSettings, nil

}

// generateMessageID creates a unique Message-ID header value
func generateMessageID() string {
	// Create a unique identifier using timestamp and random bytes
	timestamp := time.Now().Unix()

	// Generate 8 random bytes for uniqueness
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to UUID if random bytes fail
		return fmt.Sprintf("<%d.%s@prompt2.local>", timestamp, uuid.New().String())
	}

	// Convert random bytes to hex string
	randomHex := fmt.Sprintf("%x", randomBytes)
	return fmt.Sprintf("<%d.%s@prompt2.local>", timestamp, randomHex)
}

// Does not write the trailing blank line; callers append optional headers first.
func (s *MailingService) buildBaseMailHeader(message *strings.Builder, recipient, subject string) {
	fmt.Fprintf(message, "From: %s\r\n", s.senderEmail.String())
	fmt.Fprintf(message, "To: %s\r\n", recipient)
	fmt.Fprintf(message, "Subject: %s\r\n", subject)
	fmt.Fprintf(message, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(message, "Message-ID: %s\r\n", generateMessageID())
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
}
