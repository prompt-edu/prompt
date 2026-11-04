import { getStudentName, ParticipantNavigation, StickyHeader } from '@tumaet/prompt-ui-components'

import type { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'
import { useParticipantNavigation } from '../../hooks/useParticipantNavigation'

import { AssessmentProfile } from './AssessmentProfile'
import { StudentAssessmentBadges } from './StudentAssessmentBadges'

interface AssessmentHeaderProps {
  participant: AssessmentParticipationWithStudent
  studentAssessment: StudentAssessment
  remainingAssessments: number
}

export const AssessmentHeader = ({
  participant,
  studentAssessment,
  remainingAssessments,
}: AssessmentHeaderProps) => {
  const { orderedParticipations, navigateToParticipation } = useParticipantNavigation()

  const studentName = getStudentName(participant.student)

  return (
    <StickyHeader
      floating
      expandedContent={
        <AssessmentProfile
          participant={participant}
          studentAssessment={studentAssessment}
          remainingAssessments={remainingAssessments}
        />
      }
    >
      {(docked) => (
        <ParticipantNavigation
          participants={orderedParticipations}
          currentId={participant.courseParticipationID}
          onNavigate={navigateToParticipation}
          colorByStatus
          floating={docked}
        >
          {docked && (
            <>
              <span className='truncate font-semibold'>{studentName}</span>
              <StudentAssessmentBadges
                studentAssessment={studentAssessment}
                remainingAssessments={remainingAssessments}
                variant='compact'
                className='flex-nowrap hidden sm:flex xl:hidden'
              />
              <StudentAssessmentBadges
                studentAssessment={studentAssessment}
                remainingAssessments={remainingAssessments}
                className='flex-nowrap hidden xl:flex'
              />
            </>
          )}
        </ParticipantNavigation>
      )}
    </StickyHeader>
  )
}
