import axios from 'axios'
import { parseURL } from '@/utils/parseURL'
import { env } from '@/env'

const certificateServer = env.CERTIFICATE_HOST || ''

const serverBaseUrl = parseURL(certificateServer)

const authenticatedAxiosInstance = axios.create({
  baseURL: serverBaseUrl,
})

authenticatedAxiosInstance.interceptors.request.use((config) => {
  if (!!localStorage.getItem('jwt_token') && localStorage.getItem('jwt_token') !== '') {
    config.headers['Authorization'] = `Bearer ${localStorage.getItem('jwt_token') ?? ''}`
  }
  return config
})

export { authenticatedAxiosInstance as certificateAxiosInstance }
