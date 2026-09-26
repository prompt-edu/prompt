import {
  useDeleteProfilePicture,
  useOwnProfilePicture,
  useUploadProfilePicture,
} from '@core/network/hooks/useProfilePicture'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  ProfilePicture,
} from '@tumaet/prompt-ui-components'
import { Camera, ImageUp, Loader2, Trash2 } from 'lucide-react'
import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import Cropper, { type Area, type Point } from 'react-easy-crop'
import { cropProfilePicture } from './utils/cropProfilePicture'
import { ACCEPTED_PICTURE_TYPES, validatePictureFile } from './utils/validatePictureFile'

interface ProfilePictureDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const MIN_ZOOM = 1
const MAX_ZOOM = 3

export const ProfilePictureDialog = ({ open, onOpenChange }: ProfilePictureDialogProps) => {
  const { user } = useAuthStore()
  const { data: ownPicture } = useOwnProfilePicture()
  const uploadPicture = useUploadProfilePicture()
  const deletePicture = useDeleteProfilePicture()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [imageUrl, setImageUrl] = useState<string | null>(null)
  const [fileError, setFileError] = useState<string | null>(null)
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(MIN_ZOOM)
  const [croppedArea, setCroppedArea] = useState<Area | null>(null)

  // The object URL holds the whole chosen file in memory until it is revoked
  useEffect(() => {
    return () => {
      if (imageUrl) URL.revokeObjectURL(imageUrl)
    }
  }, [imageUrl])

  const reset = () => {
    setImageUrl(null)
    setFileError(null)
    setCrop({ x: 0, y: 0 })
    setZoom(MIN_ZOOM)
    setCroppedArea(null)
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) reset()
    onOpenChange(nextOpen)
  }

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    // Clearing the input lets the same file be chosen again after a reset
    event.target.value = ''
    if (!file) return

    const error = validatePictureFile(file)
    if (error) {
      setFileError(error)
      return
    }
    reset()
    setImageUrl(URL.createObjectURL(file))
  }

  const handleSave = async () => {
    if (!imageUrl || !croppedArea) return
    try {
      const picture = await cropProfilePicture(imageUrl, croppedArea)
      await uploadPicture.mutateAsync(picture)
      handleOpenChange(false)
    } catch {
      // The mutation reports its own failure; a failed crop leaves the dialog open to retry
    }
  }

  const handleRemove = async () => {
    try {
      await deletePicture.mutateAsync()
      reset()
    } catch {
      // Reported by the mutation's toast
    }
  }

  const isBusy = uploadPicture.isPending || deletePicture.isPending

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Profile picture</DialogTitle>
          <DialogDescription>
            Your picture is shown to everyone logged in to PROMPT, for example in participant lists
            and team overviews. Adding one is optional.
          </DialogDescription>
        </DialogHeader>

        {imageUrl ? (
          <div className='space-y-4'>
            <div className='relative h-72 w-full overflow-hidden rounded-md bg-muted'>
              <Cropper
                image={imageUrl}
                crop={crop}
                zoom={zoom}
                minZoom={MIN_ZOOM}
                maxZoom={MAX_ZOOM}
                aspect={1}
                cropShape='round'
                showGrid={false}
                onCropChange={setCrop}
                onZoomChange={setZoom}
                onCropComplete={(_, areaPixels) => setCroppedArea(areaPixels)}
              />
            </div>
            <label className='flex items-center gap-3 text-sm'>
              <span className='text-muted-foreground'>Zoom</span>
              <input
                type='range'
                min={MIN_ZOOM}
                max={MAX_ZOOM}
                step={0.01}
                value={zoom}
                onChange={(event) => setZoom(Number(event.target.value))}
                className='w-full accent-primary'
              />
            </label>
          </div>
        ) : (
          <div className='flex flex-col items-center gap-4 py-4'>
            <button
              type='button'
              onClick={() => fileInputRef.current?.click()}
              disabled={isBusy}
              aria-label='Choose a new profile picture'
              className='group relative rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2'
            >
              <ProfilePicture
                src={ownPicture?.url}
                firstName={user?.firstName ?? ''}
                lastName={user?.lastName ?? ''}
                size='lg'
              />
              <span className='absolute inset-0 flex items-center justify-center rounded-full bg-black/50 text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100'>
                <Camera className='h-6 w-6' />
              </span>
            </button>
            <p className='text-sm text-muted-foreground'>
              {ownPicture ? 'This is your current picture.' : 'You have no profile picture yet.'}
            </p>
          </div>
        )}

        {fileError && <p className='text-sm text-destructive'>{fileError}</p>}

        <input
          ref={fileInputRef}
          type='file'
          accept={ACCEPTED_PICTURE_TYPES.join(',')}
          className='hidden'
          onChange={handleFileChange}
        />

        <DialogFooter className='gap-2 sm:justify-between'>
          <div className='flex gap-2'>
            <Button
              variant='outline'
              onClick={() => fileInputRef.current?.click()}
              disabled={isBusy}
            >
              <ImageUp className='h-4 w-4' />
              {imageUrl ? 'Choose another' : 'Choose picture'}
            </Button>
            {ownPicture && !imageUrl && (
              <Button variant='outline' onClick={handleRemove} disabled={isBusy}>
                {deletePicture.isPending ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Trash2 className='h-4 w-4' />
                )}
                Remove
              </Button>
            )}
          </div>
          {imageUrl && (
            <Button onClick={handleSave} disabled={isBusy || !croppedArea}>
              {uploadPicture.isPending && <Loader2 className='h-4 w-4 animate-spin' />}
              Save
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
