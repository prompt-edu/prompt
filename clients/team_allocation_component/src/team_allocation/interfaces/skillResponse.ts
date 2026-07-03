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

export const SKILL_LEVEL_ORDER: SkillLevel[] = [
  SkillLevel.VERY_BAD,
  SkillLevel.BAD,
  SkillLevel.OK,
  SkillLevel.GOOD,
  SkillLevel.VERY_GOOD,
]

export const SKILL_LEVEL_LABELS: Record<SkillLevel, string> = {
  [SkillLevel.VERY_BAD]: 'Very Bad',
  [SkillLevel.BAD]: 'Bad',
  [SkillLevel.OK]: 'Ok',
  [SkillLevel.GOOD]: 'Good',
  [SkillLevel.VERY_GOOD]: 'Very Good',
}
