import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import type {
  ResultsReleasedMailReport,
  ResultsReleasedMailSettings,
} from '../../../../interfaces/resultsReleasedMail'

export const parseResultsReleasedMail = (
  coursePhase: CoursePhaseWithMetaData | undefined,
): ResultsReleasedMailSettings => {
  const mailingSettings = coursePhase?.restrictedData?.mailingSettings as
    | Record<string, unknown>
    | undefined
  const mail = mailingSettings?.resultsReleasedMail as
    | Partial<ResultsReleasedMailSettings>
    | undefined

  return {
    subject: mail?.subject ?? '',
    content: mail?.content ?? '',
    sendOnRelease: mail?.sendOnRelease ?? false,
  }
}

export const isTemplateComplete = ({ subject, content }: ResultsReleasedMailSettings) =>
  subject.trim() !== '' && content.trim() !== ''

export const describeMailReport = ({
  successfulEmails,
  failedEmails,
}: ResultsReleasedMailReport): string => {
  if (successfulEmails.length === 0 && failedEmails.length === 0) {
    return 'All students were already notified.'
  }

  const sent = `${successfulEmails.length} notification ${
    successfulEmails.length === 1 ? 'mail was' : 'mails were'
  } sent.`
  return failedEmails.length > 0 ? `${sent} ${failedEmails.length} could not be delivered.` : sent
}
