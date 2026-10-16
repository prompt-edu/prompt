import type { ApplicationProfilePictureConfig } from '@core/interfaces/application/openApplicationDetails'
import { ProfilePictureInput, useOwnProfilePicture } from '@tumaet/prompt-ui-components'
import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import type { ProfilePictureSectionRef } from '../../utils/ProfilePictureSectionRef'

interface ApplicationProfilePictureSectionProps {
  config: ApplicationProfilePictureConfig
  isInstructorView?: boolean
}

const MISSING_PICTURE_ERROR = 'Please add a profile picture.'

/** Asks a logged-in applicant for their profile picture as part of the application form. */
export const ApplicationProfilePictureSection = forwardRef<
  ProfilePictureSectionRef,
  ApplicationProfilePictureSectionProps
>(({ config, isInstructorView = false }, ref) => {
  const { data: ownPicture, refetch } = useOwnProfilePicture()
  const [error, setError] = useState<string | undefined>(undefined)
  const required = config.requirement === 'required'

  // Adding a picture resolves the error right away, not only on the next submit
  useEffect(() => {
    if (ownPicture) setError(undefined)
  }, [ownPicture])

  useImperativeHandle(ref, () => ({
    validate: async () => {
      if (!required || isInstructorView) return true
      // The picture may have changed in another tab, so the current state decides
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
      readOnly={isInstructorView}
      error={error}
      notice={
        config.hiddenUntilAccepted
          ? 'Your picture is not shown to the people reviewing your application. It becomes visible to the course team once you are accepted.'
          : 'Your picture is visible to the course team while they review your application.'
      }
    />
  )
})

ApplicationProfilePictureSection.displayName = 'ApplicationProfilePictureSection'
