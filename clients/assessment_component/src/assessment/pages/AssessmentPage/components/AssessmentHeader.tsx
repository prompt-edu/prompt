import { Button, cn } from '@tumaet/prompt-ui-components'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

import type { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'
import { StickyHeader } from '../../components/StickyHeader'

import { useParticipantNavigation } from '../hooks/useParticipantNavigation'

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
  const navigate = useNavigate()
  const { prevMember, nextMember } = useParticipantNavigation()

  const hasNavigation = Boolean(prevMember && nextMember)
  const studentName = `${participant.student.firstName} ${participant.student.lastName}`

  return (
    <StickyHeader
      expandedContent={
        <AssessmentProfile
          participant={participant}
          studentAssessment={studentAssessment}
          remainingAssessments={remainingAssessments}
        />
      }
    >
      {(docked) => (
        <div className='flex items-center gap-2'>
          {hasNavigation && (
            <Button
              variant='outline'
              className='h-10 shrink-0'
              aria-label={`Navigate to previous participant: ${prevMember?.firstName} ${prevMember?.lastName}`}
              onClick={() => navigate(`../${prevMember?.id}`, { relative: 'path' })}
            >
              <ChevronLeft className='h-4 w-4' />
              <span className='hidden md:inline'>
                {prevMember?.firstName} {prevMember?.lastName}
              </span>
            </Button>
          )}

          {/* Name + badges. Sits between the buttons only while docked; fades
                in as the full card below collapses, so it appears to move up. */}
          <div
            className={cn(
              'flex min-w-0 flex-1 items-center justify-center gap-2 px-2 transition-opacity duration-300',
              docked ? 'opacity-100' : 'pointer-events-none opacity-0',
            )}
          >
            <span className='truncate font-semibold'>{studentName}</span>
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
          </div>

          {hasNavigation && (
            <Button
              variant='outline'
              className='h-10 shrink-0'
              aria-label={`Navigate to next participant: ${nextMember?.firstName} ${nextMember?.lastName}`}
              onClick={() => navigate(`../${nextMember?.id}`, { relative: 'path' })}
            >
              <span className='hidden md:inline'>
                {nextMember?.firstName} {nextMember?.lastName}
              </span>
              <ChevronRight className='h-4 w-4' />
            </Button>
          )}
        </div>
      )}
    </StickyHeader>
  )
}
