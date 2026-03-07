declare global {
  interface Window {
    env: any
  }
}

type EnvType = {
  ENVIRONMENT: 'development' | 'staging' | 'production'
  CORE_HOST: string
  INTRO_COURSE_HOST: string
  TEAM_ALLOCATION_HOST: string
  ASSESSMENT_HOST: string
  DEVOPS_CHALLENGE_HOST: string
  INTERVIEW_HOST: string
  KEYCLOAK_HOST: string
  KEYCLOAK_REALM_NAME: string
  CHAIR_NAME_LONG: string
  CHAIR_NAME_SHORT: string
  GITHUB_SHA: string
  GITHUB_REF: string
  SERVER_IMAGE_TAG: string
  SELF_TEAM_ALLOCATION_HOST: string
  TEMPLATE_HOST: string
  CERTIFICATE_HOST: string
  SENTRY_DSN_CLIENT: string
}
export const env: EnvType = { ...window.env }
