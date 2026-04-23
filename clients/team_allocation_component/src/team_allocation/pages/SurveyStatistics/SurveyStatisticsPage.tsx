import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorPage,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { BarChart3, Star } from 'lucide-react'
import { getSurveyStatistics } from '../../network/queries/getSurveyStatistics'
import { SurveyStatistics } from '../../interfaces/surveyStatistics'
import { TeamPopularityChart } from './components/TeamPopularityChart'
import { SkillDistributionChart } from './components/SkillDistributionChart'

export const SurveyStatisticsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: statistics,
    isPending,
    isError,
  } = useQuery<SurveyStatistics>({
    queryKey: ['team_allocation_survey_statistics', phaseId],
    queryFn: () => getSurveyStatistics(phaseId ?? ''),
  })

  if (isError) {
    return <ErrorPage />
  }

  const hasTeamData =
    !isPending && (statistics?.teamPopularityStatistics ?? []).length > 0
  const hasSkillData =
    !isPending && (statistics?.skillDistributionStatistics ?? []).length > 0

  return (
    <div className='flex flex-col gap-6'>
      <ManagementPageHeader>Survey Statistics</ManagementPageHeader>

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <Star className='h-5 w-5' />
            Team Popularity
          </CardTitle>
          <CardDescription>
            Average preference rank per team across all survey respondents — lower is more popular
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isPending ? (
            <div className='h-[300px] flex items-center justify-center text-muted-foreground'>
              Loading...
            </div>
          ) : !hasTeamData ? (
            <div className='h-[300px] flex items-center justify-center text-muted-foreground'>
              No team preference data available yet.
            </div>
          ) : (
            <TeamPopularityChart data={statistics!.teamPopularityStatistics} />
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
            Self-reported skill proficiency levels across all survey respondents
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isPending ? (
            <div className='h-[300px] flex items-center justify-center text-muted-foreground'>
              Loading...
            </div>
          ) : !hasSkillData ? (
            <div className='h-[300px] flex items-center justify-center text-muted-foreground'>
              No skill data available yet.
            </div>
          ) : (
            <SkillDistributionChart data={statistics!.skillDistributionStatistics} />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
