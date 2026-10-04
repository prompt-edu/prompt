import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ManagementPageHeader,
  type ParticipantRow,
  type TableFilter,
} from '@tumaet/prompt-ui-components'
import { useMemo, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { InterviewSlot } from '../../interfaces/InterviewSlots'
import { getApplicationParticipantPath } from '../../utils/getApplicationParticipantPath'
import { useParticipationStore } from '../../zustand/useParticipationStore'
import {
  createApplicationLinkColumn,
  createInterviewDayFilter,
  createInterviewSlotColumn,
  createLocationColumn,
  createScoreColumn,
} from './columns'

export const InterviewParticipantsPage = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const navigate = useNavigate()
  const openInNewTabRef = useRef(false)
  const { participations, interviewSlots, interviewReviews } = useParticipationStore()
  const { courses } = useCourseStore()

  const openInterview = (row: ParticipantRow) => {
    if (!row.student.id) return
    const target = `/management/course/${courseId}/${phaseId}/participants/${row.student.id}`
    if (openInNewTabRef.current) {
      window.open(`${window.location.origin}${target}`, '_blank', 'noopener,noreferrer')
      return
    }
    navigate(target)
  }

  const slotByParticipation = useMemo(() => {
    const slots = new Map<string, InterviewSlot>()
    interviewSlots.forEach((slot) => {
      if (slot.courseParticipationID) slots.set(slot.courseParticipationID, slot)
    })
    return slots
  }, [interviewSlots])

  const extraColumns = useMemo(() => {
    const participationById = new Map(participations.map((p) => [p.courseParticipationID, p]))
    const getApplicationScore = (courseParticipationID: string): number | null => {
      const score = participationById.get(courseParticipationID)?.prevData?.score
      return typeof score === 'number' ? score : null
    }

    const columns = [
      createInterviewSlotColumn(participations, slotByParticipation),
      createLocationColumn(participations, slotByParticipation),
      createScoreColumn(participations, {
        id: 'applicationScore',
        header: 'Application Score',
        getScore: getApplicationScore,
      }),
      createScoreColumn(participations, {
        id: 'interviewScore',
        header: 'Interview Score',
        getScore: (courseParticipationID) => interviewReviews[courseParticipationID]?.score ?? null,
      }),
      createApplicationLinkColumn(participations, (courseParticipationID) =>
        getApplicationParticipantPath(courses, courseId, courseParticipationID),
      ),
    ]
    return columns.filter((column) => column !== undefined)
  }, [participations, slotByParticipation, interviewReviews, courses, courseId])

  const extraFilters = useMemo<TableFilter<ParticipantRow>[]>(
    () => [
      createInterviewDayFilter(participations, slotByParticipation),
      {
        type: 'numericRange',
        id: 'interviewScore',
        label: 'Interview Score',
        noValueLabel: 'Not interviewed',
      },
      {
        type: 'numericRange',
        id: 'applicationScore',
        label: 'Application Score',
        noValueLabel: 'No score',
      },
    ],
    [participations, slotByParticipation],
  )

  return (
    <div>
      <ManagementPageHeader>Interview Participants</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        Click on a participant to open their interview. Cmd/Ctrl-click to open it in a new tab.
      </p>
      <div
        className='w-full'
        onClickCapture={(e) => {
          openInNewTabRef.current = e.metaKey || e.ctrlKey
        }}
      >
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={participations}
          extraColumns={extraColumns}
          extraFilters={extraFilters}
          onClickRowAction={openInterview}
        />
      </div>
    </div>
  )
}
