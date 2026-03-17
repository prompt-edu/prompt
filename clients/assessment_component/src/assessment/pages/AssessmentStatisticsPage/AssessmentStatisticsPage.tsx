import { Loader2 } from 'lucide-react'
import { useState, useMemo } from 'react'

import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'

import { useCategoryStore } from '../../zustand/useCategoryStore'
import { useParticipationStore } from '../../zustand/useParticipationStore'
import { useScoreLevelStore } from '../../zustand/useScoreLevelStore'
import { useTeamStore } from '../../zustand/useTeamStore'

import { useGetAllAssessments } from '../hooks/useGetAllAssessments'
import { useGetAllAssessmentCompletions } from '../hooks/useGetAllAssessmentCompletions'

import { useGetParticipationsWithAssessment } from '../components/diagrams/hooks/useGetParticipationsWithAssessment'
import { useFilteredParticipations } from './hooks/useFilteredParticipations'

import { GradeDistributionDiagram } from '../components/diagrams/GradeDistributionDiagram'

import { ScoreLevelDistributionDiagram } from '../components/diagrams/ScoreLevelDistributionDiagram'
import { GenderDiagram } from '../components/diagrams/GenderDiagram'
import { AuthorDiagram } from '../components/diagrams/AuthorDiagram'
import { CategoryDiagram } from '../components/diagrams/CategoryDiagram'
import { NationalityDiagram } from '../components/diagrams/NationalityDiagram'
import { TeamDiagram } from '../components/diagrams/TeamDiagram'

import { FilterMenu, StatisticsFilter } from './components/FilterMenu'
import { FilterBadges } from './components/FilterBadges'

export const AssessmentStatisticsPage = () => {
  const [filters, setFilters] = useState<StatisticsFilter>({})

  const { categories } = useCategoryStore()
  const { participations } = useParticipationStore()
  const { scoreLevels } = useScoreLevelStore()
  const { teams } = useTeamStore()

  const {
    data: assessments,
    isPending: isAssessmentsPending,
    isError: isAssessmentsError,
    refetch: refetchAssessments,
  } = useGetAllAssessments()

  const {
    data: assessmentCompletions,
    isPending: isAssessmentCompletionsPending,
    isError: isAssessmentCompletionsError,
    refetch: refetchAssessmentCompletions,
  } = useGetAllAssessmentCompletions()

  const participationsWithAssessments = useGetParticipationsWithAssessment(
    participations || [],
    scoreLevels || [],
    assessmentCompletions || [],
    assessments || [],
  )

  const { filteredParticipations, filteredParticipationWithAssessments } =
    useFilteredParticipations({
      participations,
      assessmentCompletions: assessmentCompletions || null,
      participationsWithAssessments,
      filters,
    })

  const filteredGrades = filteredParticipationWithAssessments
    .map((p) => p.assessmentCompletion?.gradeSuggestion)
    .filter((p): p is number => p !== undefined)

  const filteredScoreLevels = useMemo(
    () =>
      filteredParticipationWithAssessments.map((p) => ({
        courseParticipationID: p.participation.courseParticipationID,
        scoreLevel: p.scoreLevel,
        scoreNumeric: p.scoreNumeric,
      })),
    [filteredParticipationWithAssessments],
  )

  const isError = isAssessmentsError || isAssessmentCompletionsError
  const isPending = isAssessmentsPending || isAssessmentCompletionsPending

  const refetch = () => {
    refetchAssessments()
    refetchAssessmentCompletions()
  }

  if (isError) {
    return <ErrorPage message='Error loading assessments' onRetry={refetch} />
  }
  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Assessment Statistics</ManagementPageHeader>
      <div className='space-y-2'>
        <div className='flex justify-between items-end gap-2'>
          <div className='flex flex-col sm:flex-row items-start sm:items-center space-y-2 sm:space-y-0 sm:space-x-4'>
            <FilterMenu filters={filters} setFilters={setFilters} teams={teams} />
          </div>
          <div className='text-sm text-muted-foreground'>
            Filters will be applied to all diagrams. Currently showing{' '}
            {filteredParticipations.length} of {participations?.length ?? 0} participants.
          </div>
        </div>

        <FilterBadges filters={filters} onRemoveFilter={setFilters} teams={teams} />
        <GradeDistributionDiagram participations={filteredParticipations} grades={filteredGrades} />
      </div>

      <h1 className='text-xl font-semibold'>Detailed Grade Statistics</h1>
      <div className='grid gap-4 grid-cols-1 lg:grid-cols-2 2xl:grid-cols-4 mb-6'>
        <GenderDiagram
          participationsWithAssessment={filteredParticipationWithAssessments}
          showGrade={true}
        />
        <AuthorDiagram
          participationsWithAssessment={filteredParticipationWithAssessments}
          showGrade
        />
        <NationalityDiagram
          participationsWithAssessment={filteredParticipationWithAssessments}
          showGrade
        />
        {teams && teams.length > 0 && (
          <TeamDiagram
            participationsWithAssessment={filteredParticipationWithAssessments}
            teams={teams}
            showGrade
          />
        )}
      </div>

      <h1 className='text-xl font-semibold'>Detailed Score Level Statistics</h1>
      <div className='grid gap-4 grid-cols-1 lg:grid-cols-2 2xl:grid-cols-4 mb-6'>
        <ScoreLevelDistributionDiagram
          participations={filteredParticipations}
          scoreLevels={filteredScoreLevels}
        />
        <GenderDiagram participationsWithAssessment={filteredParticipationWithAssessments} />
        <CategoryDiagram categories={categories} assessments={assessments} />
        <AuthorDiagram participationsWithAssessment={filteredParticipationWithAssessments} />
        <NationalityDiagram participationsWithAssessment={filteredParticipationWithAssessments} />
        {teams && teams.length > 0 && (
          <TeamDiagram
            participationsWithAssessment={filteredParticipationWithAssessments}
            teams={teams}
          />
        )}
      </div>
    </div>
  )
}
