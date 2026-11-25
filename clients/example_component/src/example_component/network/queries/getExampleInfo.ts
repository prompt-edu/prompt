import { exampleAxiosInstance } from '../exampleServerConfig'

export const getExampleInfo = async (coursePhaseID: string): Promise<string> => {
  return (await exampleAxiosInstance.get(`/example-service/api/course_phase/${coursePhaseID}/info`))
    .data
}
