export const ACCEPTED_PICTURE_TYPES = ['image/jpeg', 'image/png', 'image/webp']

// Only the chosen file is limited; the cropped 512 px JPEG that gets uploaded is far smaller.
export const MAX_PICTURE_FILE_BYTES = 5 * 1024 * 1024

interface PictureFile {
  type: string
  size: number
}

/** Returns why the file cannot be used as a profile picture, or null if it can. */
export const validatePictureFile = (file: PictureFile): string | null => {
  if (!ACCEPTED_PICTURE_TYPES.includes(file.type)) {
    return 'Please choose a JPEG, PNG, or WebP image.'
  }
  if (file.size > MAX_PICTURE_FILE_BYTES) {
    return 'Please choose an image smaller than 5 MB.'
  }
  return null
}
