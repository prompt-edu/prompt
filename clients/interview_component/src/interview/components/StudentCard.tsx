import { GraduationCap, BookOpen, Mic, FileUserIcon, Calendar, Clock, MapPin } from 'lucide-react'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Avatar,
  AvatarImage,
  AvatarFallback,
  Separator,
} from '@tumaet/prompt-ui-components'
import {
  getStudyDegreeString,
  CoursePhaseParticipationWithStudent,
} from '@tumaet/prompt-shared-state'
import { getGravatarUrl } from '@/lib/getGravatarUrl'
import { getStatusColor } from '@/lib/getStatusColor'
import { format } from 'date-fns'

interface InterviewSlotData {
  id: string
  startTime: string
  endTime: string
  location: string | null
}

interface StudentCardProps {
  participation: CoursePhaseParticipationWithStudent
  interviewSlot?: InterviewSlotData
}

export function StudentCard({ participation, interviewSlot }: StudentCardProps) {
  const assessmentScore = participation.prevData?.score ?? 'N/A'
  const interviewScore = participation.restrictedData?.score ?? 'N/A'

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
        <CardTitle className='text-left'>
          {participation.student.firstName} {participation.student.lastName}
        </CardTitle>
        {interviewSlot && (
          <div className='mt-2 space-y-1'>
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <Calendar className='h-3 w-3' />
              <span>{format(new Date(interviewSlot.startTime), 'PPP')}</span>
            </div>
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <Clock className='h-3 w-3' />
              <span>
                {format(new Date(interviewSlot.startTime), 'p')} -{' '}
                {format(new Date(interviewSlot.endTime), 'p')}
              </span>
            </div>
            {interviewSlot.location && (
              <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                <MapPin className='h-3 w-3 flex-shrink-0' />
                {interviewSlot.location.match(/^https?:\/\//) ? (
                  <a
                    href={interviewSlot.location}
                    target='_blank'
                    rel='noopener noreferrer'
                    className='text-blue-600 hover:underline truncate min-w-0'
                  >
                    {interviewSlot.location}
                  </a>
                ) : (
                  <span className='truncate min-w-0'>{interviewSlot.location}</span>
                )}
              </div>
            )}
          </div>
        )}
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
      </CardContent>
    </Card>
  )
}
