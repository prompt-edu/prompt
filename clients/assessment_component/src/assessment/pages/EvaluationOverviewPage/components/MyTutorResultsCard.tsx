import { Button, Card, CardContent } from '@tumaet/prompt-ui-components'
import { ArrowRight, GraduationCap } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

import { useGetCoursePhaseConfig } from '../../hooks/useGetCoursePhaseConfig'
import { useGetMyTutorEvaluationResults } from '../../hooks/useGetMyTutorEvaluationResults'
import { getTutorLabel } from '../../hooks/useTutorLabel'

interface MyTutorResultsCardProps {
  resultsLink: string
}

/** Points a tutor to what their students wrote about them; renders nothing for anyone else. */
export const MyTutorResultsCard = ({ resultsLink }: MyTutorResultsCardProps) => {
  const navigate = useNavigate()
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const tutorLabel = getTutorLabel(coursePhaseConfig)
  const { data: results } = useGetMyTutorEvaluationResults()

  if (!results) return null

  return (
    <Card>
      <CardContent className='p-4'>
        <div className='flex items-start gap-3'>
          <GraduationCap className='h-5 w-5 shrink-0 text-purple-600 dark:text-purple-400' />
          <div className='flex-1'>
            <h2 className='mb-1 text-base font-semibold'>
              Your {tutorLabel.title} Evaluation Results
            </h2>
            <p className='text-sm text-muted-foreground'>
              {results.evaluatorCount === 1
                ? '1 student evaluated you.'
                : `${results.evaluatorCount} students evaluated you.`}{' '}
              The results are anonymized: you see averaged scores and written feedback without
              names.
            </p>
            <div className='mt-3'>
              <Button variant='outline' onClick={() => navigate(resultsLink)} className='gap-2'>
                View your results
                <ArrowRight className='h-4 w-4' />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
