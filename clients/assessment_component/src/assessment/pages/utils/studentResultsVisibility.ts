import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'

export type StudentResultsVisibility = Pick<
  CoursePhaseConfig,
  'gradeSuggestionVisible' | 'actionItemsVisible' | 'gradingSheetVisible'
>

// The sections of the assessment results that students see after release, in the order the
// release dialog lists them. The server returns the general remarks together with the grade
// suggestion, so both follow `gradeSuggestionVisible`. A missing config counts as nothing visible.
export const getVisibleResultSections = (config?: StudentResultsVisibility): string[] => {
  const sections: string[] = []
  if (config?.gradeSuggestionVisible) sections.push('grade suggestions', 'general remarks')
  if (config?.actionItemsVisible) sections.push('action items')
  if (config?.gradingSheetVisible) sections.push('grading sheet')
  return sections
}

export const hasVisibleResultSection = (config?: StudentResultsVisibility): boolean =>
  getVisibleResultSections(config).length > 0

// Joins the visible sections for a sentence, e.g. "grade suggestions, general remarks, and
// action items".
export const describeVisibleResultSections = (config?: StudentResultsVisibility): string => {
  const sections = getVisibleResultSections(config)
  if (sections.length <= 2) {
    return sections.join(' and ')
  }
  return `${sections.slice(0, -1).join(', ')}, and ${sections[sections.length - 1]}`
}
