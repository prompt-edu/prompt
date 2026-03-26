import React from 'react'
import { Badge } from '@tumaet/prompt-ui-components'

interface PeerEvaluationCompletionBadgeProps {
  completed: number
  total: number
}

export const PeerEvaluationCompletionBadge: React.FC<PeerEvaluationCompletionBadgeProps> = ({
  completed,
  total,
}) => {
  const statusText = `${completed}/${total}`

  // Determine badge variant based on completion ratio
  const getVariantClass = () => {
    if (total === 0) {
      return (
        'bg-gray-100 dark:bg-gray-100 text-gray-800 dark:text-gray-800 ' +
        'border-gray-200 dark:border-gray-200 hover:bg-gray-200 dark:hover:bg-gray-200'
      )
    }

    const ratio = completed / total

    if (ratio === 1) {
      // Completed - Green
      return (
        'bg-green-100 dark:bg-green-100 text-green-800 dark:text-green-800 ' +
        'border-green-200 dark:border-green-200 hover:bg-green-200 dark:hover:bg-green-200'
      )
    } else if (ratio === 0) {
      // Not started - Red
      return (
        'bg-red-100 dark:bg-red-100 text-red-800 dark:text-red-800 ' +
        'border-red-200 dark:border-red-200 hover:bg-red-200 dark:hover:bg-red-200'
      )
    } else {
      // Partially completed - Orange
      return (
        'bg-orange-100 dark:bg-orange-100 text-orange-800 dark:text-orange-800 ' +
        'border-orange-200 dark:border-orange-200 hover:bg-orange-200 dark:hover:bg-orange-200'
      )
    }
  }

  return <Badge className={getVariantClass()}>{statusText}</Badge>
}
