import { devOpsChallengeAxiosInstance } from '../devOpsChallengeServerConfig'

export const getPassedStudentsCount = async (coursePhaseID: string): Promise<number> => {
  try {
    const response = await devOpsChallengeAxiosInstance.get(`/${coursePhaseID}/passed`)
    return response.data.passedStudents
  } catch (err) {
    console.error(err)
    throw err
  }
}
