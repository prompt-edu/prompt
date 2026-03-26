import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@tumaet/prompt-ui-components'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { Plus } from 'lucide-react'
import { useParams } from 'react-router-dom'

interface AddQuestionMenuProps {
  setApplicationQuestions: React.Dispatch<
    React.SetStateAction<
      (ApplicationQuestionText | ApplicationQuestionMultiSelect | ApplicationQuestionFileUpload)[]
    >
  >
  applicationQuestions: (
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload
  )[]
}

export const AddQuestionMenu = ({
  setApplicationQuestions,
  applicationQuestions,
}: AddQuestionMenuProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const handleAddNewQuestionText = () => {
    const newQuestion: ApplicationQuestionText = {
      id: `not-valid-id-question-${applicationQuestions.length + 1}`,
      title: ``,
      coursePhaseID: phaseId!,
      isRequired: false,
      orderNum: applicationQuestions.length + 1,
      allowedLength: 500,
      accessKey: '',
      accessibleForOtherPhases: false,
    }
    setApplicationQuestions([...applicationQuestions, newQuestion])
  }

  const handleAddNewQuestionMultiSelect = () => {
    const newQuestion: ApplicationQuestionMultiSelect = {
      id: `not-valid-id-question-${applicationQuestions.length + 1}`,
      title: ``,
      coursePhaseID: phaseId!,
      isRequired: false,
      orderNum: applicationQuestions.length + 1,
      minSelect: 0,
      maxSelect: 0,
      options: [],
      accessKey: '',
      accessibleForOtherPhases: false,
    }
    setApplicationQuestions([...applicationQuestions, newQuestion])
  }

  const handleAddNewCheckbox = () => {
    const newQuestion: ApplicationQuestionMultiSelect = {
      id: `not-valid-id-question-${applicationQuestions.length + 1}`,
      title: ``,
      coursePhaseID: phaseId!,
      isRequired: false,
      orderNum: applicationQuestions.length + 1,
      placeholder: 'CheckBoxQuestion',
      minSelect: 0,
      maxSelect: 1,
      options: ['Yes'],
      accessKey: '',
      accessibleForOtherPhases: false,
    }
    setApplicationQuestions([...applicationQuestions, newQuestion])
  }

  const handleAddNewFileUpload = () => {
    const newQuestion: ApplicationQuestionFileUpload = {
      id: `not-valid-id-question-${applicationQuestions.length + 1}`,
      title: ``,
      coursePhaseID: phaseId!,
      isRequired: false,
      orderNum: applicationQuestions.length + 1,
      allowedFileTypes: '.pdf',
      maxFileSizeMB: 10,
      accessKey: '',
      accessibleForOtherPhases: false,
    }
    setApplicationQuestions([...applicationQuestions, newQuestion])
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size='sm'>
          <Plus className='mr-2 h-4 w-4' />
          Add Question
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        <DropdownMenuItem onSelect={() => handleAddNewQuestionText()}>
          Text Question
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => handleAddNewQuestionMultiSelect()}>
          Multi-Select Question
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => handleAddNewCheckbox()}>Checkbox</DropdownMenuItem>
        <DropdownMenuItem onSelect={() => handleAddNewFileUpload()}>File Upload</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
