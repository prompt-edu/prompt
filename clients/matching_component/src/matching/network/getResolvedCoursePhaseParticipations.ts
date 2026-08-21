import {
  type CoursePhaseParticipationWithStudent,
  getCoursePhaseParticipations,
} from '@tumaet/prompt-shared-state'
import { resolveParticipations } from './resolveParticipations'

/**
 * Fetch the matching phase participants from core and resolve any REST-provided
 * predecessor DTOs (e.g. the interview score) into each participation's `prevData`.
 */
export const getResolvedCoursePhaseParticipations = async (
  coursePhaseID: string,
): Promise<CoursePhaseParticipationWithStudent[]> => {
  const { participations, resolutions } = await getCoursePhaseParticipations(coursePhaseID)
  return resolveParticipations(participations, resolutions)
}
