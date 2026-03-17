import React, { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Team, CoursePhaseParticipationsWithResolution, Student } from '@tumaet/prompt-shared-state'
import { CoursePhaseStudentIdentifierProps } from '@/interfaces/studentDetail'
import { Allocation } from '../team_allocation/interfaces/allocation'
import { getAllTeams } from '../team_allocation/network/queries/getAllTeams'
import { getTeamAllocations } from '../team_allocation/network/queries/getTeamAllocations'
import { getCoursePhaseParticipations } from '@/network/queries/getCoursePhaseParticipations'
import { RenderStudents } from '@/components/StudentAvatar'

export const StudentDetail: React.FC<CoursePhaseStudentIdentifierProps> = ({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  studentId,
  coursePhaseId,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  courseId: _courseId,
  courseParticipationId,
}) => {
  const { data: teams, isPending: isTeamsPending } = useQuery<Team[]>({
    queryKey: ['team_allocation_team', coursePhaseId],
    queryFn: () => getAllTeams(coursePhaseId),
  })

  const { data: coursePhaseParticipations, isPending: isParticipationsPending } =
    useQuery<CoursePhaseParticipationsWithResolution>({
      queryKey: ['participants', coursePhaseId],
      queryFn: () => getCoursePhaseParticipations(coursePhaseId),
    })

  const { data: teamAllocations, isPending: isAllocationsPending } = useQuery<Allocation[]>({
    queryKey: ['team_allocations', coursePhaseId],
    queryFn: () => getTeamAllocations(coursePhaseId),
  })

  const studentTeamInfo = useMemo(() => {
    if (!coursePhaseParticipations || !teamAllocations || !teams) {
      return null
    }

    // Use the passed courseParticipationId directly
    // Map course participation IDs to students
    const participationMap = new Map<string, Student>()
    coursePhaseParticipations.participations.forEach((p) => {
      participationMap.set(p.courseParticipationID, p.student)
    })

    // Find the allocation that contains this course participation ID
    const allocation = teamAllocations.find((a) => a.students.includes(courseParticipationId))

    if (!allocation) {
      return null
    }

    // Find the team for this allocation
    const team = teams.find((t) => t.id === allocation.projectId)

    if (!team) {
      return null
    }

    const members = allocation.students
      .map((cpId) => participationMap.get(cpId))
      .filter(Boolean) as Student[]

    const tutors = team.tutors ?? []

    return { team, members, tutors }
  }, [coursePhaseParticipations, teamAllocations, teams, courseParticipationId])

  const isPending = isTeamsPending || isAllocationsPending || isParticipationsPending

  if (isPending) {
    return null
  }

  if (!studentTeamInfo) {
    return null
  }

  const { team, members, tutors } = studentTeamInfo

  return (
    <div className='text-sm'>
      <h4 className='font-semibold text-lg'>{team.name}</h4>

      <RenderStudents
        className='mb-2 mt-1'
        students={tutors.map((tutor) => ({ ...tutor, email: 'no@mail.example' }))}
        fallback={<p className='text-muted-foreground'>No tutors in this team</p>}
      />

      <RenderStudents
        students={members}
        fallback={<p className='text-muted-foreground'>No members in this team</p>}
      />
    </div>
  )
}
