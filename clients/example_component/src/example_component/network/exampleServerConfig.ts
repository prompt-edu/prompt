import { parseURL } from '@tumaet/prompt-shared-state'
import axios from 'axios'

// EXAMPLE_HOST is injected via core's env.js but not yet part of the
// @tumaet/prompt-shared-state EnvType, so it is read from window.env directly.
const exampleServer = (window.env as { EXAMPLE_HOST?: string }).EXAMPLE_HOST ?? ''

const serverBaseUrl = parseURL(exampleServer)

export interface Patch {
  op: 'replace' | 'add' | 'remove' | 'copy'
  path: string
  value: string
}

const authenticatedAxiosInstance = axios.create({
  baseURL: serverBaseUrl,
})

authenticatedAxiosInstance.interceptors.request.use((config) => {
  if (localStorage.getItem('jwt_token') && localStorage.getItem('jwt_token') !== '') {
    config.headers.Authorization = `Bearer ${localStorage.getItem('jwt_token') ?? ''}`
  }
  return config
})

export { authenticatedAxiosInstance as exampleAxiosInstance }
