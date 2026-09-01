import { axiosInstance, notAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { isAxiosError } from 'axios'

/**
 * Core talks to its own server on the same origin, so the path prefix is the only route policy
 * there is to own.
 */
export const API_PREFIX = '/api'

export const NO_CONTENT = 204

/**
 * Every write announces plain JSON. The 41 write modules this replaces were split between
 * `application/json-path+json`, `application/json` and nothing at all, and three reads sent a
 * content type on a request with no body. Core's handlers bind with `BindJSON`, which never reads
 * the header, so the split was noise rather than a contract.
 */
const JSON_HEADERS = { 'Content-Type': 'application/json' } as const

const describeError = (error: unknown): { status?: number; message: string } => {
  if (isAxiosError(error)) {
    return { status: error.response?.status, message: error.message }
  }
  if (error instanceof Error) {
    return { message: error.message }
  }
  return { message: 'Unknown error' }
}

/**
 * Logs once, sanitized, and rethrows the original error, so the `onError` handlers that read
 * `error.response.data.error` keep working.
 *
 * This wraps each call rather than installing an interceptor: `axiosInstance` is a Module
 * Federation singleton shared with every remote, and core has no business logging their requests.
 */
const send = async <T>(
  instance: AxiosInstance,
  description: string,
  config: AxiosRequestConfig,
): Promise<AxiosResponse<T>> => {
  try {
    return await instance.request<T>(config)
  } catch (error) {
    console.error(`${description} request failed`, describeError(error))
    throw error
  }
}

interface ReadOptions {
  params?: Record<string, string | number | boolean>
}

interface WriteOptions extends ReadOptions {
  /** Sent as the request body of a DELETE, which is how core deletes a batch. */
  data?: unknown
}

const requestsThrough = (instance: AxiosInstance, description: string) => ({
  get: async <T>(url: string, options?: ReadOptions): Promise<T> =>
    (await send<T>(instance, description, { method: 'get', url, ...options })).data,

  getResponse: <T>(url: string, options?: ReadOptions): Promise<AxiosResponse<T>> =>
    send<T>(instance, description, { method: 'get', url, ...options }),

  post: async <T = void>(url: string, data?: unknown, options?: WriteOptions): Promise<T> =>
    (
      await send<T>(instance, description, {
        method: 'post',
        url,
        data,
        headers: JSON_HEADERS,
        ...options,
      })
    ).data,

  put: async <T = void>(url: string, data?: unknown, options?: WriteOptions): Promise<T> =>
    (
      await send<T>(instance, description, {
        method: 'put',
        url,
        data,
        headers: JSON_HEADERS,
        ...options,
      })
    ).data,

  del: async <T = void>(url: string, options?: WriteOptions): Promise<T> =>
    (
      await send<T>(instance, description, {
        method: 'delete',
        url,
        headers: JSON_HEADERS,
        ...options,
      })
    ).data,
})

/** The authenticated instance, which carries the Keycloak token. */
export const coreRequest = requestsThrough(axiosInstance, 'Core')

/** The public application pages, which are reached before there is a token. */
export const publicRequest = requestsThrough(notAuthenticatedAxiosInstance, 'Public core')
