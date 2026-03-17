import {
  Student,
  PassStatus,
  getGenderString,
  getStudyDegreeString,
} from '@tumaet/prompt-shared-state'
import { Mail, Flag, Book, GraduationCap, Calendar, Hash, Users, KeyRound } from 'lucide-react'
import { Card, CardContent, CardHeader } from '@tumaet/prompt-ui-components'
import { getCountryName } from '@/lib/getCountries'
import translations from '@/lib/translations.json'
import { getStatusColor } from '@/lib/getStatusColor'
import { ProfilePicture } from './StudentProfilePicture'

interface StudentProfileProps {
  student: Student
  status?: PassStatus
}

export const StudentProfile = ({ student, status }: StudentProfileProps) => {
  return (
    <Card className='relative overflow-hidden'>
      {/* Status indicator */}
      <div className={`h-16 ${status ? getStatusColor(status) : 'bg-gray-100 dark:bg-muted'}`} />
      <div className='mb-4'>
        <ProfilePicture
          email={student.email}
          firstName={student.firstName}
          lastName={student.lastName}
          size='lg'
          className='absolute border-4 border-background rounded-full transform left-3 -translate-y-1/2'
        />
      </div>
      <CardHeader>
        <div className='flex flex-col sm:flex-row items-center'>
          {/* Avatar */}
          <div className='flex-1 text-start sm:text-left'>
            <div className='flex flex-col items-start sm:items-start pt-6'>
              <h1 className='text-2xl font-bold'>
                {student.firstName} {student.lastName}
              </h1>
            </div>
            <div className='mt-2 grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm text-muted-foreground'>
              <div className='space-y-1'>
                <p className='flex items-center justify-center sm:justify-start'>
                  <Mail className='w-4 h-4 mr-2' />
                  <span className='truncate'>{student.email}</span>
                </p>
                {student.nationality && (
                  <p className='flex items-center justify-center sm:justify-start'>
                    <Flag className='w-4 h-4 mr-2' />
                    {getCountryName(student.nationality)}
                  </p>
                )}
                {student.gender && (
                  <p className='flex items-center justify-center sm:justify-start'>
                    <Users className='w-4 h-4 mr-2' />
                    {getGenderString(student.gender)}
                  </p>
                )}
              </div>
              <div className='space-y-1'>
                {student.universityLogin && (
                  <p className='flex items-center justify-center sm:justify-start'>
                    <KeyRound className='w-4 h-4 mr-2' />
                    <strong className='mr-1'>{translations.university['login-name']}:</strong>
                    {student.universityLogin}
                  </p>
                )}
                {student.matriculationNumber && (
                  <p className='flex items-center justify-center sm:justify-start'>
                    <Hash className='w-4 h-4 mr-2' />
                    <strong className='mr-1'>Matriculation:</strong>
                    {student.matriculationNumber}
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className='grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm'>
          <div className='flex items-center'>
            <Book className='w-5 h-5 mr-2 text-primary' />
            <strong className='mr-2'>Program:</strong>
            <span className='text-muted-foreground'>{student.studyProgram || 'N/A'}</span>
          </div>
          <div className='flex items-center'>
            <GraduationCap className='w-5 h-5 mr-2 text-primary' />
            <strong className='mr-2'>Degree:</strong>
            <span className='text-muted-foreground'>
              {getStudyDegreeString(student.studyDegree) || 'N/A'}
            </span>
          </div>
          <div className='flex items-center'>
            <Calendar className='w-5 h-5 mr-2 text-primary' />
            <strong className='mr-2'>Semester:</strong>
            <span className='text-muted-foreground'>{student.currentSemester || 'N/A'}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
