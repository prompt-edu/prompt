import { TeamAllocationTimeframeSettings } from './components/TeamAllocationTimeframeSettings'
import { Timeframe } from '../../interfaces/timeframe'
import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { getTimeframe } from '../../network/queries/getSurveyTimeframe'
import { useParams } from 'react-router-dom'
import { getConfig } from '../../network/queries/getConfig'
import { MissingSettings, MissingSettingsItem } from '@tumaet/prompt-ui-components'
import { useEffect, useState } from 'react'

export const SettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const {
    data: timeframe,
    isPending: isTimeframePending,
    isError: isTimeframeError,
    refetch: refetchTimeframe,
  } = useQuery<Timeframe>({
    queryKey: ['timeframe', phaseId],
    queryFn: () => getTimeframe(phaseId ?? ''),
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

  const isPending = isTimeframePending || isConfigPending
  const isError = isTimeframeError || isConfigError
  const refetch = () => {
    refetchTimeframe()
    refetchConfig()
  }

  const configToReadableTitle = (key: string): string => {
    switch (key) {
      case 'surveyTimeframe':
        return 'Survey Timeframe'
      default:
        return key.charAt(0).toUpperCase() + key.slice(1)
    }
  }

  const configToReadableDescription = (key: string): string => {
    switch (key) {
      case 'surveyTimeframe':
        return 'survey timeframe'
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

  if (isPending) {
    return (
      <div className='flex items-center justify-center h-full'>
        <Loader2 className='animate-spin' />
      </div>
    )
  }

  if (isError) {
    return <ErrorPage onRetry={refetch} />
  }

  return (
    <div>
      <ManagementPageHeader>Settings</ManagementPageHeader>
      <MissingSettings elements={missingConfigs} />
      <TeamAllocationTimeframeSettings teamAllocationTimeframe={timeframe} />
    </div>
  )
}
