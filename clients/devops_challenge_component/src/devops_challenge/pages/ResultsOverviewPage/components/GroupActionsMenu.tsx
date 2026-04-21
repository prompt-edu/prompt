import { useState } from 'react'
import { MoreHorizontal, CheckCircle, XCircle } from 'lucide-react'
import { RowModel } from '@tanstack/react-table'
import { PassStatus, CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { useUpdateCoursePhaseParticipationBatch } from '@/hooks/useUpdateCoursePhaseParticipationBatch'
import { DeveloperWithInfo } from '../../../interfaces/DeveloperWithInfo'
import { ActionDialog } from '@/components/table/GroupActionDialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Button,
} from '@tumaet/prompt-ui-components'

interface GroupActionsMenuProps {
  selectedRows: RowModel<{
    participation: CoursePhaseParticipationWithStudent
    profile: DeveloperWithInfo | undefined
  }>
  onClose: () => void
}

export const GroupActionsMenu = ({ selectedRows, onClose }: GroupActionsMenuProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const [dialogState, setDialogState] = useState<{
    type: 'setPassed' | 'setFailed' | null
    isOpen: boolean
  }>({ type: null, isOpen: false })

  const openDialog = (type: 'setPassed' | 'setFailed') => {
    setIsOpen(false)
    setDialogState({ type, isOpen: true })
  }

  const closeDialog = () => setDialogState({ type: null, isOpen: false })

  // modifiers
  const { mutate: mutateUpdateCoursePhaseParticipationBatch } =
    useUpdateCoursePhaseParticipationBatch()
  const numberOfRowsSelected = selectedRows.rows.length

  return (
    <>
      <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
        <DropdownMenuTrigger asChild>
          <Button disabled={numberOfRowsSelected < 1}>
            <MoreHorizontal className='h-4 w-4' />
            Actions
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' className='w-48'>
          <DropdownMenuLabel>Group Actions</DropdownMenuLabel>
          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={() => openDialog('setPassed')}>
            <CheckCircle className='mr-2 h-4 w-4' />
            Set Accepted
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => openDialog('setFailed')}>
            <XCircle className='mr-2 h-4 w-4' />
            Set Rejected
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      {dialogState.isOpen && dialogState.type === 'setPassed' && (
        <ActionDialog
          title='Confirm Set Passed'
          description={`Are you sure you want to mark ${numberOfRowsSelected} applications as accepted?`}
          confirmLabel='Set Accepted'
          isOpen={dialogState.type === 'setPassed' && dialogState.isOpen}
          onClose={closeDialog}
          onConfirm={() => {
            mutateUpdateCoursePhaseParticipationBatch(
              selectedRows.rows.map((row) => {
                return {
                  coursePhaseID: row.original.participation.coursePhaseID,
                  courseParticipationID: row.original.participation.courseParticipationID,
                  passStatus: PassStatus.PASSED,
                  restrictedData: row.original.participation.restrictedData,
                  studentReadableData: row.original.participation.studentReadableData,
                }
              }),
            )
            onClose()
          }}
        />
      )}

      {dialogState.isOpen && dialogState.type === 'setFailed' && (
        <ActionDialog
          title='Confirm Set Failed'
          description={`Are you sure you want to mark ${numberOfRowsSelected} applications as rejected?`}
          confirmLabel='Set Rejected'
          isOpen={dialogState.type === 'setFailed' && dialogState.isOpen}
          onClose={closeDialog}
          onConfirm={() => {
            mutateUpdateCoursePhaseParticipationBatch(
              selectedRows.rows.map((row) => {
                return {
                  coursePhaseID: row.original.participation.coursePhaseID,
                  courseParticipationID: row.original.participation.courseParticipationID,
                  passStatus: PassStatus.FAILED,
                  restrictedData: row.original.participation.restrictedData,
                  studentReadableData: row.original.participation.studentReadableData,
                }
              }),
            )
            onClose()
          }}
        />
      )}
    </>
  )
}
