import { getStudyDegreeString, mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'

import type { ActionItem } from '../../../interfaces/actionItem'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { useStudentAssessmentStore } from '../../../zustand/useStudentAssessmentStore'
import { getScoreLevelDescription } from '../../utils/getScoreLevelDescription'
import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'
import { FeedbackSection } from './FeedbackSection'
import { ScoreChip } from './ScoreChip'

interface AssessmentPrintReportProps {
  categories: CategoryWithCompetencies[]
  feedbackItems?: FeedbackItem[]
  actionItems?: ActionItem[]
}

export const AssessmentPrintReport = ({
  categories,
  feedbackItems = [],
  actionItems = [],
}: AssessmentPrintReportProps) => {
  const {
    assessmentParticipation,
    assessments,
    categoryAssessments,
    assessmentCompletion,
    studentScore,
  } = useStudentAssessmentStore()

  const student = assessmentParticipation?.student
  if (!student) return null

  const sortedCategories = [...categories].sort((a, b) => a.name.localeCompare(b.name))
  const degree = student.studyDegree ? getStudyDegreeString(student.studyDegree) : 'N/A'

  return (
    <div className='print-report hidden text-black print:block'>
      <header className='mb-6 break-inside-avoid border-b border-gray-300 pb-4'>
        <h1 className='text-2xl font-bold'>
          {student.firstName} {student.lastName}
        </h1>
        <p className='mt-1 text-sm text-gray-700'>
          {student.studyProgram || 'N/A'} · {degree} · Semester {student.currentSemester || 'N/A'}
        </p>
        <div className='mt-3 flex flex-wrap items-center gap-x-6 gap-y-1 text-sm'>
          <span>
            <strong>Status:</strong> {assessmentCompletion?.completed ? 'Finalized' : 'In progress'}
          </span>
          {studentScore && (
            <span className='flex items-center gap-2'>
              <strong>Overall score:</strong>
              <ScoreChip scoreLevel={studentScore.scoreLevel} />
              <span className='text-gray-600'>({studentScore.scoreNumeric.toFixed(1)})</span>
            </span>
          )}
          {assessmentCompletion?.gradeSuggestion ? (
            <span>
              <strong>Grade suggestion:</strong> {assessmentCompletion.gradeSuggestion.toFixed(1)}
            </span>
          ) : null}
        </div>
      </header>

      {sortedCategories.map((category) => {
        const categoryScores = assessments.filter((assessment) =>
          category.competencies.some((competency) => competency.id === assessment.competencyID),
        )
        const categoryScore = getWeightedScoreLevel(categoryScores, [category])
        const comment = categoryAssessments.find((ca) => ca.categoryID === category.id)?.comment
        const sortedCompetencies = [...category.competencies].sort((a, b) =>
          a.name.localeCompare(b.name),
        )

        return (
          <section key={category.id} className='mb-6'>
            <div className='mb-2 flex items-center justify-between gap-2 border-b border-gray-200 pb-1'>
              <h2 className='text-lg font-semibold'>{category.name}</h2>
              {categoryScores.length > 0 && (
                <ScoreChip scoreLevel={mapNumberToScoreLevel(categoryScore)} />
              )}
            </div>

            {comment && (
              <p className='mb-3 whitespace-pre-wrap text-sm italic text-gray-700'>{comment}</p>
            )}

            <div className='space-y-3'>
              {sortedCompetencies.map((competency) => {
                const assessment = categoryScores.find(
                  (item) => item.competencyID === competency.id,
                )
                return (
                  <div
                    key={competency.id}
                    className='break-inside-avoid rounded-sm border border-gray-200 p-3'
                  >
                    <div className='flex items-center justify-between gap-2'>
                      <h3 className='text-sm font-medium'>{competency.name}</h3>
                      {assessment ? (
                        <ScoreChip scoreLevel={assessment.scoreLevel} />
                      ) : (
                        <span className='text-xs text-gray-400'>Not assessed</span>
                      )}
                    </div>
                    {assessment && (
                      <p className='mt-1 text-sm text-gray-700'>
                        {getScoreLevelDescription(assessment.scoreLevel, competency)}
                      </p>
                    )}
                  </div>
                )
              })}
            </div>
          </section>
        )
      })}

      <FeedbackSection feedbackItems={feedbackItems} />

      {(assessmentCompletion?.comment ||
        actionItems.length > 0 ||
        assessmentCompletion?.gradeSuggestion) && (
        <section className='break-inside-avoid'>
          <h2 className='mb-2 border-b border-gray-200 pb-1 text-lg font-semibold'>Summary</h2>
          {assessmentCompletion?.comment && (
            <div className='mb-3'>
              <h3 className='text-sm font-medium'>General remarks</h3>
              <p className='whitespace-pre-wrap text-sm text-gray-700'>
                {assessmentCompletion.comment}
              </p>
            </div>
          )}
          {actionItems.length > 0 && (
            <div className='mb-3'>
              <h3 className='text-sm font-medium'>Action items</h3>
              <ul className='ml-5 list-disc text-sm text-gray-700'>
                {actionItems.map((item) => (
                  <li key={item.id}>{item.action}</li>
                ))}
              </ul>
            </div>
          )}
          {assessmentCompletion?.gradeSuggestion ? (
            <p className='text-sm'>
              <strong>Grade suggestion:</strong> {assessmentCompletion.gradeSuggestion.toFixed(1)}
            </p>
          ) : null}
          {assessmentCompletion?.completed && assessmentCompletion.completedAt && (
            <p className='mt-2 text-xs text-gray-500'>
              Finalized on {new Date(assessmentCompletion.completedAt).toLocaleDateString()}
            </p>
          )}
        </section>
      )}
    </div>
  )
}
