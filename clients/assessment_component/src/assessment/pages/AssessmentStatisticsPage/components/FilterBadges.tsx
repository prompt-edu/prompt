import { Gender, getGenderString, Team } from '@tumaet/prompt-shared-state'
import { StatisticsFilter } from './FilterMenu'
import { FilterBadge } from '@tumaet/prompt-ui-components'

interface FilterBadgesProps {
  filters: StatisticsFilter
  onRemoveFilter: React.Dispatch<React.SetStateAction<StatisticsFilter>>
  teams: Team[]
}

export const FilterBadges = ({ filters, onRemoveFilter, teams }: FilterBadgesProps) => {
  const handleRemoveGenderFilter = (gender: Gender) => {
    onRemoveFilter((prevFilters) => {
      const currentGenders = prevFilters.genders || []
      const newGenders = currentGenders.filter((g) => g !== gender)
      return {
        ...prevFilters,
        genders: newGenders.length > 0 ? newGenders : undefined,
      }
    })
  }

  const handleRemoveTeamFilter = (teamId: string) => {
    onRemoveFilter((prevFilters) => {
      const currentTeams = prevFilters.teams || []
      const newTeams = currentTeams.filter((t) => t !== teamId)
      return {
        ...prevFilters,
        teams: newTeams.length > 0 ? newTeams : undefined,
      }
    })
  }

  const handleRemoveSemesterFilter = () => {
    onRemoveFilter((prevFilters) => {
      const newFilters = { ...prevFilters }
      delete newFilters.semester
      return newFilters
    })
  }

  const activeBadges: Array<{ key: string; label: string; onRemove: () => void }> = []

  if (filters.genders) {
    filters.genders.forEach((gender) => {
      activeBadges.push({
        key: `gender-${gender}`,
        label: `Gender: ${getGenderString(gender)}`,
        onRemove: () => handleRemoveGenderFilter(gender),
      })
    })
  }

  if (filters.teams) {
    filters.teams.forEach((teamId) => {
      const team = teams.find((t) => t.id === teamId)
      const teamName = team ? team.name : 'Unknown Team'
      activeBadges.push({
        key: `team-${teamId}`,
        label: `Team: ${teamName}`,
        onRemove: () => handleRemoveTeamFilter(teamId),
      })
    })
  }

  if (filters.semester && (filters.semester.min || filters.semester.max)) {
    const semesterLabel = `Semester: ${filters.semester.min || '0'} - ${filters.semester.max || '∞'}`
    activeBadges.push({
      key: 'semester',
      label: semesterLabel,
      onRemove: handleRemoveSemesterFilter,
    })
  }

  return (
    <div className='flex flex-wrap gap-2 h-5'>
      {activeBadges.map(({ key, label, onRemove }) => (
        <FilterBadge key={key} label={label} onRemove={onRemove} />
      ))}
    </div>
  )
}
