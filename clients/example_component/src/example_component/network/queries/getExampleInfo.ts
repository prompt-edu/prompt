import { exampleAxiosInstance } from '../exampleServerConfig'

export const getExampleInfo = async (coursePhaseID: string): Promise<string> => {
  try {
    return (
      await exampleAxiosInstance.get(`/example-service/api/course_phase/${coursePhaseID}/info`)
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
