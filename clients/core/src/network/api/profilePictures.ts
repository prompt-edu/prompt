import type { PresignedProfilePictureUpload, ProfilePicture } from '@core/interfaces/profilePicture'
import { isAxiosError } from 'axios'
import { API_PREFIX, coreRequest } from '../client'

const NOT_FOUND = 404

const ownPath = `${API_PREFIX}/profile-pictures/me`

// Other people's pictures are read through the batched lookup in @tumaet/prompt-ui-components.
export const profilePictures = {
  // A user without a picture is the normal case, so the 404 answers null instead of failing.
  own: async (): Promise<ProfilePicture | null> => {
    try {
      return await coreRequest.get<ProfilePicture>(ownPath, { quietStatuses: [NOT_FOUND] })
    } catch (error) {
      if (isAxiosError(error) && error.response?.status === NOT_FOUND) {
        return null
      }
      throw error
    }
  },

  presignUpload: (): Promise<PresignedProfilePictureUpload> =>
    coreRequest.post(`${ownPath}/presign`),

  completeUpload: (storageKey: string): Promise<ProfilePicture> =>
    coreRequest.post(`${ownPath}/complete`, { storageKey }),

  removeOwn: (): Promise<void> => coreRequest.del(ownPath),
}
