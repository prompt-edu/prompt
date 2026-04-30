import axios from 'axios'
import { parseURL } from '@tumaet/prompt-shared-state'
import { env } from '@tumaet/prompt-shared-state'

const infrastructureSetupServer = (env as any).INFRASTRUCTURE_SETUP_HOST || ''

const serverBaseUrl = parseURL(infrastructureSetupServer)

const authenticatedAxiosInstance = axios.create({
  baseURL: serverBaseUrl,
})

authenticatedAxiosInstance.interceptors.request.use((config) => {
  if (!!localStorage.getItem('jwt_token') && localStorage.getItem('jwt_token') !== '') {
    config.headers['Authorization'] = `Bearer ${localStorage.getItem('jwt_token') ?? ''}`
  }
  return config
})

export { authenticatedAxiosInstance as infrastructureSetupAxiosInstance }
