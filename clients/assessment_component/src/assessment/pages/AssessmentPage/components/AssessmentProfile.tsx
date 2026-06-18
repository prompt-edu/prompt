import { Book, Calendar, GraduationCap } from 'lucide-react'

import { Card, CardContent, CardHeader } from '@tumaet/prompt-ui-components'

import { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'

import { StudentAssessmentBadges } from './StudentAssessmentBadges'

interface AssessmentProfileProps {
  participant: AssessmentParticipationWithStudent
  studentAssessment: StudentAssessment
  remainingAssessments: number
}

export const AssessmentProfile = ({
  participant,
  studentAssessment,
  remainingAssessments,
}: AssessmentProfileProps) => {
  return (
    <Card className='relative overflow-hidden'>
      <CardHeader>
        <div className='flex flex-wrap items-center gap-2'>
          <h1 className='text-2xl font-bold mr-2 sm:text-left sm:w-auto text-center w-full'>
            {participant.student.firstName} {participant.student.lastName}
          </h1>

          <StudentAssessmentBadges
            studentAssessment={studentAssessment}
            remainingAssessments={remainingAssessments}
          />
        </div>
      </CardHeader>

      <CardContent>
        <div className='grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm'>
          <div className='flex items-center'>
            <Book className='w-5 h-5 mr-2 text-primary' />
            <strong className='mr-2'>Program:</strong>
            <span className='text-muted-foreground'>
              {participant.student.studyProgram || 'N/A'}
            </span>
          </div>
          <div className='flex items-center'>
            <GraduationCap className='w-5 h-5 mr-2 text-primary' />
            <strong className='mr-2'>Degree:</strong>
            <span className='text-muted-foreground'>
              {participant.student.studyDegree
                ? participant.student.studyDegree.charAt(0).toUpperCase() +
                  participant.student.studyDegree.slice(1)
                : 'N/A'}
            </span>
          </div>
          <div className='flex items-center'>
            <Calendar className='w-5 h-5 mr-2 text-primary' />
            <strong className='mr-2'>Semester:</strong>
            <span className='text-muted-foreground'>
              {participant.student.currentSemester || 'N/A'}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
