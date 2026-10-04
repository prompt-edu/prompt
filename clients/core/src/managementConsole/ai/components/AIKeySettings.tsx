import { useAIEnabled } from '@core/network/hooks/useAIEnabled'
import { SettingsCard } from '@tumaet/prompt-ui-components'
import { Sparkles } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useCoursePhases } from '../hooks/useCoursePhases'
import { PhaseKeyRow } from './PhaseKeyRow'

export const AIKeySettings = () => {
  const { courseId } = useParams<{ courseId: string }>()
  const phases = useCoursePhases(courseId)
  const aiEnabled = useAIEnabled()

  if (!aiEnabled || phases.length === 0) {
    return null
  }

  return (
    <SettingsCard
      title='AI'
      description='A course phase uses AI only with a Logos key of its own. Its usage is metered per key.'
      icon={<Sparkles />}
    >
      <div className='px-5 pb-5'>
        <div className='px-2 w-full flex flex-col border border-border rounded-md divide-y'>
          {phases.map((phase) => (
            <PhaseKeyRow key={phase.id} phaseId={phase.id} phaseName={phase.name} />
          ))}
        </div>
      </div>
    </SettingsCard>
  )
}
