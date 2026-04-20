import { DeveloperProfile } from '../../interfaces/DeveloperProfile'
import { devOpsChallengeAxiosInstance } from '../devOpsChallengeServerConfig'

export const getDeveloperProfile = async (
  coursePhaseID: string,
): Promise<DeveloperProfile | undefined> => {
  try {
    return (await devOpsChallengeAxiosInstance.get(`/${coursePhaseID}/info`)).data
  } catch (err: any) {
    console.error(err)
    if (err?.response?.status === 404 && err?.response?.data?.error === 'student not found') {
      return undefined
    } else {
      throw err
    }
  }
}
