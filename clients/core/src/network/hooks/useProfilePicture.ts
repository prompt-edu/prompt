import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import axios from 'axios'

// Core presigns the picture URL for an hour; refetching before then keeps it valid.
const OWN_PICTURE_STALE_MS = 50 * 60 * 1000

export const useOwnProfilePicture = () => {
  return useQuery({
    queryKey: coreKeys.profilePictures.own(),
    queryFn: () => coreApi.profilePictures.own(),
    staleTime: OWN_PICTURE_STALE_MS,
  })
}

export const useUploadProfilePicture = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (picture: Blob) => {
      const { uploadUrl, storageKey } = await coreApi.profilePictures.presignUpload()
      // The presigned URL points at the object storage, not at core, so it must not carry the
      // Keycloak token the shared axios instance would add
      await axios.put(uploadUrl, picture, { headers: { 'Content-Type': 'image/jpeg' } })
      return coreApi.profilePictures.completeUpload(storageKey)
    },
    onSuccess: () => {
      coreCache.ownProfilePictureChanged(queryClient)
      toast({ title: 'Profile picture saved' })
    },
    onError: () => {
      toast({
        title: 'Failed to save profile picture',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  })
}

export const useDeleteProfilePicture = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => coreApi.profilePictures.removeOwn(),
    onSuccess: () => {
      coreCache.ownProfilePictureChanged(queryClient)
      toast({ title: 'Profile picture removed' })
    },
    onError: () => {
      toast({
        title: 'Failed to remove profile picture',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  })
}
