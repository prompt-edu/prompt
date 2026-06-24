export type SkillResponse = {
  skillID: string
  skillLevel: SkillLevel
}

export enum SkillLevel {
  VERY_BAD = 'very_bad',
  BAD = 'bad',
  OK = 'ok',
  GOOD = 'good',
  VERY_GOOD = 'very_good',
}
