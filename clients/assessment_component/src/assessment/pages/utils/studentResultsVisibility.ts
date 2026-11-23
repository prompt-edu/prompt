import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'

export type StudentResultsVisibility = Pick<
  CoursePhaseConfig,
  'gradeSuggestionVisible' | 'actionItemsVisible' | 'gradingSheetVisible'
>

// The sections of the assessment results that students see after release, in the order the
// release dialog lists them. A missing config counts as nothing visible.
export const getVisibleResultSections = (config?: StudentResultsVisibility): string[] => {
  const sections: string[] = []
  if (config?.gradeSuggestionVisible) sections.push('grade suggestions')
  if (config?.actionItemsVisible) sections.push('action items')
  if (config?.gradingSheetVisible) sections.push('grading sheet')
  return sections
}

export const hasVisibleResultSection = (config?: StudentResultsVisibility): boolean =>
  getVisibleResultSections(config).length > 0
