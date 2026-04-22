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
import { SurveyLinkCard } from './components/SurveyLinkCard'
import { getConfig } from '../../network/queries/getConfig'
import { MissingSettings, MissingSettingsItem } from '@/components/MissingSettings'
import { useEffect, useState } from 'react'

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

  const [missingConfigs, setMissingConfigs] = useState<MissingSettingsItem[]>([])

  const isPending = isSkillsPending || isTeamsPending || isSurveyTimeframePending || isConfigPending
  const isError = isSkillsError || isTeamsError || isSurveyTimeframeError || isConfigError
  const refetch = () => {
    refetchSkills()
    refetchTeams()
    refetchTimeframe()
    refetchConfig()
  }

  const configToReadableTitle = (key: string): string => {
    switch (key) {
      case 'surveyTimeframe':
        return 'Survey Timeframe'
      case 'teams':
        return 'Teams'
      case 'skills':
        return 'Skills'
      default:
        return key.charAt(0).toUpperCase() + key.slice(1)
    }
  }

  const configToReadableDescription = (key: string): string => {
    switch (key) {
      case 'surveyTimeframe':
        return 'survey timeframe'
      case 'teams':
        return 'teams'
      case 'skills':
        return 'skills'
      default:
        return key.slice(1)
    }
  }

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
  }, [fetchedConfig])

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
      <SurveyLinkCard />
      <MissingSettings elements={missingConfigs} />
      {/* 1. Set the survey timeframe, skills and teams for this phase. */}
      <SurveyTimeframeSettings surveyTimeframe={fetchedSurveyTimeframe} />
      {/* 2. Set up the teams */}
      <TeamSettings teams={fetchedTeams} />
      {/* 3. Set up the skills */}
      <SkillSettings skills={fetchedSkills} />
    </>
  )
}
