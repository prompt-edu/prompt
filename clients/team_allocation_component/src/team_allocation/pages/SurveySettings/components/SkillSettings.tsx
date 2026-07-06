//
// SkillSettings component uses the generic EntitySettings for skills

import { Star } from 'lucide-react'
import { useParams } from 'react-router-dom'
import type { Skill } from '../../../interfaces/skill'
import { createSkills } from '../../../network/mutations/createSkills'
import { deleteSkill } from '../../../network/mutations/deleteSkill'
import { updateSkill } from '../../../network/mutations/updateSkill'
import { EntitySettings } from './EntitySettings'

interface SkillSettingsProps {
  skills: Skill[]
}

export const SkillSettings = ({ skills }: SkillSettingsProps) => {
  // Use the same phaseId context if needed (or adjust as appropriate)
  const phaseId = useParams<{ phaseId: string }>().phaseId ?? ''

  return (
    <EntitySettings<Skill>
      items={skills}
      createFn={(names) => createSkills(phaseId, names)}
      updateFn={(id, newName) => updateSkill(phaseId, id, newName)}
      deleteFn={(id) => deleteSkill(phaseId, id)}
      queryKey={['team_allocation_skill', phaseId]}
      title='Skill'
      description='Manage your skills and their names'
      icon={<Star className='h-5 w-5' />}
      emptyIcon={<Star className='h-12 w-12 mx-auto mb-3 opacity-20' />}
      emptyMessage='No skills created yet'
      emptySubtext='Add your first skill using the form above'
    />
  )
}
