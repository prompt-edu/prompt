import type { Area } from 'react-easy-crop'

// Large enough for the biggest avatar on a high-density screen, small enough for long tables.
const PICTURE_SIZE_PX = 512
const PICTURE_JPEG_QUALITY = 0.85

const loadImage = (url: string): Promise<HTMLImageElement> =>
  new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('Failed to load the chosen image'))
    image.src = url
  })

/**
 * Crops the chosen area to a square JPEG. Drawing onto a canvas re-encodes the image, which also
 * drops its metadata, such as the GPS location phones store in photos.
 */
export const cropProfilePicture = async (imageUrl: string, area: Area): Promise<Blob> => {
  const image = await loadImage(imageUrl)
  const canvas = document.createElement('canvas')
  canvas.width = PICTURE_SIZE_PX
  canvas.height = PICTURE_SIZE_PX

  const context = canvas.getContext('2d')
  if (!context) {
    throw new Error('Canvas is not supported in this browser')
  }
  // JPEG has no transparency; without a background, transparent PNG areas would turn black
  context.fillStyle = '#ffffff'
  context.fillRect(0, 0, PICTURE_SIZE_PX, PICTURE_SIZE_PX)
  context.drawImage(
    image,
    area.x,
    area.y,
    area.width,
    area.height,
    0,
    0,
    PICTURE_SIZE_PX,
    PICTURE_SIZE_PX,
  )

  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error('Failed to encode the picture'))),
      'image/jpeg',
      PICTURE_JPEG_QUALITY,
    )
  })
}
