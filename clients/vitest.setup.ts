const TEST_ENV = {
  ENVIRONMENT: 'development',
  CORE_HOST: 'http://core.test',
  INTRO_COURSE_HOST: 'http://intro-course.test',
  TEAM_ALLOCATION_HOST: 'http://team-allocation.test',
  ASSESSMENT_HOST: 'http://assessment.test',
  DEVOPS_CHALLENGE_HOST: 'http://devops-challenge.test',
  INTERVIEW_HOST: 'http://interview.test',
  KEYCLOAK_HOST: 'http://keycloak.test',
  KEYCLOAK_REALM_NAME: 'prompt',
  CHAIR_NAME_LONG: 'Test Chair',
  CHAIR_NAME_SHORT: 'Test Chair',
  GITHUB_SHA: 'test-sha',
  GITHUB_REF: 'test-ref',
  SERVER_IMAGE_TAG: 'test-tag',
  SELF_TEAM_ALLOCATION_HOST: 'http://self-team-allocation.test',
  TEMPLATE_HOST: 'http://template.test',
  CERTIFICATE_HOST: 'http://certificate.test',
  SENTRY_DSN_CLIENT: 'http://sentry.test',
}

const memoryStorage = new Map<string, string>()

// Node exposes localStorage behind a flag, and reading the global warns; the shared axios instance
// checks for it on every request
const localStorageStub = {
  getItem: (key: string) => memoryStorage.get(key) ?? null,
  setItem: (key: string, value: string) => memoryStorage.set(key, value),
  removeItem: (key: string) => memoryStorage.delete(key),
  clear: () => memoryStorage.clear(),
}

Object.assign(globalThis, { window: { env: TEST_ENV }, localStorage: localStorageStub })
