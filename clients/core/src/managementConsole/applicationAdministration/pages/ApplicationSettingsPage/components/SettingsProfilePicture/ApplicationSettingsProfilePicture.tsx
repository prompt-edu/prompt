import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import {
  Button,
  Label,
  type ProfilePictureRequirement,
  ProfilePictureRequirementSetting,
  Switch,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'
import {
  type ApplicationProfilePictureSettings,
  parseProfilePictureSettings,
} from '../../../../utils/profilePictureSettings'

interface ApplicationSettingsProfilePictureProps {
  initialData: ApplicationMetaData
}

const EXPLANATION_PLACEHOLDER = 'Your picture helps the course team recognize you.'

const fromMetaData = (data: ApplicationMetaData): ApplicationProfilePictureSettings =>
  parseProfilePictureSettings({
    profilePictureRequirement: data.profilePictureRequirement,
    profilePictureExplanation: data.profilePictureExplanation,
    hideProfilePicturesUntilAccepted: data.hideProfilePicturesUntilAccepted,
  })

export function ApplicationSettingsProfilePicture({
  initialData,
}: ApplicationSettingsProfilePictureProps) {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  const [settings, setSettings] = useState(() => fromMetaData(initialData))
  const [savedSettings, setSavedSettings] = useState(() => fromMetaData(initialData))

  // The metadata is parsed asynchronously, so the first render can precede it.
  useEffect(() => {
    setSettings(fromMetaData(initialData))
    setSavedSettings(fromMetaData(initialData))
  }, [initialData])

  const hasChanges =
    settings.requirement !== savedSettings.requirement ||
    settings.explanation.trim() !== savedSettings.explanation.trim() ||
    settings.hideUntilAccepted !== savedSettings.hideUntilAccepted

  const {
    mutate: mutatePhase,
    isPending,
    isError,
  } = useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => coreApi.coursePhases.update(coursePhase),
    onSuccess: () => {
      setSavedSettings(settings)
      coreCache.coursePhaseChanged(queryClient, phaseId)
    },
  })

  const handleSave = () => {
    const explanation = settings.explanation.trim()
    mutatePhase({
      id: phaseId ?? '',
      restrictedData: {
        profilePictureRequirement: settings.requirement,
        // The metadata is merged, so an emptied text stores an explicit null
        profilePictureExplanation: explanation === '' ? null : explanation,
        hideProfilePicturesUntilAccepted: settings.hideUntilAccepted,
      },
    })
  }

  const update = (changes: Partial<ApplicationProfilePictureSettings>) =>
    setSettings((current) => ({ ...current, ...changes }))

  return (
    <div data-testid='application-profile-picture-settings'>
      <ProfilePictureRequirementSetting
        value={settings.requirement}
        onChange={(requirement: ProfilePictureRequirement) => update({ requirement })}
      >
        {settings.requirement !== 'off' && (
          <div className='space-y-2'>
            <Label htmlFor='profile-picture-explanation'>Explanation for applicants</Label>
            <Textarea
              id='profile-picture-explanation'
              value={settings.explanation}
              placeholder={EXPLANATION_PLACEHOLDER}
              onChange={(event) => update({ explanation: event.target.value })}
              maxLength={500}
            />
            <p className='text-sm text-muted-foreground'>
              Shown next to the picture in the application form. Explain why you ask for a picture
              and what it is used for. Applicants without a PROMPT account (external applicants) are
              not asked.
            </p>
          </div>
        )}

        <div className='flex items-start justify-between gap-4'>
          <div className='space-y-1'>
            <Label htmlFor='hide-profile-pictures'>
              Hide profile pictures during the application review
            </Label>
            <p className='text-sm text-muted-foreground'>
              Applicants' profile pictures stay hidden in this phase until they are accepted, so
              pictures cannot influence the admission decision. Hidden pictures are marked in the
              participant list and on the application page. After acceptance, pictures show
              everywhere in PROMPT.
            </p>
          </div>
          <Switch
            id='hide-profile-pictures'
            checked={settings.hideUntilAccepted}
            onCheckedChange={(hideUntilAccepted) => update({ hideUntilAccepted })}
          />
        </div>

        <div className='flex items-center justify-end gap-3'>
          {isError && <p className='text-sm text-destructive'>The settings could not be saved.</p>}
          <Button onClick={handleSave} disabled={!hasChanges || isPending}>
            {isPending && <Loader2 className='h-4 w-4 animate-spin' />}
            Save
          </Button>
        </div>
      </ProfilePictureRequirementSetting>
    </div>
  )
}
