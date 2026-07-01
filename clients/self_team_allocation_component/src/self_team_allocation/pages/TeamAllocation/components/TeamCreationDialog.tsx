import { Team } from '@tumaet/prompt-shared-state'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Input,
} from '@tumaet/prompt-ui-components'
import { PlusCircle } from 'lucide-react'
import { useState } from 'react'

interface Props {
  disabled?: boolean
  teams: Team[]
  onCreate: (name: string) => void
}

export const TeamCreationDialog = ({ disabled, teams, onCreate }: Props) => {
  const [name, setName] = useState('')
  const [open, setOpen] = useState(false) // Manage dialog open state

  // Check if the entered name already exists (case-insensitive)
  const nameExists = teams.some((team) => team.name.toLowerCase() === name.trim().toLowerCase())

  const handleCreate = (newName: string) => {
    onCreate(newName)
    setName('') // Clear input field after creation
    setOpen(false) // Close the dialog
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className='w-full max-w-lg mx-auto' disabled={disabled}>
          <PlusCircle className='mr-2 h-4 w-4' />
          Create new team
        </Button>
      </DialogTrigger>
      <DialogContent>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            const trimmedName = name.trim()
            if (!trimmedName || nameExists) return
            handleCreate(trimmedName)
          }}
        >
          <DialogHeader>
            <DialogTitle>Create new team</DialogTitle>
            <DialogDescription>The new team will automatically include you.</DialogDescription>
          </DialogHeader>

          <Input
            id='teamName'
            placeholder='Enter team name'
            className={`my-4 ${
              nameExists ? 'border-red-500 focus:border-red-500 focus:ring-red-500' : ''
            }`}
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={disabled}
          />
          {nameExists && (
            <p className='text-sm text-red-600'>A team with this name already exists.</p>
          )}

          <DialogFooter>
            <Button type='submit' disabled={!name.trim() || nameExists || disabled}>
              Create team
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
