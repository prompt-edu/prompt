import type { ApplicationProfilePictureConfig } from '@core/interfaces/application/openApplicationDetails'
import { ProfilePictureInput, useOwnProfilePicture } from '@tumaet/prompt-ui-components'
import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import type { ProfilePictureSectionRef } from '../../utils/ProfilePictureSectionRef'

interface ApplicationProfilePictureSectionProps {
  config: ApplicationProfilePictureConfig
  isInstructorView?: boolean
}

const MISSING_PICTURE_ERROR = 'Please add a profile picture.'

export const ApplicationProfilePictureSection = forwardRef<
  ProfilePictureSectionRef,
  ApplicationProfilePictureSectionProps
>(({ config, isInstructorView = false }, ref) => {
  const { data: ownPicture, refetch } = useOwnProfilePicture()
  const [error, setError] = useState<string | undefined>(undefined)
  const required = config.requirement === 'required'

  useEffect(() => {
    if (ownPicture) setError(undefined)
  }, [ownPicture])

  useImperativeHandle(ref, () => ({
    validate: async () => {
      if (!required || isInstructorView) return true
      const { data } = await refetch()
      const hasPicture = Boolean(data)
      setError(hasPicture ? undefined : MISSING_PICTURE_ERROR)
      return hasPicture
    },
  }))

  return (
    <ProfilePictureInput
      required={required}
      explanation={config.explanation}
      preview={isInstructorView}
      error={error}
      notice={
        config.hiddenUntilAccepted
          ? 'Your picture is hidden in the application review until you are accepted.'
          : 'Reviewers can see your picture.'
      }
    />
  )
})

ApplicationProfilePictureSection.displayName = 'ApplicationProfilePictureSection'
