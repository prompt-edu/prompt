import { Card, CardContent, CardHeader, CardTitle, Progress } from '@tumaet/prompt-ui-components'
import { Star, TrendingDown, Users } from 'lucide-react'
import { SurveyStatistics, TeamPopularityStats } from '../../../interfaces/surveyStatistics'

interface SurveySummaryCardsProps {
  statistics: SurveyStatistics
  participantCount: number | undefined
}

const getFirstChoiceCount = (team: TeamPopularityStats): number =>
  team.preferenceCounts.find((pc) => pc.rank === 1)?.count ?? 0

const formatTeamSummary = (team: TeamPopularityStats): string => {
  const firstChoices = getFirstChoiceCount(team)
  const avgRank = team.avgPreference !== null ? team.avgPreference.toFixed(1) : '—'
  return `${firstChoices} first ${firstChoices === 1 ? 'choice' : 'choices'} · avg. rank ${avgRank}`
}

export const SurveySummaryCards = ({ statistics, participantCount }: SurveySummaryCardsProps) => {
  const teamStats = statistics.teamPopularityStatistics

  const respondentCount = statistics.respondentCount
  const responseRate =
    participantCount && participantCount > 0
      ? Math.min(100, Math.round((respondentCount / participantCount) * 100))
      : null

  const ratedTeams = teamStats
    .filter((team) => team.avgPreference !== null)
    .sort((a, b) => (a.avgPreference ?? 0) - (b.avgPreference ?? 0))
  const mostPopular = ratedTeams[0]
  const leastPopular = ratedTeams.length > 1 ? ratedTeams[ratedTeams.length - 1] : undefined

  return (
    <div className='grid gap-4 md:grid-cols-3'>
      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Survey Responses</CardTitle>
          <Users className='h-4 w-4 text-muted-foreground' />
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold'>
            {respondentCount}
            {participantCount !== undefined && participantCount > 0 && (
              <span className='text-sm font-normal text-muted-foreground'>
                {' '}
                / {participantCount} students
              </span>
            )}
          </div>
          {responseRate !== null ? (
            <>
              <Progress value={responseRate} className='mt-2 h-2' />
              <p className='text-xs text-muted-foreground mt-1'>{responseRate}% response rate</p>
            </>
          ) : (
            <p className='text-xs text-muted-foreground mt-1'>completed the survey</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Most Popular Team</CardTitle>
          <Star className='h-4 w-4 text-muted-foreground' />
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold truncate'>{mostPopular?.teamName ?? '—'}</div>
          <p className='text-xs text-muted-foreground mt-1'>
            {mostPopular ? formatTeamSummary(mostPopular) : 'No responses yet'}
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Least Popular Team</CardTitle>
          <TrendingDown className='h-4 w-4 text-muted-foreground' />
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold truncate'>{leastPopular?.teamName ?? '—'}</div>
          <p className='text-xs text-muted-foreground mt-1'>
            {leastPopular ? formatTeamSummary(leastPopular) : 'No responses yet'}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
