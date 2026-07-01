import { StudentNameUpdateRequest } from '../../interfaces/studentNameUpdateRequest'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const addStudentNamesToTeams = async (
  studentNames: StudentNameUpdateRequest,
): Promise<void> => {
  try {
    await teamAllocationAxiosInstance.post(
      `/team-allocation/api/course_phase/${studentNames.coursePhaseID}/team/student-names`,
      studentNames,
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
