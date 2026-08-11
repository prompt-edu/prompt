import { CoursePhaseMailing, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useCoursePhaseStore } from '../../zustand/useCoursePhaseStore'
import { QuestionConfiguration } from './QuestionConfiguration'

export const SettingsPage = () => {
  const { coursePhase } = useCoursePhaseStore()
  return (
    <div className='flex flex-col gap-8 p-4'>
      <ManagementPageHeader>Settings</ManagementPageHeader>
      <QuestionConfiguration />
      <section className='space-y-4'>
        <h2 className='text-2xl font-bold'>Mailing</h2>
        <CoursePhaseMailing coursePhase={coursePhase} />
      </section>
    </div>
  )
}
