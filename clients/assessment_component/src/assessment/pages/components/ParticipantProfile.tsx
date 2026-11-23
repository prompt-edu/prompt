import type { Student } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardHeader,
  getStudentName,
  ProfilePicture,
} from '@tumaet/prompt-ui-components'
import { Book, Calendar, GraduationCap } from 'lucide-react'
import type { ReactNode } from 'react'

interface ParticipantProfileProps {
  courseParticipationId?: string
  person: Pick<Student, 'firstName' | 'lastName'>
  /** Program, degree and semester. Tutors have none, so they leave it out and the row is hidden. */
  studyDetails?: Pick<Student, 'studyProgram' | 'studyDegree' | 'currentSemester'>
  /** Names the page above the participant, e.g. "Self Evaluation". */
  label?: string
  /** Badges next to the name. */
  children?: ReactNode
}

export const ParticipantProfile = ({
  courseParticipationId,
  person,
  studyDetails,
  label,
  children,
}: ParticipantProfileProps) => {
  return (
    <Card>
      <CardHeader>
        {label && (
          <p className='text-sm font-medium text-muted-foreground text-center sm:text-left'>
            {label}
          </p>
        )}
        <div className='flex flex-wrap items-center gap-2'>
          <ProfilePicture
            courseParticipationId={courseParticipationId}
            firstName={person.firstName}
            lastName={person.lastName}
            size='lg'
            className='mx-auto sm:mx-0 sm:mr-2'
          />
          <h1 className='text-2xl font-bold w-full text-center sm:w-auto sm:text-left'>
            {getStudentName(person)}
          </h1>

          {children}
        </div>
      </CardHeader>

      {studyDetails && (
        <CardContent>
          <div className='grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm'>
            <div className='flex items-center'>
              <Book className='w-5 h-5 mr-2 text-primary' />
              <strong className='mr-2'>Program:</strong>
              <span className='text-muted-foreground'>{studyDetails.studyProgram || 'N/A'}</span>
            </div>
            <div className='flex items-center'>
              <GraduationCap className='w-5 h-5 mr-2 text-primary' />
              <strong className='mr-2'>Degree:</strong>
              <span className='text-muted-foreground'>
                {studyDetails.studyDegree
                  ? studyDetails.studyDegree.charAt(0).toUpperCase() +
                    studyDetails.studyDegree.slice(1)
                  : 'N/A'}
              </span>
            </div>
            <div className='flex items-center'>
              <Calendar className='w-5 h-5 mr-2 text-primary' />
              <strong className='mr-2'>Semester:</strong>
              <span className='text-muted-foreground'>{studyDetails.currentSemester || 'N/A'}</span>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  )
}
