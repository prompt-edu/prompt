import { getStudyDegreeString, type Student } from '@tumaet/prompt-shared-state'
import { getStudentName } from '@tumaet/prompt-ui-components'

import type { ActionItem } from '../../../interfaces/actionItem'
import type { Assessment } from '../../../interfaces/assessment'
import type { AssessmentCompletion } from '../../../interfaces/assessmentCompletion'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { CategoryAssessment } from '../../../interfaces/categoryAssessment'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import type { StudentScore } from '../../../interfaces/studentScore'
import { useStudentAssessmentStore } from '../../../zustand/useStudentAssessmentStore'
import { PrintReport } from '../PrintReport/PrintReport'
import { ScoreChip } from '../PrintReport/ScoreChip'

interface AssessmentPrintReportProps {
  categories: CategoryWithCompetencies[]
  feedbackItems?: FeedbackItem[]
  actionItems?: ActionItem[]
}

interface AssessmentReportProps extends AssessmentPrintReportProps {
  student: Student
  assessments: Assessment[]
  categoryAssessments: CategoryAssessment[]
  assessmentCompletion?: AssessmentCompletion
  studentScore?: StudentScore
  className?: string
}

export const AssessmentPrintReport = (props: AssessmentPrintReportProps) => {
  const {
    assessmentParticipation,
    assessments,
    categoryAssessments,
    assessmentCompletion,
    studentScore,
  } = useStudentAssessmentStore()

  const student = assessmentParticipation?.student
  if (!student) return null

  return (
    <AssessmentReport
      {...props}
      student={student}
      assessments={assessments}
      categoryAssessments={categoryAssessments}
      assessmentCompletion={assessmentCompletion}
      studentScore={studentScore}
    />
  )
}

export const AssessmentReport = ({
  categories,
  feedbackItems = [],
  actionItems = [],
  student,
  assessments,
  categoryAssessments,
  assessmentCompletion,
  studentScore,
  className,
}: AssessmentReportProps) => {
  const degree = student.studyDegree ? getStudyDegreeString(student.studyDegree) : 'N/A'
  const categoryComments = Object.fromEntries(
    categoryAssessments
      .filter((categoryAssessment) => categoryAssessment.comment)
      .map((categoryAssessment) => [categoryAssessment.categoryID, categoryAssessment.comment]),
  )

  return (
    <PrintReport
      className={className}
      title={getStudentName(student)}
      subtitle={`${student.studyProgram || 'N/A'} · ${degree} · Semester ${student.currentSemester || 'N/A'}`}
      meta={
        <>
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
        </>
      }
      categories={categories}
      scores={assessments}
      categoryComments={categoryComments}
      feedbackItems={feedbackItems}
    >
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
    </PrintReport>
  )
}
