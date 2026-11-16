import type { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import type { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import type { Student } from '@tumaet/prompt-shared-state'
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  ScrollArea,
} from '@tumaet/prompt-ui-components'
import { Eye } from 'lucide-react'
import { useState } from 'react'
import { ApplicationWelcomeText } from '../../components/ApplicationWelcomeText'
import { ApplicationFormView } from '../ApplicationForm/ApplicationFormView'

interface ApplicationPreviewProps {
  questionsText: ApplicationQuestionText[]
  questionsMultiSelect: ApplicationQuestionMultiSelect[]
  questionsFileUpload: ApplicationQuestionFileUpload[]
  welcomeText?: string
}

export const ApplicationPreview = ({
  questionsText,
  questionsMultiSelect,
  questionsFileUpload,
  welcomeText,
}: ApplicationPreviewProps) => {
  const [dialogOpen, setDialogOpen] = useState(false)

  const initialStudent: Student = {
    firstName: 'Demo',
    lastName: 'Student',
    email: 'example@application.com',
    hasUniversityAccount: true,
  }
  return (
    <Dialog open={dialogOpen} onOpenChange={(isOpen) => setDialogOpen(isOpen)}>
      <DialogTrigger asChild>
        <Button size='sm'>
          <Eye className='mr-2 h-4 w-4' />
          Preview Application
        </Button>
      </DialogTrigger>
      <DialogContent className='w-[90vw] max-w-[900px] h-[90vh] [&>*]:min-w-0'>
        <DialogHeader>
          <DialogTitle>Application Preview</DialogTitle>
        </DialogHeader>
        <ScrollArea>
          <div className='space-y-6'>
            <ApplicationWelcomeText welcomeText={welcomeText} />
            <ApplicationFormView
              questionsText={questionsText}
              questionsMultiSelect={questionsMultiSelect}
              questionsFileUpload={questionsFileUpload}
              student={initialStudent}
              onSubmit={() => {}}
            />
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
