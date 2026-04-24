import { Team } from '@tumaet/prompt-shared-state'

import { Skill } from './skill'

export type SurveyForm = {
  teams: Team[]
  skills: Skill[]
  deadline: Date
  preferenceMode?: 'teams' | 'fields'
}
