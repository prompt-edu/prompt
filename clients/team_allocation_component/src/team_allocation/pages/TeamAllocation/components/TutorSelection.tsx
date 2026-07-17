import type { Student } from '@tumaet/prompt-shared-state'
import {
  Button,
  Checkbox,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'

interface TutorSelectionProps {
  tutors: Student[]
  selectedTutors: string[]
  onTutorToggle: (tutorId: string) => void
  onSelectAll: () => void
  teamOptions: { id: string; name: string }[]
  teamAssignments: Record<string, string>
  onTeamChange: (tutorId: string, teamId: string) => void
}

export const TutorSelection = ({
  tutors,
  selectedTutors,
  onTutorToggle,
  onSelectAll,
  teamOptions,
  teamAssignments,
  onTeamChange,
}: TutorSelectionProps) => {
  const selectableCount = tutors.filter((t) => t.universityLogin).length
  return (
    <div className='grid gap-2'>
      <div className='flex items-center justify-between'>
        <Label>Select Tutors</Label>
        <Button variant='outline' size='sm' onClick={onSelectAll} disabled={selectableCount === 0}>
          {selectableCount > 0 && selectedTutors.length === selectableCount
            ? 'Deselect All'
            : 'Select All'}
        </Button>
      </div>
      <div className='overflow-y-auto border rounded-md p-2 max-h-[300px]'>
        {tutors.map((tutor) => {
          const hasLogin = !!tutor.universityLogin
          return (
            <div key={tutor.id} className='flex flex-col gap-1 py-2'>
              <div className='flex items-center space-x-2'>
                <Checkbox
                  id={`tutor-${tutor.id}`}
                  checked={selectedTutors.includes(tutor.id ?? '')}
                  disabled={!hasLogin}
                  onCheckedChange={() => hasLogin && tutor.id && onTutorToggle(tutor.id)}
                />
                <Label
                  htmlFor={`tutor-${tutor.id}`}
                  className='flex-1 cursor-pointer aria-disabled:cursor-not-allowed aria-disabled:text-muted-foreground'
                  aria-disabled={!hasLogin}
                >
                  {tutor.firstName} {tutor.lastName}
                  {!hasLogin && ' (no university login)'}
                </Label>
              </div>
              {selectedTutors.includes(tutor.id ?? '') && (
                <Select
                  value={teamAssignments[tutor.id ?? ''] || ''}
                  onValueChange={(val) => tutor.id && onTeamChange(tutor.id, val)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder='Assign team' />
                  </SelectTrigger>
                  <SelectContent>
                    {teamOptions.map((team) => (
                      <SelectItem key={team.id} value={team.id}>
                        {team.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
