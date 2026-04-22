import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const unreleaseResults = async (coursePhaseID: string): Promise<void> => {
  try {
    await assessmentAxiosInstance.post(
      `assessment/api/course_phase/${coursePhaseID}/config/unrelease`,
      {},
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
