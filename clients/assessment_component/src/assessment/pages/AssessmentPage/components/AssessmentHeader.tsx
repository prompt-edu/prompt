import { Button, cn } from '@tumaet/prompt-ui-components'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import type { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'

import { useParticipantNavigation } from '../hooks/useParticipantNavigation'

import { AssessmentProfile } from './AssessmentProfile'
import { StudentAssessmentBadges } from './StudentAssessmentBadges'

// The global management header (`h-14`) stays at the top; the bar docks beneath it.
const HEADER_OFFSET_PX = 56
const UNDOCK_HYSTERESIS_PX = 8

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

  const placeholderRef = useRef<HTMLDivElement>(null)
  const barRef = useRef<HTMLDivElement>(null)
  const dockedRef = useRef(false)
  const [docked, setDocked] = useState(false)

  // CSS `position: sticky` is unreliable here because the core's scroll
  // container (`#management-children`, `overflow-auto`) sits between this bar
  // and the viewport. Instead we pin the bar with `position: fixed` once it
  // reaches the header. `getBoundingClientRect` + a capture-phase scroll
  // listener work no matter which ancestor actually scrolls.
  useEffect(() => {
    const placeholder = placeholderRef.current
    const bar = barRef.current
    if (!placeholder || !bar) return

    const update = () => {
      const { top, left, width } = placeholder.getBoundingClientRect()

      // Small hysteresis buffer so a docked bar doesn't jitter at the threshold.
      const shouldDock = dockedRef.current
        ? top <= HEADER_OFFSET_PX + UNDOCK_HYSTERESIS_PX
        : top <= HEADER_OFFSET_PX

      if (shouldDock) {
        placeholder.style.height = `${bar.offsetHeight}px`
        bar.style.position = 'fixed'
        bar.style.top = `${HEADER_OFFSET_PX}px`
        bar.style.left = `${left}px`
        bar.style.width = `${width}px`
      } else {
        placeholder.style.height = ''
        bar.style.cssText = ''
      }

      dockedRef.current = shouldDock
      setDocked(shouldDock)
    }

    update()
    window.addEventListener('scroll', update, true)
    window.addEventListener('resize', update)
    return () => {
      window.removeEventListener('scroll', update, true)
      window.removeEventListener('resize', update)
    }
  }, [])

  const hasNavigation = Boolean(prevMember && nextMember)
  const studentName = `${participant.student.firstName} ${participant.student.lastName}`

  return (
    <>
      <div ref={placeholderRef}>
        <div
          ref={barRef}
          className={cn(
            'z-20 transition-colors duration-300',
            docked && 'rounded-md bg-background shadow-sm',
          )}
        >
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
        </div>
      </div>

      {/* Full profile card. Collapses to zero height while docked and scrolls
          away under the pinned bar. */}
      <div
        className={cn(
          'grid transition-all duration-300',
          docked ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100',
        )}
      >
        <div className='overflow-hidden'>
          <AssessmentProfile
            participant={participant}
            studentAssessment={studentAssessment}
            remainingAssessments={remainingAssessments}
          />
        </div>
      </div>
    </>
  )
}
