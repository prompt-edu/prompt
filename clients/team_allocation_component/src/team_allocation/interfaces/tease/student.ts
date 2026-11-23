import type { Comment } from './comment'
import type { Device } from './device'
import type { Gender } from './gender'
import type { Language } from './language'
import type { ProjectPreference } from './projectPreference'
import type { SkillProficiency, UNKNOWN_SKILL_PROFICIENCY } from './skillProficiency'
import type { StudentSkill } from './studentSkill'

type IntroProficiency = SkillProficiency | typeof UNKNOWN_SKILL_PROFICIENCY

export interface TeaseStudent {
  devices: Array<Device>
  email: string
  firstName: string
  gender: Gender
  id: string
  introCourseProficiency: IntroProficiency
  introSelfEvaluation: IntroProficiency
  languages: Array<Language>
  lastName: string
  nationality: string
  projectPreferences: Array<ProjectPreference>
  semester: number
  skills: Array<StudentSkill>
  studentComments: Array<Comment>
  studyDegree: string
  studyProgram: string
  tutorComments: Array<Comment>
}
