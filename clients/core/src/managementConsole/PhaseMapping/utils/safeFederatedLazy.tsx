import React from 'react'

export function safeFederatedLazyStudentDetail<P>(
  loader: () => Promise<{ StudentDetail?: React.ComponentType<P> }>,
  Fallback: React.ComponentType<P>,
) {
  return React.lazy(async () => {
    try {
      const module = await loader()

      if (!module.StudentDetail) {
        console.warn('[MF] StudentDetail export not found, using fallback')
        return { default: Fallback }
      }

      return { default: module.StudentDetail }
    } catch (err) {
      console.error('[MF] Failed to load remote module, using fallback', err)
      return { default: Fallback }
    }
  })
}
