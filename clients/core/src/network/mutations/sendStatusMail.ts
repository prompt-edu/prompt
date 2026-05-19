import { axiosInstance } from '@tumaet/prompt-shared-state'
import { SendStatusMail, MailingReport } from '@tumaet/prompt-shared-state'

export const sendStatusMail = async (
  coursePhaseID: string,
  status: SendStatusMail,
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
