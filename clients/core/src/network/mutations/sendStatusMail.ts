import { axiosInstance, type MailingReport, type SendStatusMail } from '@tumaet/prompt-shared-state'

export interface SendStatusMailRequest extends SendStatusMail {
  recipientCourseParticipationIDs?: string[]
}

export const sendStatusMail = async (
  coursePhaseID: string,
  status: SendStatusMailRequest,
): Promise<MailingReport> => {
  try {
    return (
      await axiosInstance.put(`/api/mailing/${coursePhaseID}`, status, {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      })
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
