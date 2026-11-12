import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  ScoreLevel,
} from '@tumaet/prompt-shared-state'

import type { ActionItem } from '../../interfaces/actionItem'
import type { Assessment } from '../../interfaces/assessment'
import type {
  AggregatedEvaluationResult,
  StudentAssessmentResults,
} from '../../interfaces/assessmentResults'
import type { CategoryWithCompetencies } from '../../interfaces/category'
import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import type { StudentEvaluationResults } from '../../interfaces/evaluationResults'
import { getWeightedScoreLevel } from './getWeightedScoreLevel'

type ExampleConfig = Pick<
  CoursePhaseConfig,
  | 'gradingSheetVisible'
  | 'actionItemsVisible'
  | 'gradeSuggestionVisible'
  | 'selfEvaluationEnabled'
  | 'peerEvaluationEnabled'
>

export const EXAMPLE_PARTICIPATION_ID = 'example-participation'
const EXAMPLE_AUTHOR = 'Example Tutor'
const EXAMPLE_DATE = '2026-01-01T00:00:00Z'

const EXAMPLE_LEVELS = [
  ScoreLevel.Good,
  ScoreLevel.VeryGood,
  ScoreLevel.Ok,
  ScoreLevel.Good,
  ScoreLevel.Bad,
]

const EXAMPLE_ACTION_ITEMS = [
  'Break larger features into smaller, reviewable pull requests.',
  'Share your progress in the team meeting before you get stuck.',
]

const exampleLevel = (index: number) => EXAMPLE_LEVELS[index % EXAMPLE_LEVELS.length]

const competencyIDs = (categories: CategoryWithCompetencies[]) =>
  categories.flatMap((category) => category.competencies.map((competency) => competency.id))

const averagedResults = (ids: string[], offset: number): AggregatedEvaluationResult[] =>
  ids.map((competencyID, index) => ({
    competencyID,
    averageScoreNumeric: mapScoreLevelToNumber(exampleLevel(index + offset)),
  }))

export const buildExampleAssessmentResults = (
  coursePhaseID: string,
  categories: CategoryWithCompetencies[],
  config: ExampleConfig,
): StudentAssessmentResults => {
  const ids = config.gradingSheetVisible ? competencyIDs(categories) : []

  const assessments: Assessment[] = ids.map((competencyID, index) => ({
    id: `example-assessment-${competencyID}`,
    courseParticipationID: EXAMPLE_PARTICIPATION_ID,
    coursePhaseID,
    competencyID,
    scoreLevel: exampleLevel(index),
    assessedAt: EXAMPLE_DATE,
    author: EXAMPLE_AUTHOR,
    authorID: EXAMPLE_AUTHOR,
  }))

  const scoreNumeric = getWeightedScoreLevel(assessments, categories)

  const actionItems: ActionItem[] = config.actionItemsVisible
    ? EXAMPLE_ACTION_ITEMS.map((action, index) => ({
        id: `example-action-item-${index}`,
        coursePhaseID,
        courseParticipationID: EXAMPLE_PARTICIPATION_ID,
        action,
        createdAt: new Date(EXAMPLE_DATE),
        author: EXAMPLE_AUTHOR,
      }))
    : []

  return {
    courseParticipationID: EXAMPLE_PARTICIPATION_ID,
    coursePhaseID,
    assessmentCompletion: {
      courseParticipationID: EXAMPLE_PARTICIPATION_ID,
      coursePhaseID,
      completedAt: EXAMPLE_DATE,
      completed: true,
      author: EXAMPLE_AUTHOR,
      comment: config.gradeSuggestionVisible
        ? 'Reliable team member with strong technical contributions.'
        : '',
      gradeSuggestion: config.gradeSuggestionVisible ? 1.7 : 0,
    },
    assessments,
    categoryAssessments: [],
    studentScore:
      assessments.length > 0
        ? { scoreNumeric, scoreLevel: mapNumberToScoreLevel(scoreNumeric) }
        : undefined,
    selfEvaluationResults: config.selfEvaluationEnabled ? averagedResults(ids, 1) : [],
    peerEvaluationResults: config.peerEvaluationEnabled ? averagedResults(ids, 3) : [],
    actionItems,
  }
}

export const buildExampleEvaluationResults = (
  coursePhaseID: string,
  selfCategories: CategoryWithCompetencies[],
  peerCategories: CategoryWithCompetencies[],
): StudentEvaluationResults => ({
  courseParticipationID: EXAMPLE_PARTICIPATION_ID,
  coursePhaseID,
  selfResults: competencyIDs(selfCategories).map((competencyID, index) => ({
    competencyID,
    scoreLevel: exampleLevel(index),
  })),
  peerResults: averagedResults(competencyIDs(peerCategories), 1),
})
