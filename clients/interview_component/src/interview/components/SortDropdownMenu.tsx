import {
  Button,
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@tumaet/prompt-ui-components'
import { ArrowUpDown } from 'lucide-react'
import { SORT_OPTIONS, type SortOption } from '../utils/sortOptions'

interface SortDropdownMenuProps {
  sortBy: SortOption
  setSortBy: (value: SortOption) => void
}

export const SortDropdownMenu = ({ sortBy, setSortBy }: SortDropdownMenuProps) => {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='outline'>
          <ArrowUpDown className='h-4 w-4' />
          Sort
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        {SORT_OPTIONS.map((option) => (
          <DropdownMenuCheckboxItem
            key={option.id}
            onClick={() => setSortBy(option.id)}
            checked={sortBy === option.id}
          >
            {option.label}
          </DropdownMenuCheckboxItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
