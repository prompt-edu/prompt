import type { TeaseStudent } from '../interfaces/tease/student'

export const SURVEY_STATUS_COLUMN_ID = 'surveyStatus'
export const SURVEY_SUBMITTED = 'Submitted'
export const SURVEY_NOT_SUBMITTED = 'Not submitted'
export const SURVEY_STATUS_OPTIONS = [SURVEY_SUBMITTED, SURVEY_NOT_SUBMITTED]

export const hasSubmittedSurvey = (student: TeaseStudent | undefined): boolean =>
  (student?.projectPreferences?.length ?? 0) > 0

export const getSurveyStatus = (student: TeaseStudent | undefined): string =>
  hasSubmittedSurvey(student) ? SURVEY_SUBMITTED : SURVEY_NOT_SUBMITTED

export const NOT_SUBMITTED_FILTER_SEARCH = `?filters=${encodeURIComponent(
  JSON.stringify([{ id: SURVEY_STATUS_COLUMN_ID, value: [SURVEY_NOT_SUBMITTED] }]),
)}`
