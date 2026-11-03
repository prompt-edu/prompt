import { Button, Checkbox, getStudentName, Label, MultiSelect } from '@tumaet/prompt-ui-components'
import type { TutorCandidate } from '../../../interfaces/tutor'
import { isSelectableTutor } from '../utils/tutorImport'

interface TutorSelectionProps {
  tutors: TutorCandidate[]
  selectedTutors: string[]
  onTutorToggle: (tutorId: string) => void
  onSelectAll: () => void
  teamOptions: { id: string; name: string }[]
  teamAssignments: Record<string, string[]>
  onTeamChange: (tutorId: string, teamIds: string[]) => void
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
  const selectableCount = tutors.filter(isSelectableTutor).length
  const options = teamOptions.map((team) => ({ label: team.name, value: team.id }))
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
          const selectable = isSelectableTutor(tutor)
          const selected = selectedTutors.includes(tutor.id)
          return (
            <div key={tutor.id} className='flex flex-col gap-1 py-2'>
              <div className='flex items-center space-x-2'>
                <Checkbox
                  id={`tutor-${tutor.id}`}
                  checked={selected}
                  disabled={!selectable}
                  onCheckedChange={() => selectable && onTutorToggle(tutor.id)}
                />
                <Label
                  htmlFor={`tutor-${tutor.id}`}
                  className='flex-1 cursor-pointer aria-disabled:cursor-not-allowed aria-disabled:text-muted-foreground'
                  aria-disabled={!selectable}
                >
                  {getStudentName(tutor)}
                  {!selectable && ' (no university login)'}
                </Label>
              </div>
              {selected && (
                <MultiSelect
                  aria-label={`Teams of ${getStudentName(tutor)}`}
                  options={options}
                  defaultValue={teamAssignments[tutor.id] ?? []}
                  onValueChange={(teamIds) => onTeamChange(tutor.id, teamIds)}
                  placeholder='Assign teams'
                  maxCount={3}
                  modalPopover
                />
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
