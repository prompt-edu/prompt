import type { ApplicationMailingMetaData } from '../../../interfaces/applicationMailingMetaData'

export const parseApplicationMailingMetaData = (
  restrictedData: any,
): ApplicationMailingMetaData => {
  const {
    mailingSettings: {
      confirmationMailSubject = '',
      confirmationMailContent = '',
      failedMailSubject = '',
      failedMailContent = '',
      passedMailSubject = '',
      passedMailContent = '',
      sendConfirmationMail = false,
      sendRejectionMail = false,
      sendAcceptanceMail = false,
    } = {},
  } = restrictedData || {}

  return {
    confirmationMailSubject,
    confirmationMailContent,
    failedMailSubject,
    failedMailContent,
    passedMailSubject,
    passedMailContent,
    sendConfirmationMail,
    sendRejectionMail,
    sendAcceptanceMail,
  }
}
