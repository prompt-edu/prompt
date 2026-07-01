import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import { Button, DeleteConfirmation, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { RefreshCw, Users } from 'lucide-react'
import type React from 'react'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Timeframe } from '../../../interfaces/timeframe'
import { createTeam } from '../../../network/mutations/createTeam'
import { createTeamAssignment } from '../../../network/mutations/createTeamAllocation'
import { deleteTeam } from '../../../network/mutations/deleteTeam'
import { deleteTeamAssignment } from '../../../network/mutations/deleteTeamAllocation'
import { DeadlineDisplay } from './DeadlineDisplay'
import { TeamCard } from './TeamCard'
import { TeamCreationDialog } from './TeamCreationDialog'
import { TutorImportDialog } from './TutorImportDialog'

interface Props {
  teams: Team[]
  courseParticipationID?: string
  refetchTeams: () => void
  timeframe: Timeframe
}

export const TeamSelection: React.FC<Props> = ({
  teams,
  courseParticipationID,
  refetchTeams,
  timeframe,
}) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const isLecturer = !courseParticipationID
  const queryClient = useQueryClient()

  const [submitError, setSubmitError] = useState<string | null>(null)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [teamToDelete, setTeamToDelete] = useState<Team | null>(null)

  const joinMutation = useMutation({
    mutationFn: (teamId: string) => createTeamAssignment(phaseId ?? '', teamId),
    onSuccess: () => {
      setSubmitError(null)
      queryClient.invalidateQueries({ queryKey: ['self_team_allocations', phaseId] })
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    onError: () => setSubmitError('Failed to join team. Please try again.'),
  })

  const leaveMutation = useMutation({
    mutationFn: (teamId: string) => deleteTeamAssignment(phaseId ?? '', teamId),
    onSuccess: () => {
      setSubmitError(null)
      queryClient.invalidateQueries({ queryKey: ['self_team_allocations', phaseId] })
    },
    onError: () => setSubmitError('Failed to leave team. Please try again.'),
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => createTeam(phaseId ?? '', [name]),
    onSuccess: () => {
      setSubmitError(null)
      queryClient.invalidateQueries({ queryKey: ['self_team_allocations', phaseId] })
    },
    onError: () => setSubmitError('Failed to create team. Please try again.'),
  })

  const deleteMutation = useMutation({
    mutationFn: (teamId: string) => deleteTeam(phaseId ?? '', teamId),
    onSuccess: () => {
      setSubmitError(null)
      queryClient.invalidateQueries({ queryKey: ['self_team_allocations', phaseId] })
    },
    onError: () => setSubmitError('Failed to delete team. Please try again.'),
  })

  const myTeam = useMemo(
    () => teams.find((team) => team.members.some((m) => m.id === courseParticipationID)),
    [teams, courseParticipationID],
  )

  const disabled =
    timeframe.timeframeSet && (timeframe.endTime < new Date() || timeframe.startTime > new Date())
  const joiningDisabled = Boolean(myTeam) || isLecturer

  const sortedTeams = useMemo(
    () => [...teams].sort((a, b) => (myTeam?.id === a.id ? -1 : myTeam?.id === b.id ? 1 : 0)),
    [teams, myTeam],
  )

  return (
    <div className='container mx-auto py-6 space-y-8'>
      <div>
        <ManagementPageHeader>Team Allocation</ManagementPageHeader>
        {timeframe.timeframeSet && <DeadlineDisplay timeframe={timeframe} />}
      </div>

      {!disabled && (
        <div className='flex flex-col sm:flex-row items-start gap-3'>
          <TeamCreationDialog
            onCreate={(name) => createMutation.mutate(name)}
            teams={teams}
            disabled={myTeam?.id !== undefined || isLecturer}
          />
          {isLecturer && <TutorImportDialog />}
          <Button variant='outline' onClick={refetchTeams}>
            <RefreshCw className='h-4 w-4' />
            Refresh Teams
          </Button>
        </div>
      )}

      {submitError && (
        <div className='p-4 border border-red-200 bg-red-50 text-red-700 rounded-md'>
          {submitError}
        </div>
      )}

      {sortedTeams.length > 0 ? (
        <div className='grid gap-6 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
          {sortedTeams.map((team) => (
            <TeamCard
              key={team.id}
              team={team}
              isMember={myTeam?.id === team.id}
              full={team.members.length >= 3}
              isLecturer={isLecturer}
              joiningDisabled={joiningDisabled}
              disabled={disabled}
              courseParticipationID={courseParticipationID}
              onJoin={(id) => joinMutation.mutate(id)}
              onLeave={(id) => leaveMutation.mutate(id)}
              onDelete={(delTeam) => {
                setTeamToDelete(delTeam)
                setDeleteDialogOpen(true)
              }}
            />
          ))}
        </div>
      ) : (
        <div className='flex flex-col items-center justify-center p-12 border rounded-lg'>
          <Users className='h-12 w-12 text-muted-foreground mb-4 mt-4' />
          <h3 className='text-lg font-medium'>No Teams Available</h3>
          <p className='text-muted-foreground'>Create a new team to get started.</p>
        </div>
      )}

      <DeleteConfirmation
        isOpen={deleteDialogOpen}
        setOpen={setDeleteDialogOpen}
        onClick={(ok) => ok && teamToDelete && deleteMutation.mutate(teamToDelete.id)}
        deleteMessage='Are you sure you want to delete this team?'
      />
    </div>
  )
}
