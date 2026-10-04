import type { CourseTemplateStatus } from '@core/interfaces/courseTemplateStatus'
import { coreApi } from '@core/network/api'
import { useCallback, useEffect, useState } from 'react'

export const useCourseTemplate = (courseId: string) => {
  const [templateStatus, setTemplateStatus] = useState<CourseTemplateStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isUpdating, setIsUpdating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchTemplateStatus = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const status = await coreApi.courses.templateStatus(courseId)
      setTemplateStatus(status)
    } catch (err) {
      setError('Failed to fetch template status')
      console.error('Error fetching template status:', err)
    } finally {
      setIsLoading(false)
    }
  }, [courseId])

  const updateTemplateStatus = async (isTemplate: boolean) => {
    try {
      setIsUpdating(true)
      setError(null)
      await coreApi.courses.setTemplateStatus(courseId, { isTemplate })
      setTemplateStatus({ isTemplate })
    } catch (err) {
      setError('Failed to update template status')
      console.error('Error updating template status:', err)
    } finally {
      setIsUpdating(false)
    }
  }

  useEffect(() => {
    if (courseId) {
      fetchTemplateStatus()
    }
  }, [courseId, fetchTemplateStatus])

  return {
    templateStatus,
    isLoading,
    isUpdating,
    error,
    updateTemplateStatus,
    refetch: fetchTemplateStatus,
  }
}
