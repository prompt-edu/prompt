import { DeveloperWithInfo } from '../../interfaces/DeveloperWithInfo'
import { devOpsChallengeAxiosInstance } from '../devOpsChallengeServerConfig'

interface StudentResponse {
  CourseParticipationId: string
  GithubUsername: string
  Attempts: number
  Passed: boolean
  PassedAt: string
  PassingPosition?: number
}

export const getAllDeveloperProfiles = async (
  coursePhaseID: string,
): Promise<DeveloperWithInfo[]> => {
  try {
    const students = (await devOpsChallengeAxiosInstance.get(`/${coursePhaseID}/students`)).data
      .students

    // custom mapping as the response uses Uppercase keys, clashing with PROMPTs naming schema
    return students.map((student: StudentResponse) => ({
      courseParticipationID: student.CourseParticipationId,
      attempts: student.Attempts,
      hasPassed: student.Passed,
      passedAt: student.PassedAt ? new Date(student.PassedAt) : null,
      passingPosition: student.PassingPosition, // or another appropriate default
    }))
  } catch (err) {
    console.error(err)
    throw err
  }
}
