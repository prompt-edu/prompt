import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { getPassedStudentsCount } from '../network/queries/getPassedStudentsCount'
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@tumaet/prompt-ui-components'
import { Loader2, Trophy, Users, AlertCircle } from 'lucide-react'

export const PassCountIndicator = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: passedCount,
    isLoading,
    isError,
    refetch,
  } = useQuery<number>({
    queryKey: ['devOpsPassedStudentsCount', phaseId],
    queryFn: () => getPassedStudentsCount(phaseId ?? ''),
  })

  useEffect(() => {
    // refetch passed students count every 2 minutes
    const interval = setInterval(() => {
      refetch()
    }, 120000)

    return () => clearInterval(interval)
  }, [refetch])

  return (
    <Card className='w-full'>
      <CardHeader className='pb-2'>
        <CardTitle className='text-xl flex items-center justify-between'>
          <div className='flex items-center'>
            <Trophy className='mr-2 h-5 w-5 text-amber-500' />
            <span>Challenge Status</span>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='flex flex-col space-y-1'>
          <span className='text-sm text-muted-foreground'>
            Students who already passed the challenge
          </span>
          <div className='flex items-center gap-2'>
            {isLoading ? (
              <div className='flex items-center gap-2'>
                <Loader2 className='h-4 w-4 animate-spin text-muted-foreground' />
                <span className='text-muted-foreground'>Loading...</span>
              </div>
            ) : isError ? (
              <div className='flex items-center gap-2 text-red-500'>
                <AlertCircle className='h-4 w-4' />
                <span>Error fetching data</span>
              </div>
            ) : (
              <div className='flex items-center gap-2'>
                <Users className='h-5 w-5 text-primary' />
                <div>
                  <span className='text-3xl font-bold'>{passedCount}</span>
                  <span className='ml-2 text-muted-foreground'>students</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </CardContent>
      <CardFooter className='pt-2 pb-3'>
        <div className='flex items-start gap-2 text-xs text-muted-foreground border-t pt-2 w-full'>
          <AlertCircle className='h-3.5 w-3.5 flex-shrink-0 mt-0.5' />
          <p>
            This is just a rough indicator, not live data, and should only be used as guidance.
            Numbers may not reflect the current state and are updated periodically.
          </p>
        </div>
      </CardFooter>
    </Card>
  )
}
