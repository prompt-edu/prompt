import axios from 'axios'
import { env } from '@tumaet/prompt-shared-state'
import { parseURL } from '@tumaet/prompt-shared-state'

const templateServer = env.TEMPLATE_HOST || ''

const serverBaseUrl = parseURL(templateServer)

export interface Patch {
  op: 'replace' | 'add' | 'remove' | 'copy'
  path: string
  value: string
}

const authenticatedAxiosInstance = axios.create({
  baseURL: serverBaseUrl,
})

authenticatedAxiosInstance.interceptors.request.use((config) => {
  if (!!localStorage.getItem('jwt_token') && localStorage.getItem('jwt_token') !== '') {
    config.headers['Authorization'] = `Bearer ${localStorage.getItem('jwt_token') ?? ''}`
  }
  return config
})

export { authenticatedAxiosInstance as templateAxiosInstance }
