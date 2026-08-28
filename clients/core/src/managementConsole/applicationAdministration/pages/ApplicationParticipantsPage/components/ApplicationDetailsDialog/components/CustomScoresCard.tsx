import type { AdditionalScore } from '@core/managementConsole/applicationAdministration/interfaces/additionalScore/additionalScore'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  cn,
} from '@tumaet/prompt-ui-components'

interface CustomScoresCardProps {
  additionalScores: AdditionalScore[]
  restrictedData: { [key: string]: unknown }
}

const formatScore = (value: unknown): string | null => {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return String(value)
  }
  if (typeof value === 'string' && value.trim() !== '') {
    return value
  }
  return null
}

export const CustomScoresCard = ({ additionalScores, restrictedData }: CustomScoresCardProps) => {
  if (additionalScores.length === 0) {
    return null
  }

  return (
    <Card className='w-full' data-testid='application-custom-scores'>
      <CardHeader>
        <CardTitle>Custom Scores</CardTitle>
        <CardDescription>Scores uploaded for this applicant.</CardDescription>
      </CardHeader>
      <CardContent>
        <dl className='divide-y divide-border'>
          {additionalScores.map((additionalScore) => {
            const score = formatScore(restrictedData[additionalScore.key])

            return (
              <div
                key={additionalScore.key}
                className='flex items-center justify-between gap-4 py-2 first:pt-0 last:pb-0'
              >
                <dt className='text-sm text-muted-foreground'>{additionalScore.name}</dt>
                <dd
                  className={cn(
                    'text-sm',
                    score === null ? 'italic text-muted-foreground' : 'font-medium tabular-nums',
                  )}
                >
                  {score ?? 'No score uploaded'}
                </dd>
              </div>
            )
          })}
        </dl>
      </CardContent>
    </Card>
  )
}
