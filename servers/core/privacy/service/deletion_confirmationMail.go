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
		subject = deletionSuccessSubject
		body = deletionSuccessBody
	case db.PrivacyDeletionRequestStatusFailed:
		subject = deletionFailureSubject
		body = deletionFailureBody
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

const (
	deletionSuccessSubject = "Your PROMPT account and data have been deleted"
	deletionSuccessBody    = `<p>Hello,</p>
<p>Your PROMPT account and the personal data associated with it have been deleted.</p>
<p>You can no longer sign in to the platform. If you have any questions about this deletion, please contact the PROMPT team.</p>
<p>Kind regards,<br>The PROMPT Team</p>`

	deletionFailureSubject = "Deletion of your PROMPT data could not be completed"
	deletionFailureBody    = `<p>Hello,</p>
<p>We attempted to delete your PROMPT data, but one or more steps did not complete successfully.</p>
<p>Some of your data may still be present on the platform until the issue is resolved. If you have any questions, please contact the PROMPT team.</p>
<p>Kind regards,<br>The PROMPT Team</p>`
)
