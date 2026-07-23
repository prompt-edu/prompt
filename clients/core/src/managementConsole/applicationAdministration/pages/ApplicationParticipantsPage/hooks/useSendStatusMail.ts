import { sendStatusMail } from '@core/network/mutations/sendStatusMail'
import { type UseMutationResult, useMutation } from '@tanstack/react-query'
import type { MailingReport, PassStatus } from '@tumaet/prompt-shared-state'
import { useToast } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

interface SendStatusMailVariables {
  status: PassStatus
  recipientCourseParticipationIDs: string[]
}

export const useSendStatusMail = (): UseMutationResult<
  MailingReport,
  Error,
  SendStatusMailVariables
> => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { toast } = useToast()

  return useMutation({
    mutationFn: ({ status, recipientCourseParticipationIDs }: SendStatusMailVariables) =>
      sendStatusMail(phaseId ?? 'undefined', {
        statusMailToBeSend: status,
        recipientCourseParticipationIDs,
      }),
    onSuccess: (report: MailingReport) => {
      const sent = report.successfulEmails.length
      const failed = report.failedEmails.length
      toast({
        title: `Sent ${sent} mail${sent === 1 ? '' : 's'}.`,
        description: failed > 0 ? `${failed} could not be delivered.` : undefined,
        variant: failed > 0 ? 'destructive' : undefined,
      })
    },
    onError: () => {
      toast({
        title: 'Error',
        description: 'Failed to send the mails.',
        variant: 'destructive',
      })
    },
  })
}
