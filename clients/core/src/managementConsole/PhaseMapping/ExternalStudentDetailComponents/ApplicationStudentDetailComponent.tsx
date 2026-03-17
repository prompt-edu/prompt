import React, { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'
import type { ApplicationParticipation } from '../../applicationAdministration/interfaces/applicationParticipation'

import { getApplicationParticipations } from '@core/network/queries/applicationParticipations'
import { getStatusString } from '../../applicationAdministration/pages/ApplicationParticipantsPage/utils/getStatusBadge'
import { Link } from 'react-router-dom'

export const ApplicationStudentDetailComponent: React.FC<CoursePhaseStudentIdentifierProps> = ({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  studentId: _studentId,
  coursePhaseId,

  courseId,
  courseParticipationId,
}) => {
  const { data: participations, isPending } = useQuery<ApplicationParticipation[]>({
    // align with the key used in useGetApplicationParticipations() so we reuse cache
    queryKey: ['application_participations', 'students', coursePhaseId],
    queryFn: () => getApplicationParticipations(coursePhaseId),
  })

  const application = useMemo(() => {
    if (!participations) return null
    return participations.find((p) => p.courseParticipationID === courseParticipationId) ?? null
  }, [participations, courseParticipationId])

  if (isPending || !application) return null

  const applicationURL = `/management/course/${courseId}/${coursePhaseId}/participants/${courseParticipationId}`

  return (
    <div className='text-sm'>
      <div>
        <span className='text-muted-foreground'>Application: </span>
        <span className='font-medium'>{getStatusString(application.passStatus)}</span>
        <span className='text-muted-foreground'> · Score: </span>
        <span className='font-medium'>{application.score ?? '-'}</span>
      </div>
      <Link to={applicationURL} className='mt-2 text-blue-600'>
        Application
      </Link>
    </div>
  )
}
