import { Gender, getGenderString, Team } from '@tumaet/prompt-shared-state'
import {
  Button,
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Input,
} from '@tumaet/prompt-ui-components'
import { Filter } from 'lucide-react'

export interface StatisticsFilter {
  genders?: Gender[]
  semester?: {
    min?: number
    max?: number
  }
  teams?: string[]
}

interface FilterMenuProps {
  filters: StatisticsFilter
  setFilters: React.Dispatch<React.SetStateAction<StatisticsFilter>>
  teams: Team[]
}

export const FilterMenu = ({ filters, setFilters, teams }: FilterMenuProps) => {
  const handleGenderFilterChange = (gender: Gender) => {
    setFilters((prevFilters) => {
      const currentGenders = prevFilters.genders || []
      const newGenders = currentGenders.includes(gender)
        ? currentGenders.filter((g) => g !== gender)
        : [...currentGenders, gender]

      return {
        ...prevFilters,
        genders: newGenders.length > 0 ? newGenders : undefined,
      }
    })
  }

  const handleTeamFilterChange = (teamId: string) => {
    setFilters((prevFilters) => {
      const currentTeams = prevFilters.teams || []
      const newTeams = currentTeams.includes(teamId)
        ? currentTeams.filter((t) => t !== teamId)
        : [...currentTeams, teamId]

      return {
        ...prevFilters,
        teams: newTeams.length > 0 ? newTeams : undefined,
      }
    })
  }

  const handleSemesterFilterChange = (field: 'min' | 'max', value: string) => {
    setFilters((prevFilters) => {
      const numValue = value === '' ? undefined : parseInt(value, 10)
      const newSemester = {
        ...prevFilters.semester,
        [field]: numValue,
      }

      // Remove empty semester filter
      if (!newSemester.min && !newSemester.max) {
        const newFilters = { ...prevFilters }
        delete newFilters.semester
        return newFilters
      }

      return {
        ...prevFilters,
        semester: newSemester,
      }
    })
  }

  const resetFilters = () => {
    setFilters({})
  }

  const hasActiveFilters =
    (filters.genders && filters.genders.length > 0) ||
    (filters.semester && (filters.semester.min || filters.semester.max)) ||
    (filters.teams && filters.teams.length > 0)

  const getActiveFilterCount = () => {
    let count = 0
    if (filters.genders && filters.genders.length > 0) count += filters.genders.length
    if (filters.semester && (filters.semester.min || filters.semester.max)) count++
    if (filters.teams && filters.teams.length > 0) count += filters.teams.length
    return count
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='outline' className='justify-start'>
          <Filter className='mr-2 h-4 w-4' />
          Filter
          {hasActiveFilters && (
            <span className='ml-1 rounded-full bg-primary text-primary-foreground text-xs px-1 min-w-[16px] h-4 flex items-center justify-center'>
              {getActiveFilterCount()}
            </span>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className='w-56'>
        <DropdownMenuLabel>
          <div className='text-sm text-muted-foreground'>
            Filters will be applied to all diagrams.
          </div>
        </DropdownMenuLabel>

        <DropdownMenuLabel>Gender</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {Object.values(Gender).map((gender) => {
          const genderLabel = getGenderString(gender)
          return (
            <DropdownMenuCheckboxItem
              key={gender}
              checked={filters.genders?.includes(gender) || false}
              onClick={(e) => {
                e.preventDefault()
                handleGenderFilterChange(gender)
              }}
            >
              {genderLabel}
            </DropdownMenuCheckboxItem>
          )
        })}

        <DropdownMenuLabel>Teams</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {teams.map((team) => (
          <DropdownMenuCheckboxItem
            key={team.id}
            checked={filters.teams?.includes(team.id) || false}
            onClick={(e) => {
              e.preventDefault()
              handleTeamFilterChange(team.id)
            }}
          >
            {team.name}
          </DropdownMenuCheckboxItem>
        ))}

        <DropdownMenuLabel>Semester of Study</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <div className='p-2 space-y-2'>
          <div className='flex items-center gap-2'>
            <Input
              type='number'
              placeholder='Min'
              value={filters.semester?.min?.toString() ?? ''}
              onChange={(e) => handleSemesterFilterChange('min', e.target.value)}
              className='w-full'
            />
            <Input
              type='number'
              placeholder='Max'
              value={filters.semester?.max?.toString() ?? ''}
              onChange={(e) => handleSemesterFilterChange('max', e.target.value)}
              className='w-full'
            />
          </div>
        </div>

        <DropdownMenuSeparator />
        <div className='p-2'>
          <Button
            variant='outline'
            size='sm'
            className='w-full justify-center'
            onClick={(e) => {
              e.preventDefault()
              resetFilters()
            }}
            disabled={!hasActiveFilters}
          >
            Clear Filters
          </Button>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
