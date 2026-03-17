import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'

import type { StudentAssessment } from '../assessment/interfaces/studentAssessment'

import { getStudentAssessment } from '../assessment/network/queries/getStudentAssessment'
import { GradeSuggestionBadge } from '../assessment/pages/components/badges'
import { Link } from 'react-router-dom'

export const StudentDetail: React.FC<CoursePhaseStudentIdentifierProps> = ({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  studentId,
  coursePhaseId,
  courseId,
  courseParticipationId,
}) => {
  const { data: studentAssessment, isPending } = useQuery<StudentAssessment>({
    queryKey: ['assessments', coursePhaseId, courseParticipationId],
    queryFn: () => getStudentAssessment(coursePhaseId, courseParticipationId),
    enabled: Boolean(courseParticipationId),
  })

  if (isPending) return null
  if (!studentAssessment) return null

  const completion = studentAssessment.assessmentCompletion
  if (!completion) return null

  const completionStatus = completion.completed ? 'Completed' : 'Not completed'
  const completedAt = completion.completedAt
    ? new Date(completion.completedAt).toLocaleString()
    : '-'
  const comment = completion.comment || '-'
  const author = completion.author || '-'

  return (
    <div className='grid grid-cols-2 gap-y-4 gap-x-4 text-sm'>
      <div className='flex flex-col min-h-[92px]'>
        <div className='text-muted-foreground text-sm'>{completionStatus}</div>
        {completionStatus === 'Completed' && (
          <div className='font-medium break-words'>{completedAt}</div>
        )}
        <div className='mt-1'>
          <GradeSuggestionBadge gradeSuggestion={completion.gradeSuggestion} text={true} />
        </div>
      </div>

      <div className='flex flex-col items-end text-right min-h-[92px]'>
        <div>
          <h4 className='text-muted-foreground text-sm'>Author</h4>
          <div className='font-medium'>{author}</div>
        </div>
      </div>

      <div className='col-span-2 mt-2 space-y-2'>
        <hr />
        <div>
          <h4 className='text-muted-foreground text-sm mb-1'>Comment</h4>
          <div className='whitespace-pre-wrap break-words'>{comment}</div>
        </div>
      </div>
      <Link
        to={`/management/course/${courseId}/${coursePhaseId}/participants/${courseParticipationId}`}
        className='mt-2 text-blue-600'
      >
        Assessment
      </Link>
    </div>
  )
}
