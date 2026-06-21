package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing"
	log "github.com/sirupsen/logrus"
)

func sendDeletionConfirmationMail(ctx context.Context, requestID uuid.UUID, recipientEmail string, status db.PrivacyDeletionRequestStatus) {
	if recipientEmail == "" {
		log.WithField("requestID", requestID).Info("skipping deletion confirmation mail: no recipient email on record")
		return
	}

	var subject, body string
	switch status {
	case db.PrivacyDeletionRequestStatusSucceeded:
		subject = deletionSuccessSubject()
		body = deletionSuccessBody()
	case db.PrivacyDeletionRequestStatusFailed:
		subject = deletionFailureSubject()
		body = deletionFailureBody()
	default:
		log.WithFields(log.Fields{"requestID": requestID, "status": status}).
			Warn("skipping deletion confirmation mail: unexpected terminal status")
		return
	}

	if err := mailing.SendMail(recipientEmail, subject, body); err != nil {
		log.WithError(err).WithField("requestID", requestID).
			Error("failed to send deletion confirmation mail")
	}

	if err := PrivacyServiceSingleton.queries.ClearDeletionRequestRecipientEmail(ctx, requestID); err != nil {
		log.WithError(err).WithField("requestID", requestID).
			Warn("failed to clear recipient email after deletion confirmation")
	}
}

func deletionSuccessSubject() string {
	return "Your PROMPT data deletion request has been completed"
}

func deletionSuccessBody() string {
	return `<p>Hello,</p>
<p>Your request to delete your personal data from PROMPT has been processed successfully.</p>
<p>All personal data associated with your account has been removed from our records. You can no longer sign in to the platform.</p>
<p>Kind regards,<br>The PROMPT Team</p>`
}

func deletionFailureSubject() string {
	return "Your PROMPT data deletion request could not be completed"
}

func deletionFailureBody() string {
	return `<p>Hello,</p>
<p>We attempted to process your PROMPT data deletion request, but one or more steps did not complete successfully.</p>
<p>You do not need to take any further action right now. If you have questions, please contact your course administration.</p>
<p>Kind regards,<br>The PROMPT Team</p>`
}
