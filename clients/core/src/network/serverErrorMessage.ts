import { isAxiosError } from 'axios'

export const serverErrorMessage = (error: unknown): string =>
  (isAxiosError<{ error?: string }>(error) && error.response?.data?.error) ||
  'Please try again later'
