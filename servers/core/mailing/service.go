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
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	log "github.com/sirupsen/logrus"
)

type MailingService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	senderEmail  mail.Address
	clientURL    string
	queries      db.Queries
	conn         *pgxpool.Pool
}

var MailingServiceSingleton *MailingService

func SendApplicationConfirmationMail(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) (bool, error) {
	isApplicationPhase, err := MailingServiceSingleton.queries.CheckIfCoursePhaseIsApplicationPhase(ctx, coursePhaseID)
	if err != nil {
		return false, fmt.Errorf("failed to verify if course phase %s is an application phase: %v", coursePhaseID, err)
	}
	if !isApplicationPhase {
		return false, fmt.Errorf("course phase %s is not an application phase, cannot send confirmation mail", coursePhaseID)
	}

	mailingInfo, err := MailingServiceSingleton.queries.GetConfirmationMailingInformation(ctx, db.GetConfirmationMailingInformationParams{
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

	courseMailingSettings, err := getSenderInformation(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to get sender information")
		return false, fmt.Errorf("failed to get sender information for course phase %s: %v", coursePhaseID, err)
	}

	log.Info("Sending confirmation mail to ", mailingInfo.Email.String)

	applicationURL := fmt.Sprintf("%s/apply/%s", MailingServiceSingleton.clientURL, coursePhaseID.String())
	placeholderValues := getApplicationConfirmationPlaceholderValues(mailingInfo, applicationURL)
	finalMessage := replacePlaceholders(mailingInfo.ConfirmationMailContent, placeholderValues)

	// replace values in subject
	finalSubject := replacePlaceholders(mailingInfo.ConfirmationMailSubject, placeholderValues)

	err = SendMail(courseMailingSettings, mailingInfo.Email.String, finalSubject, finalMessage)
	if err != nil {
		log.Error("failed to send confirmation mail: ", err)
		return false, fmt.Errorf("failed to send confirmation mail to %s: %v", mailingInfo.Email.String, err)
	}

	return true, nil
}

func SendStatusMailManualTrigger(ctx context.Context, coursePhaseID uuid.UUID, status db.PassStatus) (mailingDTO.MailingReport, error) {
	response := mailingDTO.MailingReport{}
	mailingInfo := mailingDTO.MailingInfo{}

	// 1.) get mailing info for course phase
	switch status {
	case db.PassStatusPassed:
		infos, err := MailingServiceSingleton.queries.GetPassedMailingInformation(ctx, coursePhaseID)
		if err != nil {
			log.Error("failed to get mailing information: ", err)
			return mailingDTO.MailingReport{}, fmt.Errorf("failed to retrieve passed status mailing information for course phase %s: %v", coursePhaseID, err)
		}
		mailingInfo = mailingDTO.GetMailingInfoFromPassedMailingInformation(infos)

	case db.PassStatusFailed:
		infos, err := MailingServiceSingleton.queries.GetFailedMailingInformation(ctx, coursePhaseID)
		if err != nil {
			log.Error("failed to get mailing information: ", err)
			return mailingDTO.MailingReport{}, fmt.Errorf("failed to retrieve failed status mailing information for course phase %s: %v", coursePhaseID, err)
		}
		mailingInfo = mailingDTO.GetMailingInfoFromFailedMailingInformation(infos)

	default:
		log.Error("invalid status")
		return mailingDTO.MailingReport{}, fmt.Errorf("invalid pass status '%s': expected 'passed' or 'failed'", status)

	}

	// Get the course mailing settings
	courseMailingSettings, err := getSenderInformation(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to get sender information")
		return mailingDTO.MailingReport{}, fmt.Errorf("failed to get course mailing settings: %v", err)
	}

	// 2.) Check if mailing is configured -> return if not
	if mailingInfo.MailSubject == "" || mailingInfo.MailContent == "" {
		log.Error("mailing template is not correctly configured")
		return mailingDTO.MailingReport{}, fmt.Errorf("mailing template incomplete: subject ('%s') or content ('%s') is empty", mailingInfo.MailSubject, mailingInfo.MailContent)
	}

	// 3.) Get all participants that have not been accepted incl. information
	participants, err := MailingServiceSingleton.queries.GetParticipantMailingInformation(ctx, db.GetParticipantMailingInformationParams{
		ID:         coursePhaseID,
		PassStatus: db.NullPassStatus{PassStatus: status, Valid: true},
	})
	if err != nil {
		log.Error("failed to get participant mailing information: ", err)
		return mailingDTO.MailingReport{}, fmt.Errorf("failed to retrieve participant information for course phase %s with status %s: %v", coursePhaseID, status, err)
	}

	// 4.) Send mail to all participants
	for _, participant := range participants {
		placeholderMap := getStatusEmailPlaceholderValues(mailingInfo.CourseName, mailingInfo.CourseStartDate, mailingInfo.CourseEndDate, participant)
		// replace values in subject
		finalSubject := replacePlaceholders(mailingInfo.MailSubject, placeholderMap)

		// replace values in content
		finalMessage := replacePlaceholders(mailingInfo.MailContent, placeholderMap)

		err = SendMail(courseMailingSettings, participant.Email.String, finalSubject, finalMessage)
		if err != nil {
			log.Error("failed to send status mail to participant: ", err)
			response.FailedEmails = append(response.FailedEmails, participant.Email.String)
		} else {
			log.Debug("Successfully sent status mail to: ", participant.Email.String)
			response.SuccessfulEmails = append(response.SuccessfulEmails, participant.Email.String)
		}
	}

	return response, nil
}

// SendMail sends an email with the specified HTML body, recipient, and subject.
func SendMail(courseMailingSettings mailingDTO.CourseMailingSettings, recipientAddress, subject, htmlBody string) error {
	log.Debug("Starting mail validation")
	log.Debug("Sender email address: ", MailingServiceSingleton.senderEmail.Address)
	log.Debug("Sender username: ", MailingServiceSingleton.smtpUsername)
	log.Debug("Recipient address: ", recipientAddress)
	log.Debug("Subject: ", subject)
	log.Debug("HTML body length: ", len(htmlBody))

	if MailingServiceSingleton.senderEmail.Address == "" {
		log.Debug("Validation failed: sender email address is empty")
		return errors.New("mailing is not correctly configured: sender email address is empty")
	}
	if recipientAddress == "" {
		log.Debug("Validation failed: recipient address is empty")
		return errors.New("mailing is not correctly configured: recipient address is empty")
	}
	if subject == "" {
		log.Debug("Validation failed: subject is empty")
		return errors.New("mailing is not correctly configured: subject is empty")
	}
	if htmlBody == "" {
		log.Debug("Validation failed: HTML body is empty")
		return errors.New("mailing is not correctly configured: HTML body is empty")
	}

	log.Debug("Mail validation passed successfully")

	to := mail.Address{Address: recipientAddress}

	var message strings.Builder
	buildMailHeader(&message, courseMailingSettings, to.String(), subject)
	message.WriteString(htmlBody)

	// Send the email
	addr := net.JoinHostPort(MailingServiceSingleton.smtpHost, MailingServiceSingleton.smtpPort)
	log.Debug("Connecting to SMTP server: ", addr)

	// Create connection with timeout (30 seconds)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Error("failed to connect to SMTP server: ", err.Error())
		return fmt.Errorf("failed to connect to SMTP server %s: %v", addr, err)
	}

	// Set deadline for the entire SMTP operation
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		_ = conn.Close()
		log.Error("failed to set connection deadline: ", err)
		return fmt.Errorf("failed to set SMTP connection timeout: %v", err)
	}

	client, err := smtp.NewClient(conn, MailingServiceSingleton.smtpHost)
	if err != nil {
		_ = conn.Close()
		log.Error("failed to create SMTP client: ", err.Error())
		return fmt.Errorf("failed to create SMTP client for %s: %v", MailingServiceSingleton.smtpHost, err)
	}
	defer func() { _ = client.Close() }()

	// Enable STARTTLS if the server supports it (required for port 587)
	if ok, _ := client.Extension("STARTTLS"); ok {
		log.Debug("STARTTLS is supported, enabling TLS")
		config := &tls.Config{
			ServerName: MailingServiceSingleton.smtpHost,
			MinVersion: tls.VersionTLS12,
		}
		if err = client.StartTLS(config); err != nil {
			log.Error("failed to start TLS: ", err)
			return fmt.Errorf("failed to establish TLS connection with %s: %v", MailingServiceSingleton.smtpHost, err)
		}
		log.Debug("TLS connection established")
	} else {
		log.Debug("STARTTLS not supported by server")
	}

	// Use SMTP authentication if username and password are provided
	if MailingServiceSingleton.smtpUsername != "" && MailingServiceSingleton.smtpPassword != "" {
		log.Debug("Authenticating with SMTP server")
		auth := smtp.PlainAuth("", MailingServiceSingleton.smtpUsername, MailingServiceSingleton.smtpPassword, MailingServiceSingleton.smtpHost)
		if err := client.Auth(auth); err != nil {
			log.Error("failed to authenticate with SMTP server: ", err)
			return fmt.Errorf("SMTP authentication failed for user '%s' on server %s: %v", MailingServiceSingleton.smtpUsername, MailingServiceSingleton.smtpHost, err)
		}
		log.Debug("SMTP authentication successful")
	} else {
		log.Debug("No SMTP authentication configured")
	}

	// Set the sender and recipient
	log.Debug("Setting sender and recipients")
	if err := client.Mail(MailingServiceSingleton.senderEmail.Address); err != nil {
		log.Error("failed to set sender: ", err)
		return fmt.Errorf("SMTP server rejected sender address '%s': %v", MailingServiceSingleton.senderEmail.Address, err)
	}

	if err := client.Rcpt(recipientAddress); err != nil {
		log.Error("failed to set recipient: ", err)
		return fmt.Errorf("SMTP server rejected recipient address '%s': %v", recipientAddress, err)
	}

	// set all cc mails
	for _, cc := range courseMailingSettings.CC {
		log.Debug("Adding CC recipient: ", cc.Address)
		if err := client.Rcpt(cc.Address); err != nil {
			log.Error("failed to set cc: ", err)
			return fmt.Errorf("SMTP server rejected CC address '%s': %v", cc.Address, err)
		}
	}

	// set all bcc mails
	for _, bcc := range courseMailingSettings.BCC {
		log.Debug("Adding BCC recipient: ", bcc.Address)
		if err := client.Rcpt(bcc.Address); err != nil {
			log.Error("failed to set bcc: ", err)
			return fmt.Errorf("SMTP server rejected BCC address '%s': %v", bcc.Address, err)
		}
	}

	// Send the data
	log.Debug("Sending email data")
	writer, err := client.Data()
	if err != nil {
		log.Error("failed to send data: ", err)
		return fmt.Errorf("SMTP server failed to accept email data: %v", err)
	}
	_, err = writer.Write([]byte(message.String()))
	if err != nil {
		log.Error("failed to write message: ", err)
		return fmt.Errorf("failed to write email content to SMTP server: %v", err)
	}

	if err := writer.Close(); err != nil {
		log.Error("failed to close writer: ", err)
		return fmt.Errorf("failed to finalize email transmission: %v", err)
	}

	log.Debug("Email sent successfully")
	return client.Quit()
}

func getSenderInformation(ctx context.Context, coursePhaseID uuid.UUID) (mailingDTO.CourseMailingSettings, error) {
	courseMailing, err := MailingServiceSingleton.queries.GetCourseMailingSettingsForCoursePhaseID(ctx, coursePhaseID)
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

func buildMailHeader(message *strings.Builder, courseMailingSettings mailingDTO.CourseMailingSettings, recipient, subject string) {
	// using this instead of map to get a nicely formatted Mailing Header
	fmt.Fprintf(message, "From: %s\r\n", MailingServiceSingleton.senderEmail.String())
	fmt.Fprintf(message, "To: %s\r\n", recipient)
	fmt.Fprintf(message, "Reply-To: %s\r\n", courseMailingSettings.ReplyTo.String())
	fmt.Fprintf(message, "Subject: %s\r\n", subject)

	// Add Date header in RFC 2822 format
	fmt.Fprintf(message, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))

	// Add unique Message-ID header
	fmt.Fprintf(message, "Message-ID: %s\r\n", generateMessageID())

	message.WriteString("MIME-Version: 1.0\r\n") // Improve Spam Score by setting explicit MIME-Version
	message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")

	if len(courseMailingSettings.CC) > 0 {
		var ccString string
		for _, cc := range courseMailingSettings.CC {
			ccString += cc.String() + ","
		}
		fmt.Fprintf(message, "CC: %s\r\n", ccString)
	}

	// BCC are set in the client and not the header
	message.WriteString("\r\n")
}
