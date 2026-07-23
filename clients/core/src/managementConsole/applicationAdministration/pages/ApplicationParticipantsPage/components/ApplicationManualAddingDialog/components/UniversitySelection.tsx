import { Button } from '@tumaet/prompt-ui-components'
import { School, UserPlus } from 'lucide-react'

interface UniversitySelectionProps {
  setHasUniversityAccount: (hasUniversityAccount: boolean) => void
}

export const UniversitySelection = ({ setHasUniversityAccount }: UniversitySelectionProps) => {
  return (
    <div className='space-y-4'>
      <p className='text-sm text-muted-foreground'>Choose how you want to add a new application:</p>
      <div className='grid gap-4 sm:grid-cols-2'>
        <Button
          variant='outline'
          className='h-auto min-h-24 flex flex-col items-center justify-center whitespace-normal px-2 text-center'
          onClick={() => setHasUniversityAccount(true)}
        >
          <School className='h-6 w-6 mb-2' />
          Add a student with university account
        </Button>
        <Button
          variant='outline'
          className='h-auto min-h-24 flex flex-col items-center justify-center whitespace-normal px-2 text-center'
          onClick={() => setHasUniversityAccount(false)}
        >
          <UserPlus className='h-6 w-6 mb-2' />
          Add an external student
        </Button>
      </div>
    </div>
  )
}
