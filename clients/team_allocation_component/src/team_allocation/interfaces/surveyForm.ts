import type { Team } from '@tumaet/prompt-shared-state'

import type { Skill } from './skill'

export type SurveyForm = {
  teams: Team[]
  skills: Skill[]
  deadline: Date
}
