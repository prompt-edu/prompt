import { Skill } from '../../interfaces/skill'
import { getAllSkills } from '../../network/queries/getAllSkills'
import { useParams } from 'react-router-dom'
import { Team } from '@tumaet/prompt-shared-state'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { Loader2 } from 'lucide-react'
import { ManagementPageHeader, ErrorPage } from '@tumaet/prompt-ui-components'
import { useQuery } from '@tanstack/react-query'
import { getSurveyTimeframe } from '../../network/queries/getSurveyTimeframe'
import { SurveyTimeframe } from '../../interfaces/timeframe'
import { TeamSettings } from './components/TeamSettings'
import { SkillSettings } from './components/SkillSettings'
import { SurveyTimeframeSettings } from './components/SurveyTimeframeSettings'
import {
  AllocationProfileSettings,
  getTeamAllocationProfile,
} from './components/AllocationProfileSettings'
import { CompanyCsvImport } from './components/CompanyCsvImport'
import { FieldAlignmentSettings } from './components/FieldAlignmentSettings'
import { getConfig } from '../../network/queries/getConfig'
import { MissingSettings, MissingSettingsItem } from '@/components/MissingSettings'
import { useCallback, useEffect, useState } from 'react'
import { useGetCoursePhase } from '@/hooks/useGetCoursePhase'
import { TEAM_ALLOCATION_PROFILE_1000_PLUS } from '../../interfaces/companyImport'

export const SurveySettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: fetchedSkills,
    isPending: isSkillsPending,
    isError: isSkillsError,
    refetch: refetchSkills,
  } = useQuery<Skill[]>({
    queryKey: ['team_allocation_skill', phaseId],
    queryFn: () => getAllSkills(phaseId ?? ''),
  })

  const {
    data: fetchedTeams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useQuery<Team[]>({
    queryKey: ['team_allocation_team', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const {
    data: fetchedSurveyTimeframe,
    isPending: isSurveyTimeframePending,
    isError: isSurveyTimeframeError,
    refetch: refetchTimeframe,
  } = useQuery<SurveyTimeframe>({
    queryKey: ['team_allocation_survey_timeframe', phaseId],
    queryFn: () => getSurveyTimeframe(phaseId ?? ''),
  })

  const {
    data: fetchedConfig,
    isPending: isConfigPending,
    isError: isConfigError,
    refetch: refetchConfig,
  } = useQuery<Record<string, boolean>>({
    queryKey: ['team_allocation_config', phaseId],
    queryFn: () => getConfig(phaseId ?? ''),
  })

  const {
    data: fetchedCoursePhase,
    isPending: isCoursePhasePending,
    isError: isCoursePhaseError,
    refetch: refetchCoursePhase,
  } = useGetCoursePhase()

  const [missingConfigs, setMissingConfigs] = useState<MissingSettingsItem[]>([])

  const selectedProfile = getTeamAllocationProfile(fetchedCoursePhase)
  const isProjectWeekProfile = selectedProfile === TEAM_ALLOCATION_PROFILE_1000_PLUS

  const isPending =
    isSkillsPending ||
    isTeamsPending ||
    isSurveyTimeframePending ||
    isConfigPending ||
    isCoursePhasePending
  const isError =
    isSkillsError || isTeamsError || isSurveyTimeframeError || isConfigError || isCoursePhaseError
  const refetch = () => {
    refetchSkills()
    refetchTeams()
    refetchTimeframe()
    refetchConfig()
    refetchCoursePhase()
  }

  const configToReadableTitle = useCallback(
    (key: string): string => {
      switch (key) {
        case 'surveyTimeframe':
          return 'Survey Timeframe'
        case 'teams':
          return isProjectWeekProfile ? 'Hidden Company Projects' : 'Teams'
        case 'skills':
          return 'Skills'
        default:
          return key.charAt(0).toUpperCase() + key.slice(1)
      }
    },
    [isProjectWeekProfile],
  )

  const configToReadableDescription = useCallback(
    (key: string): string => {
      switch (key) {
        case 'surveyTimeframe':
          return 'survey timeframe'
        case 'teams':
          return isProjectWeekProfile ? 'hidden company projects' : 'teams'
        case 'skills':
          return 'skills'
        default:
          return key.slice(1)
      }
    },
    [isProjectWeekProfile],
  )

  useEffect(() => {
    if (!fetchedConfig) {
      setMissingConfigs([])
      return
    }
    const items: MissingSettingsItem[] = Object.entries(fetchedConfig)
      .filter(([, isSet]) => !isSet)
      .map(([key]) => ({
        title: configToReadableTitle(key),
        icon: Loader2,
        description: `The ${configToReadableDescription(key)} configuration is missing.`,
      }))
    setMissingConfigs(items)
  }, [fetchedConfig, configToReadableTitle, configToReadableDescription])

  if (isError) {
    return <ErrorPage onRetry={refetch} />
  }

  if (isPending) {
    return (
      <div className='flex justify-center items-center flex-grow'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <>
      <ManagementPageHeader>Survey Settings</ManagementPageHeader>
      <MissingSettings elements={missingConfigs} />
      <AllocationProfileSettings coursePhase={fetchedCoursePhase} />
      <SurveyTimeframeSettings surveyTimeframe={fetchedSurveyTimeframe} />
      {isProjectWeekProfile ? (
        <>
          <FieldAlignmentSettings coursePhase={fetchedCoursePhase} />
          <CompanyCsvImport phaseId={phaseId ?? ''} coursePhase={fetchedCoursePhase} />
        </>
      ) : (
        <TeamSettings teams={fetchedTeams} />
      )}
      <SkillSettings skills={fetchedSkills} />
    </>
  )
}
