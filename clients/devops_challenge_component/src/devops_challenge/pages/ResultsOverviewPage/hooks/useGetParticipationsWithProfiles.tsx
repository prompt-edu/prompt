import { useMemo } from 'react'

import { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { DeveloperWithInfo } from '../../../interfaces/DeveloperWithInfo'

export const useGetParticipationsWithProfiles = (
  participants: CoursePhaseParticipationWithStudent[],
  developerProfiles: DeveloperWithInfo[],
) => {
  return useMemo(() => {
    return (
      console.log('participants', participants),
      participants.map((participation) => {
        if (!developerProfiles || developerProfiles.length === 0) {
          return { participation, profile: undefined }
        }

        const profile =
          developerProfiles?.find(
            (devProfile) =>
              devProfile.courseParticipationID === participation.courseParticipationID,
          ) || undefined

        return { participation, profile }
      }) || []
    )
  }, [participants, developerProfiles])
}
