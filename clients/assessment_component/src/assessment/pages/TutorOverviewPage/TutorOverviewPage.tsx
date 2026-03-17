import { ReactNode, useMemo } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'

import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { PromptTable } from '@tumaet/prompt-ui-components'

import { useTeamStore } from '../../zustand/useTeamStore'
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'

import type { ColumnDef } from '@tanstack/react-table'

interface TutorRow {
  id: string
  firstName: string
  lastName: string
  teamName: string
}

export const TutorOverviewPage = (): ReactNode => {
  const navigate = useNavigate()
  const path = useLocation().pathname

  const { teams } = useTeamStore()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

  const data: TutorRow[] = useMemo(() => {
    return teams.flatMap((team) =>
      team.tutors
        .filter((tutor) => tutor.id)
        .map((tutor) => ({
          id: tutor.id!,
          firstName: tutor.firstName,
          lastName: tutor.lastName,
          teamName: team.name,
        })),
    )
  }, [teams])

  const columns: ColumnDef<TutorRow>[] = useMemo(
    () => [
      {
        accessorKey: 'firstName',
        header: 'First Name',
      },
      {
        accessorKey: 'lastName',
        header: 'Last Name',
      },
      {
        accessorKey: 'teamName',
        header: 'Team',
      },
    ],
    [],
  )

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Tutor Overview</ManagementPageHeader>

      {coursePhaseConfig?.tutorEvaluationEnabled && (
        <p className='text-sm text-muted-foreground mb-4'>
          Click on a tutor to view their evaluation results from students.
        </p>
      )}

      <PromptTable<TutorRow>
        data={data}
        columns={columns}
        onRowClick={(row) => {
          if (coursePhaseConfig?.tutorEvaluationEnabled) {
            navigate(`${path}/${row.id}`)
          }
        }}
      />
    </div>
  )
}
