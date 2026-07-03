import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorPage,
  ManagementPageHeader,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import dayjs from 'dayjs'
import { BarChart3, Clock, Star } from 'lucide-react'
import { getSurveyStatistics } from '../../network/queries/getSurveyStatistics'
import { getSurveyTimeframe } from '../../network/queries/getSurveyTimeframe'
import { SurveyStatistics } from '../../interfaces/surveyStatistics'
import { SurveyTimeframe } from '../../interfaces/timeframe'
import { TeamPopularityChart } from './components/TeamPopularityChart'
import { SkillDistributionChart } from './components/SkillDistributionChart'
import { SurveySummaryCards } from './components/SurveySummaryCards'

const StatisticsSkeleton = () => (
  <>
    <div className='grid gap-4 md:grid-cols-3'>
      <Skeleton className='h-[120px]' />
      <Skeleton className='h-[120px]' />
      <Skeleton className='h-[120px]' />
    </div>
    <Skeleton className='h-[420px]' />
    <Skeleton className='h-[420px]' />
  </>
)

export const SurveyStatisticsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: statistics,
    isPending,
    isError,
  } = useQuery<SurveyStatistics>({
    queryKey: ['team_allocation_survey_statistics', phaseId],
    queryFn: () => getSurveyStatistics(phaseId!),
    enabled: !!phaseId,
  })

  const { data: participations } = useGetCoursePhaseParticipants()

  const { data: timeframe } = useQuery<SurveyTimeframe>({
    queryKey: ['team_allocation_survey_timeframe', phaseId],
    queryFn: () => getSurveyTimeframe(phaseId!),
    enabled: !!phaseId,
  })

  if (isError) {
    return <ErrorPage />
  }

  const surveyStillOpen =
    !!timeframe?.timeframeSet && dayjs(timeframe.surveyDeadline).isAfter(dayjs())

  const hasTeamData = (statistics?.teamPopularityStatistics ?? []).length > 0
  const hasSkillData = (statistics?.skillDistributionStatistics ?? []).length > 0

  return (
    <div className='flex flex-col gap-6'>
      <ManagementPageHeader>Survey Statistics</ManagementPageHeader>

      {isPending ? (
        <StatisticsSkeleton />
      ) : (
        <>
          {surveyStillOpen && (
            <Alert>
              <Clock className='h-4 w-4' />
              <AlertTitle>Survey still open</AlertTitle>
              <AlertDescription>
                Students can respond until{' '}
                {dayjs(timeframe.surveyDeadline).format('MMM D, YYYY [at] h:mm A')} — these results
                are live and will change as more responses come in.
              </AlertDescription>
            </Alert>
          )}

          <SurveySummaryCards
            statistics={statistics}
            participantCount={participations?.participations?.length}
          />

          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2'>
                <Star className='h-5 w-5' />
                Team Popularity
              </CardTitle>
              <CardDescription>
                Number of students who ranked each team as one of their top choices — taller bars
                mean higher demand. Hover or tap a bar for the full rank breakdown.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {hasTeamData ? (
                <TeamPopularityChart data={statistics.teamPopularityStatistics} />
              ) : (
                <div className='h-[300px] flex items-center justify-center text-muted-foreground'>
                  No team preference data available yet.
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2'>
                <BarChart3 className='h-5 w-5' />
                Skill Distribution
              </CardTitle>
              <CardDescription>
                Self-reported proficiency levels across all survey respondents per skill
              </CardDescription>
            </CardHeader>
            <CardContent>
              {hasSkillData ? (
                <SkillDistributionChart data={statistics.skillDistributionStatistics} />
              ) : (
                <div className='h-[300px] flex items-center justify-center text-muted-foreground'>
                  No skill data available yet.
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
