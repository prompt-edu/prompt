import { SkillResponse } from './skillResponse'
import { TeamPreference } from './teamPreference'

export type SurveyResponse = {
  skillResponses: SkillResponse[]
  teamPreferences?: TeamPreference[]
  fieldPreferences?: TeamPreference[]
  preferenceMode?: 'teams' | 'fields'
}
