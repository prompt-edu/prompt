import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { CoursePhaseMailing } from '@tumaet/prompt-ui-components'
import { useCoursePhaseStore } from '../../zustand/useCoursePhaseStore'

export const MailingPage = () => {
  const { coursePhase } = useCoursePhaseStore()
  return (
    <div>
      <ManagementPageHeader>Mailing</ManagementPageHeader>
      <CoursePhaseMailing coursePhase={coursePhase} />
    </div>
  )
}
