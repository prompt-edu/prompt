import type { SkillResponse } from './skillResponse'
import type { TeamPreference } from './teamPreference'

export type SurveyResponse = {
  skillResponses: SkillResponse[]
  teamPreferences: TeamPreference[]
}
