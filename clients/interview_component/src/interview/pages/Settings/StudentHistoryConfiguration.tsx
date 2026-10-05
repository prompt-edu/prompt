import { useUpdateCoursePhaseMetaData } from '@tumaet/prompt-shared-state'
import { Label, Switch } from '@tumaet/prompt-ui-components'
import { useCoursePhaseStore } from '../../zustand/useCoursePhaseStore'

export const StudentHistoryConfiguration = () => {
  const { coursePhase } = useCoursePhaseStore()
  const { mutate, isPending, error } = useUpdateCoursePhaseMetaData()
  const showStudentHistory = coursePhase?.restrictedData?.showStudentHistory === true

  const handleChange = (checked: boolean) => {
    if (coursePhase) {
      mutate({ id: coursePhase.id, restrictedData: { showStudentHistory: checked } })
    }
  }

  return (
    <section className='space-y-4'>
      <header>
        <h2 className='text-2xl font-bold'>Student History</h2>
        <p className='text-muted-foreground'>
          Show the student notes and course history from the application assessment on the interview
          detail page.
        </p>
      </header>
      <div className='flex items-center space-x-2'>
        <Switch
          id='show-student-history'
          checked={showStudentHistory}
          onCheckedChange={handleChange}
          disabled={!coursePhase || isPending}
        />
        <Label htmlFor='show-student-history'>Show student notes and history</Label>
      </div>
      {error && <div className='text-xs text-destructive'>Error: {error.message}</div>}
    </section>
  )
}
