import { useQuery } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import {
  ErrorPage,
  ManagementPageHeader,
  MissingSettings,
  type MissingSettingsItem,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Skill } from '../../interfaces/skill'
import type { SurveyTimeframe } from '../../interfaces/timeframe'
import { getAllSkills } from '../../network/queries/getAllSkills'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { getConfig } from '../../network/queries/getConfig'
import { getSurveyTimeframe } from '../../network/queries/getSurveyTimeframe'
import { SkillSettings } from './components/SkillSettings'
import { SurveyLinkCard } from './components/SurveyLinkCard'
import { SurveyTimeframeSettings } from './components/SurveyTimeframeSettings'
import { TeamSettings } from './components/TeamSettings'

export const SurveySettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const skillsQuery = useQuery<Skill[]>({
    queryKey: ['team_allocation_skill', phaseId],
    queryFn: () => getAllSkills(phaseId ?? ''),
  })

  const teamsQuery = useQuery<Team[]>({
    queryKey: ['team_allocation_team', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })

  const surveyTimeframeQuery = useQuery<SurveyTimeframe>({
    queryKey: ['team_allocation_survey_timeframe', phaseId],
    queryFn: () => getSurveyTimeframe(phaseId ?? ''),
  })

  const configQuery = useQuery<Record<string, boolean>>({
    queryKey: ['team_allocation_config', phaseId],
    queryFn: () => getConfig(phaseId ?? ''),
  })

  const fetchedSkills = skillsQuery.data
  const fetchedTeams = teamsQuery.data
  const fetchedSurveyTimeframe = surveyTimeframeQuery.data
  const fetchedConfig = configQuery.data

  const [missingConfigs, setMissingConfigs] = useState<MissingSettingsItem[]>([])

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

  return (
    <QueryGate queries={[skillsQuery, teamsQuery, surveyTimeframeQuery, configQuery]}>
      {() =>
        !fetchedSurveyTimeframe || !fetchedTeams || !fetchedSkills ? (
          <ErrorPage description='Could not fetch the survey settings' />
        ) : (
          <>
            <ManagementPageHeader>Survey Settings</ManagementPageHeader>
            <MissingSettings elements={missingConfigs} />
            <SurveyLinkCard />
            {/* 1. Set the survey timeframe, skills and teams for this phase. */}
            <SurveyTimeframeSettings surveyTimeframe={fetchedSurveyTimeframe} />
            {/* 2. Set up the teams */}
            <TeamSettings teams={fetchedTeams} />
            {/* 3. Set up the skills */}
            <SkillSettings skills={fetchedSkills} />
          </>
        )
      }
    </QueryGate>
  )
}
