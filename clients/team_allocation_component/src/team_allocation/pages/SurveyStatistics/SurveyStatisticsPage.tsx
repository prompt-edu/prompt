import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorPage,
  ManagementPageHeader,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import { AxiosError } from 'axios'
import { BarChart3, Clock, Star } from 'lucide-react'
import { getSurveyStatistics } from '../../network/queries/getSurveyStatistics'
import { SurveyStatistics } from '../../interfaces/surveyStatistics'
import { TeamPopularityChart } from './components/TeamPopularityChart'
import { SkillDistributionChart } from './components/SkillDistributionChart'
import { SurveySummaryCards } from './components/SurveySummaryCards'

const isSurveyStillOpenError = (error: unknown): boolean => {
  if (error instanceof AxiosError) {
    const message: string = (error.response?.data as { error?: string })?.error ?? ''
    return message.includes('not available')
  }
  return false
}

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
    error,
  } = useQuery<SurveyStatistics>({
    queryKey: ['team_allocation_survey_statistics', phaseId],
    queryFn: () => getSurveyStatistics(phaseId!),
    enabled: !!phaseId,
  })

  const { data: participations } = useGetCoursePhaseParticipants()

  if (isError && isSurveyStillOpenError(error)) {
    return (
      <div className='flex flex-col gap-6'>
        <ManagementPageHeader>Survey Statistics</ManagementPageHeader>
        <Card>
          <CardHeader>
            <CardTitle className='flex items-center gap-2'>
              <Clock className='h-5 w-5' />
              Statistics Not Yet Available
            </CardTitle>
            <CardDescription>
              Survey statistics will be available once the survey deadline has passed.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className='text-sm text-muted-foreground'>
              Please check back after the survey closes to view team popularity and skill
              distribution data.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (isError) {
    return <ErrorPage />
  }

  const hasTeamData = (statistics?.teamPopularityStatistics ?? []).length > 0
  const hasSkillData = (statistics?.skillDistributionStatistics ?? []).length > 0

  return (
    <div className='flex flex-col gap-6'>
      <ManagementPageHeader>Survey Statistics</ManagementPageHeader>

      {isPending ? (
        <StatisticsSkeleton />
      ) : (
        <>
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
                mean higher demand. Hover a bar for the full rank breakdown.
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
