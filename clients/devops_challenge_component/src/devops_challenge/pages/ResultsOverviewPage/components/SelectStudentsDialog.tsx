import React from 'react'
import { Table } from '@tanstack/react-table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Button,
} from '@tumaet/prompt-ui-components'

interface SelectStudentsDialogProps {
  isOpen: boolean
  onClose: () => void
  selectCount: number
  setSelectCount: (count: number) => void
  table: Table<any>
  setRowSelection: React.Dispatch<React.SetStateAction<Record<string, boolean>>>
  studentsPassedChallengeCount: number
}

export const SelectStudentsDialog: React.FC<SelectStudentsDialogProps> = ({
  isOpen,
  onClose,
  selectCount,
  setSelectCount,
  table,
  setRowSelection,
  studentsPassedChallengeCount,
}: SelectStudentsDialogProps) => {
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Select Students</DialogTitle>
          <DialogDescription>
            Specify how many rows to select (based on their challenge passing position). Currently{' '}
            {studentsPassedChallengeCount} student
            {studentsPassedChallengeCount === 1 ? ' has' : 's have'} passed the challenge.
          </DialogDescription>
        </DialogHeader>
        <Input
          placeholder='Select first ... students'
          value={selectCount}
          type='number'
          min='0'
          max={studentsPassedChallengeCount}
          onChange={(e) => {
            const value = parseInt(e.target.value)
            if (isNaN(value) || value < 0 || value > studentsPassedChallengeCount) {
              e.target.setCustomValidity(
                `Please enter a number between 0 and ${studentsPassedChallengeCount} (max passed students count).`,
              )
            } else {
              e.target.setCustomValidity('')
              setSelectCount(value)
            }
            e.target.reportValidity()
          }}
        />
        <DialogFooter>
          <Button variant='outline' onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={() => {
              const newSelection: Record<string, boolean> = {}
              table.getFilteredRowModel().rows.forEach((row) => {
                const profile = row.original.profile
                if (profile && profile.passingPosition && profile.passingPosition <= selectCount) {
                  newSelection[row.id] = true
                }
              })
              setRowSelection(newSelection)
              onClose()
            }}
          >
            Select
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
