import { createAuthenticatedAxiosInstance, env } from '@tumaet/prompt-shared-state'
import { type AxiosResponse, isAxiosError } from 'axios'

export const WRITE_TIMEOUT_MS = 10_000

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const

export const assessmentAxiosInstance = createAuthenticatedAxiosInstance(env.ASSESSMENT_HOST)

const describeError = (error: unknown): { status?: number; message: string } => {
  if (isAxiosError(error)) {
    return { status: error.response?.status, message: error.message }
  }
  if (error instanceof Error) {
    return { message: error.message }
  }
  return { message: 'Unknown error' }
}

assessmentAxiosInstance.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    console.error('Assessment request failed', describeError(error))
    return Promise.reject(error)
  },
)

export const coursePhasePath = (coursePhaseID: string): string =>
  `assessment/api/course_phase/${coursePhaseID}`

interface WriteOptions {
  timeoutMs?: number
}

const writeConfig = ({ timeoutMs }: WriteOptions = {}) => ({
  headers: JSON_HEADERS,
  ...(timeoutMs === undefined ? {} : { timeout: timeoutMs }),
})

export const assessmentRequest = {
  get: async <T>(url: string, params?: Record<string, string>): Promise<T> =>
    (await assessmentAxiosInstance.get<T>(url, params ? { params } : undefined)).data,

  getResponse: <T>(url: string): Promise<AxiosResponse<T>> => assessmentAxiosInstance.get<T>(url),

  post: async <T = void>(url: string, body?: unknown, options?: WriteOptions): Promise<T> =>
    (await assessmentAxiosInstance.post<T>(url, body, writeConfig(options))).data,

  put: async <T = void>(url: string, body?: unknown, options?: WriteOptions): Promise<T> =>
    (await assessmentAxiosInstance.put<T>(url, body, writeConfig(options))).data,

  del: async (url: string): Promise<void> => {
    await assessmentAxiosInstance.delete(url, writeConfig())
  },
}
