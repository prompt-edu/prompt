import axios from 'axios'
import type { PresentationApiError } from './interfaces'

export const formatDateTime = (value: string): string =>
  new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))

export const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  const units = ['KB', 'MB', 'GB']
  let value = bytes / 1024
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`
}

export const toDateTimeLocal = (value: string): string => {
  const date = new Date(value)
  const offset = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

export const getApiError = (error: unknown): PresentationApiError => {
  if (!axios.isAxiosError<PresentationApiError>(error)) return {}
  return error.response?.data ?? {}
}

export const getErrorMessage = (error: unknown, fallback: string): string => {
  const apiError = getApiError(error)
  return apiError.message ?? apiError.error ?? fallback
}
