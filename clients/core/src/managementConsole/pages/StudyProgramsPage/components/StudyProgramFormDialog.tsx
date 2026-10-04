import {
  Alert,
  AlertDescription,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import type { CreateStudyProgram } from '../../../shared/interfaces/StudyProgram'
import {
  renameWarning,
  reservedStudyProgramError,
  type StudyProgramWithStudentCount,
} from '../utils/studyProgramStudents'

const MAX_NAME_LENGTH = 100
const MAX_SHORT_NAME_LENGTH = 20

interface StudyProgramFormDialogProps {
  title: string
  editing?: StudyProgramWithStudentCount
  isPending: boolean
  onClose: () => void
  onSubmit: (values: CreateStudyProgram) => void
}

export const StudyProgramFormDialog = ({
  title,
  editing,
  isPending,
  onClose,
  onSubmit,
}: StudyProgramFormDialogProps) => {
  const [name, setName] = useState(editing?.name ?? '')
  const [shortName, setShortName] = useState(editing?.shortName ?? '')

  const trimmedName = name.trim()
  const nameError = reservedStudyProgramError(name)
  const shortNameError = reservedStudyProgramError(shortName)
  const isReserved = nameError !== undefined || shortNameError !== undefined
  const isRename = editing !== undefined && trimmedName !== '' && trimmedName !== editing.name

  const handleSubmit = () => {
    if (trimmedName && !isReserved) {
      onSubmit({ name: trimmedName, shortName: shortName.trim() })
    }
  }

  return (
    <Dialog open onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <div className='flex flex-col gap-4 py-2'>
          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='study-program-name'>Name</Label>
            <Input
              id='study-program-name'
              value={name}
              maxLength={MAX_NAME_LENGTH}
              onChange={(e) => setName(e.target.value)}
              placeholder='e.g. Computer Science'
            />
            {nameError && <p className='text-destructive text-sm'>{nameError}</p>}
          </div>
          <div className='flex flex-col gap-1.5'>
            <Label htmlFor='study-program-short-name'>Short Name (optional)</Label>
            <Input
              id='study-program-short-name'
              value={shortName}
              maxLength={MAX_SHORT_NAME_LENGTH}
              onChange={(e) => setShortName(e.target.value)}
              placeholder='e.g. CS'
            />
            {shortNameError ? (
              <p className='text-destructive text-sm'>{shortNameError}</p>
            ) : (
              <p className='text-muted-foreground text-sm'>
                Shown in charts where the full name does not fit.
              </p>
            )}
          </div>
          {isRename && editing.studentCount !== 0 && (
            <Alert>
              <AlertDescription>
                {renameWarning(editing.studentCount, editing.name, trimmedName)}
              </AlertDescription>
            </Alert>
          )}
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!trimmedName || isReserved || isPending}>
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
