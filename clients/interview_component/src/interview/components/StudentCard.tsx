import {
  type CoursePhaseParticipationWithStudent,
  getGravatarUrl,
  getStatusColor,
  getStudyDegreeString,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  getStudentName,
  Separator,
} from '@tumaet/prompt-ui-components'
import { BookOpen, ExternalLink, FileUserIcon, GraduationCap, Mic } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import { getApplicationParticipantPath } from '../utils/getApplicationParticipantPath'
import { useParticipationStore } from '../zustand/useParticipationStore'

interface StudentCardProps {
  participation: CoursePhaseParticipationWithStudent
}

export function StudentCard({ participation }: StudentCardProps) {
  const { interviewReviews } = useParticipationStore()
  const navigate = useNavigate()
  const { courseId } = useParams<{ courseId: string }>()
  const { courses } = useCourseStore()

  const assessmentScore = participation.prevData?.score ?? 'N/A'
  const interviewScore = interviewReviews[participation.courseParticipationID]?.score ?? 'N/A'

  const applicationLink = getApplicationParticipantPath(
    courses,
    courseId,
    participation.courseParticipationID,
  )

  return (
    <Card className='h-full relative overflow-hidden'>
      <div className={`h-16 ${getStatusColor(participation.passStatus)}`} />

      <div className='mb-8'>
        <Avatar className='absolute w-24 h-24 border-4 border-background rounded-full transform left-3 -translate-y-1/2'>
          <AvatarImage
            src={getGravatarUrl(participation.student.email)}
            alt={participation.student.lastName}
          />
          <AvatarFallback className='rounded-full font-bold text-lg'>
            {participation.student.firstName[0]}
            {participation.student.lastName[0]}
          </AvatarFallback>
        </Avatar>
      </div>

      <CardHeader>
        <CardTitle className='text-left'>{getStudentName(participation.student)}</CardTitle>
      </CardHeader>

      <CardContent className='grid gap-2'>
        <div className='flex items-center gap-2'>
          <GraduationCap className='h-4 w-4 text-muted-foreground' />
          <span className='text-sm'>{getStudyDegreeString(participation.student.studyDegree)}</span>
        </div>
        <div className='flex items-center gap-2'>
          <BookOpen className='h-4 w-4 text-muted-foreground' />
          <span className='text-sm'>{participation.student.studyProgram}</span>
        </div>
        <Separator />

        <div className='grid gap-2'>
          <h4 className='text-sm font-semibold'>Scores</h4>
          <div className='grid gap-2 sm:grid-cols-2'>
            <div className='flex items-center gap-3'>
              <FileUserIcon className='h-4 w-4 text-muted-foreground mr-2' />
              <div className='flex flex-col'>
                <span className='text-xs text-muted-foreground'>Application</span>
                <span className='text-sm'>{assessmentScore}</span>
              </div>
            </div>
            <div className='flex items-center gap-3'>
              <Mic className='h-4 w-4 text-muted-foreground mr-2' />
              <div className='flex flex-col'>
                <span className='text-xs text-muted-foreground'>Interview</span>
                <span className='text-sm font-medium'>{interviewScore}</span>
              </div>
            </div>
          </div>
        </div>

        {applicationLink && (
          <>
            <Separator />
            <Button
              variant='outline'
              size='sm'
              className='w-full'
              onClick={(e) => {
                e.stopPropagation()
                navigate(applicationLink)
              }}
            >
              <ExternalLink className='h-4 w-4 mr-2' />
              View full application
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  )
}
