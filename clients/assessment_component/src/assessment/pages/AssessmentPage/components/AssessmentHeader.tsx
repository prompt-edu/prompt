import { getStatusBadge, ParticipantNavigation, StickyHeader } from '@tumaet/prompt-ui-components'

import type { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'
import { ParticipantProfile } from '../../components/ParticipantProfile'
import { useParticipantNavigation } from '../../hooks/useParticipantNavigation'

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

  return (
    <StickyHeader
      floating
      expandedContent={
        <ParticipantProfile participant={participant}>
          {getStatusBadge(participant.passStatus)}
          <StudentAssessmentBadges
            studentAssessment={studentAssessment}
            remainingAssessments={remainingAssessments}
          />
        </ParticipantProfile>
      }
    >
      {(docked) => (
        <ParticipantNavigation
          participants={orderedParticipations}
          currentId={participant.courseParticipationID}
          onNavigate={navigateToParticipation}
          colorByStatus
          compact={docked}
        >
          {getStatusBadge(participant.passStatus)}
          <StudentAssessmentBadges
            studentAssessment={studentAssessment}
            remainingAssessments={remainingAssessments}
            variant='compact'
            className='flex-nowrap xl:hidden'
          />
          <StudentAssessmentBadges
            studentAssessment={studentAssessment}
            remainingAssessments={remainingAssessments}
            className='flex-nowrap hidden xl:flex'
          />
        </ParticipantNavigation>
      )}
    </StickyHeader>
  )
}
