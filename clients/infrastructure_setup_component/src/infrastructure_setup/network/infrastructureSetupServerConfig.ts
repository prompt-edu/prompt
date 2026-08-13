import { parseURL } from '@tumaet/prompt-shared-state'
import axios from 'axios'

// INFRASTRUCTURE_SETUP_HOST is injected via core's env.js but not yet part of the
// @tumaet/prompt-shared-state EnvType, so it is read from window.env directly.
const infrastructureSetupServer =
  (window.env as { INFRASTRUCTURE_SETUP_HOST?: string }).INFRASTRUCTURE_SETUP_HOST ?? ''

const serverBaseUrl = parseURL(infrastructureSetupServer)

const authenticatedAxiosInstance = axios.create({
  baseURL: serverBaseUrl,
})

authenticatedAxiosInstance.interceptors.request.use((config) => {
  if (localStorage.getItem('jwt_token') && localStorage.getItem('jwt_token') !== '') {
    config.headers['Authorization'] = `Bearer ${localStorage.getItem('jwt_token') ?? ''}`
  }
  return config
})

export { authenticatedAxiosInstance as infrastructureSetupAxiosInstance }
