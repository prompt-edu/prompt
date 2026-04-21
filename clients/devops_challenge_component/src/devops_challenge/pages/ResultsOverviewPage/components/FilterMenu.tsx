import { Filter } from 'lucide-react'
import { ColumnFiltersState } from '@tanstack/react-table'
import { PassStatus } from '@tumaet/prompt-shared-state'
import { getChallengeStatusBadge } from '../utils/getChallengeStatusBadge'
import { ChallengeStatus } from '../interfaces/challengeStatus'
import { getStatusBadge } from '@/utils/getStatusBadge'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  Button,
} from '@tumaet/prompt-ui-components'

interface ColumnFiltersProps {
  columnFilters: ColumnFiltersState
  setColumnFilters: React.Dispatch<React.SetStateAction<ColumnFiltersState>>
}

export const FilterMenu = ({ columnFilters, setColumnFilters }: ColumnFiltersProps) => {
  const isSelected = <T extends string>(id: string, value: T) => {
    return columnFilters.some(
      (filter) => filter.id === id && Array.isArray(filter.value) && filter.value.includes(value),
    )
  }

  const handleFilterChange = <T extends string>(id: string, value: T) => {
    setColumnFilters((prevFilters) => {
      const existingFilter = prevFilters.find((filter) => filter.id === id)

      if (existingFilter && Array.isArray(existingFilter.value)) {
        const updatedValue = existingFilter.value.includes(value)
          ? existingFilter.value.filter((v) => v !== value) // Remove if exists
          : [...existingFilter.value, value] // Add if not exists

        return updatedValue.length > 0
          ? prevFilters.map((filter) =>
              filter.id === id ? { ...filter, value: updatedValue } : filter,
            )
          : prevFilters.filter((filter) => filter.id !== id) // Remove filter if no values
      } else {
        return [...prevFilters, { id, value: [value] }]
      }
    })
  }

  const renderFilterItems = <T extends string>(
    id: string,
    items: Record<string, T>,
    getDisplay: (value: T) => React.ReactNode,
  ) => {
    return Object.values(items).map((value) => {
      return (
        <DropdownMenuCheckboxItem
          key={value}
          checked={isSelected(id, value)}
          onClick={(e) => {
            e.preventDefault()
            handleFilterChange(id, value)
          }}
        >
          {getDisplay(value)}
        </DropdownMenuCheckboxItem>
      )
    })
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='outline' className='ml-auto'>
          <Filter className='h-4 w-4' />
          Filters
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-56'>
        <DropdownMenuLabel>Assessment</DropdownMenuLabel>
        {renderFilterItems('passStatus', PassStatus, getStatusBadge)}

        <DropdownMenuSeparator />

        <DropdownMenuLabel>Challenge Status</DropdownMenuLabel>
        {renderFilterItems('challengeStatus', ChallengeStatus, getChallengeStatusBadge)}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
