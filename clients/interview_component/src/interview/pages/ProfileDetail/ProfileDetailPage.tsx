import { type ExportedApplicationAnswer, PassStatus } from '@tumaet/prompt-shared-state'
import {
  Badge,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  ExportedApplicationAnswerTable,
  getStudentName,
  ParticipantNavigation,
  ProfilePicture,
  StickyHeader,
} from '@tumaet/prompt-ui-components'
import { FileUserIcon, Mic } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { InterviewCard } from '../../components/InterviewCard'
import { StudentCard } from '../../components/StudentCard'
import { StudentHistoryCard } from '../../components/StudentHistoryCard'
import { useInterviewNavigation } from '../../hooks/useInterviewNavigation'
import { useCoursePhaseStore } from '../../zustand/useCoursePhaseStore'
import { useParticipationStore } from '../../zustand/useParticipationStore'

export interface ProfileDetailLocationState {
  tableSearch?: string
  /** The participants table's order when the profile was opened, for previous / next. */
  orderedParticipationIds?: string[]
}

/** The interview decides on acceptance, so the status reads like the application's. */
const statusBadges: Record<PassStatus, { label: string; className: string }> = {
  [PassStatus.PASSED]: { label: 'Accepted', className: 'bg-green-500 hover:bg-green-500' },
  [PassStatus.FAILED]: { label: 'Rejected', className: 'bg-red-500 hover:bg-red-500' },
  [PassStatus.NOT_ASSESSED]: { label: 'Not Assessed', className: 'bg-gray-500 hover:bg-gray-500' },
}

export const ProfileDetailPage = () => {
  const { studentId } = useParams<{ studentId: string }>()
  const { participations, interviewReviews } = useParticipationStore()
  const { coursePhase } = useCoursePhaseStore()
  const showStudentHistory = coursePhase?.restrictedData?.showStudentHistory === true
  const participation = participations.find((p) => p.student.id === studentId)
  const { orderedParticipations, currentParticipationId, navigateToParticipation } =
    useInterviewNavigation()

  const applicationAnswers =
    (participation?.prevData?.applicationAnswers as ExportedApplicationAnswer[]) ?? []

  return (
    <div>
      <div className='mb-8'>
        <StickyHeader
          floating
          expandedContent={participation && <StudentCard participation={participation} />}
        >
          {(docked) => (
            <ParticipantNavigation
              participants={orderedParticipations}
              currentId={currentParticipationId}
              onNavigate={navigateToParticipation}
              colorByStatus
              compact={docked}
            >
              {participation && (
                <>
                  <ProfilePicture
                    studentId={participation.student.id}
                    firstName={participation.student.firstName}
                    lastName={participation.student.lastName}
                    size='sm'
                    className='shrink-0'
                  />
                  <span className='truncate font-semibold'>
                    {getStudentName(participation.student)}
                  </span>
                  <div className='hidden shrink-0 items-center gap-1 sm:flex'>
                    <Badge className={statusBadges[participation.passStatus].className}>
                      {statusBadges[participation.passStatus].label}
                    </Badge>
                    <Badge variant='outline' className='gap-1' title='Application score'>
                      <FileUserIcon className='h-3 w-3' />
                      {participation.prevData?.score ?? 'N/A'}
                    </Badge>
                    <Badge variant='outline' className='gap-1' title='Interview score'>
                      <Mic className='h-3 w-3' />
                      {interviewReviews[participation.courseParticipationID]?.score ?? 'N/A'}
                    </Badge>
                  </div>
                </>
              )}
            </ParticipantNavigation>
          )}
        </StickyHeader>
        {!participation && (
          <div className='flex justify-center items-center h-64'>
            <p className='text-lg text-muted-foreground'>Participant not found</p>
          </div>
        )}
      </div>
      {participation && (
        <div className='grid grid-cols-1 lg:grid-cols-2 xl:has-[>[data-student-history]]:grid-cols-[1fr_1fr_20rem] gap-8'>
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center'>
                <FileUserIcon className='h-5 w-5 mr-2' />
                Application
              </CardTitle>
            </CardHeader>
            <CardContent>
              {applicationAnswers.length === 0 ? (
                <div>
                  No Application Exported - Please check your export Settings in Application
                  Configuration
                </div>
              ) : (
                <ExportedApplicationAnswerTable applicationAnswers={applicationAnswers} />
              )}
            </CardContent>
          </Card>
          <InterviewCard />
          {showStudentHistory && participation.student.id && (
            <StudentHistoryCard studentId={participation.student.id} />
          )}
        </div>
      )}
    </div>
  )
}
