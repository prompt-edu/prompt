import {
  Button,
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  Separator,
} from '@tumaet/prompt-ui-components'
import { GraduationCap } from 'lucide-react'
import { translations } from '@tumaet/prompt-shared-state'
import { useNavigate } from 'react-router-dom'

interface ApplicationLoginCardProps {
  universityLoginAvailable: boolean
  externalStudentsAllowed: boolean
  onContinueWithoutLogin: (isExternalStudent: boolean) => void
}

export const ApplicationLoginCard = ({
  universityLoginAvailable,
  externalStudentsAllowed,
  onContinueWithoutLogin,
}: ApplicationLoginCardProps) => {
  const navigate = useNavigate()
  const path = window.location.pathname

  return (
    <Card>
      <CardHeader>
        <CardTitle>Please log in to continue:</CardTitle>
      </CardHeader>
      <CardContent className='space-y-6'>
        {!externalStudentsAllowed && (
          <p className='text-yellow-600 bg-yellow-50 p-3 rounded-md'>
            This course is only open for {translations.university.name} students.
            {universityLoginAvailable &&
              ' If you cannot log in, please reach out to the instructor of the course.'}
          </p>
        )}
        <div className='space-y-4'>
          <p>
            Are you a {translations.university.name} student?
            {universityLoginAvailable &&
              ` Then please log in using your ${translations.university['login-name']}.`}
          </p>
          <Button
            className='w-full bg-[#0065BD] hover:bg-[#005299] text-white'
            size='lg'
            onClick={() => {
              if (universityLoginAvailable) {
                navigate(path + '/authenticated')
              } else {
                onContinueWithoutLogin(false)
              }
            }}
          >
            <GraduationCap className='mr-2 h-5 w-5' />
            {universityLoginAvailable
              ? `Log in with ${translations.university['login-name']}`
              : `Continue as ${translations.university.name} student`}
          </Button>
        </div>
        {externalStudentsAllowed && (
          <>
            <Separator className='my-4' />
            <div className='space-y-4'>
              <p>Are you an external student?</p>
              <Button
                variant='outline'
                className='w-full'
                onClick={() => onContinueWithoutLogin(true)}
              >
                Continue without a {translations.university.name}-Account
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
