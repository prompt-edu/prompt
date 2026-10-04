import type { ExportedApplicationAnswer } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  ExportedApplicationAnswerTable,
} from '@tumaet/prompt-ui-components'
import { ChevronLeft, FileUserIcon } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import { InterviewCard } from '../../components/InterviewCard'
import { StudentCard } from '../../components/StudentCard'
import { StudentHistoryCard } from '../../components/StudentHistoryCard'
import { useCoursePhaseStore } from '../../zustand/useCoursePhaseStore'
import { useParticipationStore } from '../../zustand/useParticipationStore'

export const ProfileDetailPage = () => {
  const { studentId } = useParams<{ studentId: string }>()
  const { participations } = useParticipationStore()
  const { coursePhase } = useCoursePhaseStore()
  const showStudentHistory = coursePhase?.restrictedData?.showStudentHistory === true
  const participation = participations.find((p) => p.student.id === studentId)
  const navigate = useNavigate()

  const applicationAnswers =
    (participation?.prevData?.applicationAnswers as ExportedApplicationAnswer[]) ?? []

  return (
    <div className=''>
      <div className='relative pb-4'>
        <Button
          onClick={() => navigate('..', { relative: 'path' })}
          variant='ghost'
          size='sm'
          className='absolute top-0 left-0'
        >
          <ChevronLeft className='h-4 w-4' />
          <span>Back</span>
        </Button>
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
          <div
            className={
              showStudentHistory
                ? 'grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-[1fr_1fr_20rem] gap-8'
                : 'grid grid-cols-1 lg:grid-cols-2 gap-8'
            }
          >
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
        </>
      )}
    </div>
  )
}
