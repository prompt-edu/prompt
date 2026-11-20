export enum SkillProficiency {
  Novice = 'Novice',
  Intermediate = 'Intermediate',
  Advanced = 'Advanced',
  Expert = 'Expert',
}

// The server sends this instead of a proficiency when the previous phases have no score level.
export const UNKNOWN_SKILL_PROFICIENCY = 'Unknown'
