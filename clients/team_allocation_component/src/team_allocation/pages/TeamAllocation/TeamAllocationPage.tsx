import type React from 'react'
import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Loader2, Users } from 'lucide-react'

import type { CoursePhaseParticipationsWithResolution, Student } from '@tumaet/prompt-shared-state'
import type { Team } from '@tumaet/prompt-shared-state'
import type { Allocation } from '../../interfaces/allocation'

import { getAllTeams } from '../../network/queries/getAllTeams'
import { getCoursePhaseParticipations } from '@tumaet/prompt-shared-state'
import { getTeamAllocations } from '../../network/queries/getTeamAllocations'
import { AllocationSummaryCard } from './components/AllocationSummaryCard'

import { getGravatarUrl } from '@tumaet/prompt-shared-state'

import {
  ErrorPage,
  ManagementPageHeader,
  Card,
  CardContent,
  CardHeader,
  Badge,
  Separator,
  Avatar,
  AvatarFallback,
  AvatarImage,
} from '@tumaet/prompt-ui-components'

export const TeamAllocationPage: React.FC = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: fetchedTeams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useQuery<Team[]>({
    queryKey: ['team_allocation_team', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const {
    data: coursePhaseParticipations,
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })

  const {
    data: teamAllocations,
    isPending: isTeamAllocationsPending,
    isError: isTeamAllocationsError,
    refetch: refetchTeamAllocations,
  } = useQuery<Allocation[]>({
    queryKey: ['team_allocations', phaseId],
    queryFn: () => getTeamAllocations(phaseId ?? ''),
  })

  const participationMap = useMemo(() => {
    const map = new Map<string, Student>()
    coursePhaseParticipations?.participations.forEach((p) => {
      map.set(p.courseParticipationID, p.student)
    })
    return map
  }, [coursePhaseParticipations])

  const teamsWithMembersAndTutors = useMemo(() => {
    if (!fetchedTeams) return []

    return fetchedTeams.map((team) => {
      const allocation = teamAllocations?.find((a) => a.projectId === team.id)

      const members = allocation
        ? (allocation.students
            .map((cpId) => participationMap.get(cpId))
            .filter(Boolean) as Student[])
        : []

      const tutors = team.tutors ?? []

      return {
        id: team.id,
        name: team.name,
        members,
        tutors,
      }
    })
  }, [fetchedTeams, teamAllocations, participationMap])

  const isPending = isTeamsPending || isCoursePhaseParticipationsPending || isTeamAllocationsPending
  const isError = isTeamsError || isParticipationsError || isTeamAllocationsError

  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  if (isError) {
    return (
      <ErrorPage
        onRetry={() => {
          refetchTeams()
          refetchCoursePhaseParticipations()
          refetchTeamAllocations()
        }}
      />
    )
  }

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Team Allocations</ManagementPageHeader>
      <AllocationSummaryCard
        coursePhaseParticipations={coursePhaseParticipations}
        teamAllocations={teamAllocations}
      />

      {teamsWithMembersAndTutors.length === 0 ? (
        <Card className='bg-muted/40'>
          <CardContent className='pt-6 flex flex-col items-center justify-center text-center p-8'>
            <div className='rounded-full bg-muted p-3 mt-4 mb-4'>
              <Users className='h-8 w-8 text-muted-foreground' />
            </div>
            <h3 className='text-lg font-medium mb-2'>No Teams Created Yet</h3>
            <p className='text-muted-foreground mb-4'>Start by creating teams.</p>
          </CardContent>
        </Card>
      ) : (
        <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6'>
          {teamsWithMembersAndTutors.map((team) => (
            <Card
              key={team.id}
              className='overflow-hidden border-l-4 hover:shadow-md transition-shadow duration-200'
            >
              <CardHeader className='pb-2'>
                <div className='flex justify-between items-center'>
                  <h3 className='text-lg font-semibold tracking-tight'>{team.name}</h3>
                  <Badge variant='outline' className='ml-2'>
                    <Users className='h-3.5 w-3.5 mr-1' />
                    {team.members.length}
                  </Badge>
                </div>
              </CardHeader>
              <Separator />
              <CardContent className='pt-4 space-y-4'>
                <div>
                  <p className='text-sm font-medium mb-1'>Tutors:</p>
                  {team.tutors.length === 0 ? (
                    <p className='text-sm italic text-muted-foreground'>No tutors allocated yet.</p>
                  ) : (
                    <ul className='space-y-2'>
                      {team.tutors.map((tutor) => (
                        <li key={tutor.id} className='text-sm flex items-center gap-2'>
                          <Avatar className='h-6 w-6'>
                            <AvatarFallback className='text-xs font-medium'>
                              {tutor.firstName[0]}
                              {tutor.lastName[0]}
                            </AvatarFallback>
                          </Avatar>
                          <span>
                            {tutor.firstName} {tutor.lastName}
                          </span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>

                <div>
                  <p className='text-sm font-medium mb-1'>Members:</p>
                  {team.members.length === 0 ? (
                    <p className='text-sm italic text-muted-foreground'>
                      No members allocated yet.
                    </p>
                  ) : (
                    <ul className='space-y-2'>
                      {team.members.map((member) => (
                        <li key={member.id} className='text-sm flex items-center gap-2'>
                          <Avatar className='h-6 w-6'>
                            <AvatarImage
                              src={getGravatarUrl(member.email)}
                              alt={`${member.firstName} ${member.lastName}`}
                            />
                            <AvatarFallback className='text-xs font-medium'>
                              {member.firstName[0]}
                              {member.lastName[0]}
                            </AvatarFallback>
                          </Avatar>
                          <span>
                            {member.firstName} {member.lastName}
                          </span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
