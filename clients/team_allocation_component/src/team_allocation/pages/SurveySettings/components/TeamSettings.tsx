import type { Team } from '@tumaet/prompt-shared-state'
import { User, Users2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { createTeams } from '../../../network/mutations/createTeams'
import { deleteTeam } from '../../../network/mutations/deleteTeam'
import { updateTeam } from '../../../network/mutations/updateTeam'
import { EntitySettings } from './EntitySettings'

interface TeamSettingsProps {
  teams: Team[]
}

export const TeamSettings = ({ teams }: TeamSettingsProps) => {
  // Use the same phaseId context if needed (or adjust as appropriate)
  const phaseId = useParams<{ phaseId: string }>().phaseId ?? ''

  return (
    <EntitySettings<Team>
      items={teams}
      createFn={(names) => createTeams(phaseId, names)}
      updateFn={(id, newName) => updateTeam(phaseId, id, newName)}
      deleteFn={(id) => deleteTeam(phaseId, id)}
      queryKey={['team_allocation_team', phaseId]}
      title='Team'
      description='Manage your teams and their names'
      icon={<Users2 className='h-5 w-5' />}
      emptyIcon={<User className='h-12 w-12 mx-auto mb-3 opacity-20' />}
      emptyMessage='No teams created yet'
      emptySubtext='Add your first team using the form above'
    />
  )
}
