import { Card, CardContent, CardHeader, CardTitle } from '@tumaet/prompt-ui-components'
import { ChevronRight } from 'lucide-react'
import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'

import type { AssessmentType } from '../../../interfaces/assessmentType'
import { AssessmentStatusBadge, DeadlineBadge, TeamBadge } from '../../components/badges'

export interface EvaluationTarget {
  id: string
  name: string
  navigationPath: string
  completed: boolean
  evaluationCount: number
}

interface EvaluationSectionProps {
  title: string
  icon: ReactNode
  assessmentType: AssessmentType
  deadline?: Date | string
  teamName?: string
  competencyCount: number
  targets: EvaluationTarget[]
}

export const EvaluationSection = ({
  title,
  icon,
  assessmentType,
  deadline,
  teamName,
  competencyCount,
  targets,
}: EvaluationSectionProps) => {
  const navigate = useNavigate()

  const completedCount = targets.filter((target) => target.completed).length
  const allCompleted = completedCount === targets.length

  if (targets.length === 0) {
    return null
  }

  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 px-4 py-3'>
        <div className='flex items-center gap-2'>
          {icon}
          <CardTitle className='text-base'>{title}</CardTitle>
          {teamName && <TeamBadge teamName={teamName} />}
        </div>
        <div className='flex items-center gap-3'>
          {targets.length > 1 && (
            <span className='text-sm text-muted-foreground'>
              {completedCount}/{targets.length}
              <span className='hidden sm:inline'> completed</span>
            </span>
          )}
          {allCompleted ? (
            <AssessmentStatusBadge remainingAssessments={0} isFinalized />
          ) : (
            deadline && <DeadlineBadge deadline={deadline} type={assessmentType} />
          )}
        </div>
      </CardHeader>
      <CardContent className='border-t p-0'>
        <div className='divide-y'>
          {targets.map((target) => (
            <button
              key={target.id}
              type='button'
              onClick={() => navigate(target.navigationPath)}
              className='flex w-full items-center justify-between px-4 py-3 text-left transition-colors hover:bg-muted/50'
            >
              <span className='text-sm font-medium'>{target.name}</span>
              <span className='flex items-center gap-2'>
                <AssessmentStatusBadge
                  remainingAssessments={Math.max(0, competencyCount - target.evaluationCount)}
                  isFinalized={target.completed}
                  assessmentType={assessmentType}
                />
                <ChevronRight className='h-4 w-4 text-muted-foreground' />
              </span>
            </button>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
