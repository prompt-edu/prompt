import type { ExportedApplicationAnswer } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  ExportedApplicationAnswerTable,
  ParticipantNavigation,
} from '@tumaet/prompt-ui-components'
import { FileUserIcon } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { InterviewCard } from '../../components/InterviewCard'
import { StudentCard } from '../../components/StudentCard'
import { useInterviewNavigation } from '../../hooks/useInterviewNavigation'
import { useParticipationStore } from '../../zustand/useParticipationStore'

export const ProfileDetailPage = () => {
  const { studentId } = useParams<{ studentId: string }>()
  const { participations } = useParticipationStore()
  const participation = participations.find((p) => p.student.id === studentId)
  const { orderedParticipations, currentParticipationId, navigateToParticipation } =
    useInterviewNavigation()

  const applicationAnswers =
    (participation?.prevData?.applicationAnswers as ExportedApplicationAnswer[]) ?? []

  return (
    <div className=''>
      <div className='pb-4'>
        <ParticipantNavigation
          participants={orderedParticipations}
          currentId={currentParticipationId}
          onNavigate={navigateToParticipation}
          colorByStatus
        />
        {!participation && (
          <div className='flex justify-center items-center h-64'>
            <p className='text-lg text-muted-foreground'>Participant not found</p>
          </div>
        )}
      </div>
      {participation && (
        <>
          <div className='pt-6 mb-8'>
            <StudentCard participation={participation} />
          </div>
          <div className='grid grid-cols-1 lg:grid-cols-2 gap-8'>
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
          </div>
        </>
      )}
    </div>
  )
}
