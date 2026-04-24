import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Layers2, Loader2, RefreshCw } from 'lucide-react'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Separator,
} from '@tumaet/prompt-ui-components'
import type { CoursePhaseWithMetaData, UpdateCoursePhase } from '@tumaet/prompt-shared-state'

import { updateCoursePhase } from '@/network/mutations/updateCoursePhase'
import { setAllocationProfile } from '../../../network/mutations/setAllocationProfile'
import {
  TEAM_ALLOCATION_PROFILE_1000_PLUS,
  TEAM_ALLOCATION_PROFILE_STANDARD,
  type TeamAllocationProfile,
} from '../../../interfaces/companyImport'

interface AllocationProfileSettingsProps {
  coursePhase: CoursePhaseWithMetaData
}

const profileOptions: Array<{
  value: TeamAllocationProfile
  title: string
  description: string
  strategy: string
}> = [
  {
    value: TEAM_ALLOCATION_PROFILE_STANDARD,
    title: 'Standard Team Allocation',
    description:
      'Students rank the teams configured below. TEASE uses the existing preference flow.',
    strategy: 'legacy_preference_lp',
  },
  {
    value: TEAM_ALLOCATION_PROFILE_1000_PLUS,
    title: '1000+ Project Allocation',
    description:
      'Students rank fields of business. Imported companies are synchronized as hidden TEASE projects.',
    strategy: TEAM_ALLOCATION_PROFILE_1000_PLUS,
  },
]

export const getTeamAllocationProfile = (
  coursePhase: CoursePhaseWithMetaData | undefined,
): TeamAllocationProfile => {
  const storedProfile = coursePhase?.restrictedData?.teamAllocationProfile
  return storedProfile === TEAM_ALLOCATION_PROFILE_1000_PLUS
    ? TEAM_ALLOCATION_PROFILE_1000_PLUS
    : TEAM_ALLOCATION_PROFILE_STANDARD
}

export const AllocationProfileSettings = ({ coursePhase }: AllocationProfileSettingsProps) => {
  const queryClient = useQueryClient()
  const selectedProfile = getTeamAllocationProfile(coursePhase)
  const [draftProfile, setDraftProfile] = useState<TeamAllocationProfile>(selectedProfile)

  useEffect(() => {
    setDraftProfile(selectedProfile)
  }, [selectedProfile])

  const updateProfileMutation = useMutation({
    mutationFn: async ({
      coursePhaseUpdate,
      profile,
    }: {
      coursePhaseUpdate: UpdateCoursePhase
      profile: TeamAllocationProfile
    }) => {
      await updateCoursePhase(coursePhaseUpdate)
      await setAllocationProfile(coursePhase.id, profile).catch((error) => {
        console.error(
          'Failed to synchronize allocation profile with Team Allocation service.',
          error,
        )
      })
    },
    onSuccess: (_data, variables) => {
      queryClient.setQueryData<CoursePhaseWithMetaData>(
        ['course_phase', coursePhase.id],
        (currentCoursePhase) =>
          currentCoursePhase
            ? {
                ...currentCoursePhase,
                restrictedData:
                  variables.coursePhaseUpdate.restrictedData ?? currentCoursePhase.restrictedData,
                studentReadableData:
                  variables.coursePhaseUpdate.studentReadableData ??
                  currentCoursePhase.studentReadableData,
              }
            : currentCoursePhase,
      )
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['course_phase', coursePhase.id] })
      void queryClient.invalidateQueries({ queryKey: ['team_allocation_config', coursePhase.id] })
      void queryClient.invalidateQueries({ queryKey: ['team_allocation_team', coursePhase.id] })
    },
  })

  const updateProfile = () => {
    const option = profileOptions.find((item) => item.value === draftProfile)
    if (!option || draftProfile === selectedProfile) return

    updateProfileMutation.mutate({
      profile: draftProfile,
      coursePhaseUpdate: {
        id: coursePhase.id,
        restrictedData: {
          ...coursePhase.restrictedData,
          teamAllocationProfile: draftProfile,
          teaseStrategy: option.strategy,
        },
        studentReadableData: coursePhase.studentReadableData,
      },
    })
  }

  return (
    <Card className='w-full shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-start justify-between gap-4'>
          <div>
            <CardTitle className='text-2xl font-bold flex items-center gap-2'>
              <Layers2 className='h-5 w-5' />
              Allocation Profile
            </CardTitle>
            <CardDescription className='mt-1.5'>
              Select how this Team Allocation phase should collect preferences and prepare TEASE.
            </CardDescription>
          </div>
          <Badge variant='outline'>
            {selectedProfile === TEAM_ALLOCATION_PROFILE_1000_PLUS ? '1000+ Project' : 'Standard'}
          </Badge>
        </div>
      </CardHeader>

      <Separator />

      <CardContent className='grid gap-3 pt-6 md:grid-cols-2'>
        {profileOptions.map((option) => {
          const isSelected = option.value === draftProfile
          const isActive = option.value === selectedProfile
          return (
            <button
              key={option.value}
              type='button'
              onClick={() => setDraftProfile(option.value)}
              disabled={updateProfileMutation.isPending}
              className={`rounded-md border p-4 text-left transition-colors ${
                isSelected ? 'border-primary bg-primary/5' : 'hover:bg-accent/5'
              }`}
            >
              <div className='flex items-center justify-between gap-2'>
                <span className='font-medium'>{option.title}</span>
                {isActive && <Badge>Active</Badge>}
              </div>
              <p className='mt-2 text-sm text-muted-foreground'>{option.description}</p>
              <p className='mt-3 text-xs text-muted-foreground'>
                TEASE strategy: {option.strategy}
              </p>
            </button>
          )
        })}
        <Button
          type='button'
          className='gap-2 justify-self-start md:col-span-2'
          disabled={draftProfile === selectedProfile || updateProfileMutation.isPending}
          onClick={updateProfile}
        >
          {updateProfileMutation.isPending ? (
            <Loader2 className='h-4 w-4 animate-spin' />
          ) : (
            <RefreshCw className='h-4 w-4' />
          )}
          Update Profile
        </Button>
      </CardContent>
    </Card>
  )
}
