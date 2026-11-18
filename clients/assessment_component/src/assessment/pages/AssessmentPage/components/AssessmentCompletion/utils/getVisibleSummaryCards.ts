import type { CoursePhaseConfig } from '../../../../../interfaces/coursePhaseConfig'

export interface VisibleSummaryCards {
  generalRemarks: boolean
  grade: boolean
  actionItems: boolean
}

type SummaryVisibilityConfig = Pick<
  CoursePhaseConfig,
  'gradeSuggestionVisible' | 'actionItemsVisible'
>

/**
 * Decides which summary cards to render. Tutors always see every card. The read-only student
 * view only shows the sections released to students: the server returns the general remarks
 * together with the grade suggestion, so both follow `gradeSuggestionVisible`.
 */
export const getVisibleSummaryCards = (
  readOnly: boolean,
  config: SummaryVisibilityConfig | undefined,
): VisibleSummaryCards => {
  if (!readOnly) {
    return { generalRemarks: true, grade: true, actionItems: true }
  }

  const gradeSuggestionVisible = config?.gradeSuggestionVisible ?? false
  return {
    generalRemarks: gradeSuggestionVisible,
    grade: gradeSuggestionVisible,
    actionItems: config?.actionItemsVisible ?? false,
  }
}
