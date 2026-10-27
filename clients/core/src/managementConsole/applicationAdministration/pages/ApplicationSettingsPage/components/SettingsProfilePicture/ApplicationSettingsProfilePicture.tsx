import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  Button,
  Label,
  type ProfilePictureRequirement,
  ProfilePictureRequirementSetting,
  Separator,
  Switch,
  Textarea,
} from '@tumaet/prompt-ui-components'
import { AlertTriangle, Info, Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ApplicationMetaData } from '../../../../interfaces/applicationMetaData'
import {
  type ApplicationProfilePictureSettings,
  parseProfilePictureSettings,
} from '../../../../utils/profilePictureSettings'
import { ShowProfilePicturesConfirmation } from './ShowProfilePicturesConfirmation'

interface ApplicationSettingsProfilePictureProps {
  initialData: ApplicationMetaData
}

const DESCRIPTION =
  'Ask applicants for a profile picture, so the course team can recognize them once they are accepted.'

const OPTION_DESCRIPTIONS: Record<ProfilePictureRequirement, string> = {
  off: 'Applicants are not asked for a picture.',
  optional: 'Applicants are asked for a picture but can apply without one.',
  required: 'Applicants must add a picture to apply.',
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
  const [isShowConfirmationOpen, setShowConfirmationOpen] = useState(false)

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

  // Only switching off the saved hiding reveals pictures applicants were told are hidden
  const handleHideUntilAcceptedChange = (hideUntilAccepted: boolean) => {
    if (!hideUntilAccepted && savedSettings.hideUntilAccepted) {
      setShowConfirmationOpen(true)
    } else {
      update({ hideUntilAccepted })
    }
  }

  const asksForPicture = settings.requirement !== 'off'

  return (
    <div data-testid='application-profile-picture-settings'>
      <ProfilePictureRequirementSetting
        value={settings.requirement}
        onChange={(requirement: ProfilePictureRequirement) => update({ requirement })}
        description={DESCRIPTION}
        optionDescriptions={OPTION_DESCRIPTIONS}
      >
        {asksForPicture &&
          (initialData.universityLoginAvailable ? (
            <Alert>
              <AlertDescription className='flex items-start gap-3'>
                <Info className='mt-0.5 h-4 w-4 shrink-0' />
                Only applicants who log in with their university account are asked.
              </AlertDescription>
            </Alert>
          ) : (
            <Alert>
              <AlertDescription className='flex items-start gap-3'>
                <AlertTriangle className='mt-0.5 h-4 w-4 shrink-0' />
                Applicants apply without logging in, so the form cannot ask them for a picture. To
                ask for one, turn on Require Student Login in the general settings above.
              </AlertDescription>
            </Alert>
          ))}

        {asksForPicture && (
          <div className='space-y-2'>
            <div className='space-y-1'>
              <Label htmlFor='profile-picture-explanation'>
                Explanation for applicants (optional)
              </Label>
              <p className='text-sm text-muted-foreground'>
                Shown next to the upload in the application form.
              </p>
            </div>
            <Textarea
              id='profile-picture-explanation'
              rows={2}
              value={settings.explanation}
              placeholder={EXPLANATION_PLACEHOLDER}
              onChange={(event) => update({ explanation: event.target.value })}
              maxLength={500}
            />
          </div>
        )}

        <Separator />

        <div className='space-y-3'>
          <h4 className='font-medium'>Application review</h4>
          <div className='flex items-center justify-between gap-4'>
            <div className='space-y-1'>
              <Label htmlFor='hide-profile-pictures'>Hide pictures until acceptance</Label>
              <p className='text-sm text-muted-foreground'>
                Reviewers see no applicant pictures until the applicant is accepted, so pictures
                cannot influence the decision.
              </p>
            </div>
            <Switch
              id='hide-profile-pictures'
              checked={settings.hideUntilAccepted}
              onCheckedChange={handleHideUntilAcceptedChange}
            />
          </div>
        </div>

        <ShowProfilePicturesConfirmation
          isOpen={isShowConfirmationOpen}
          setOpen={setShowConfirmationOpen}
          onConfirm={() => update({ hideUntilAccepted: false })}
        />

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
