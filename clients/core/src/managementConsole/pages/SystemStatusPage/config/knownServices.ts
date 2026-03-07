import { env } from '@/env'
import { KnownService } from '../interfaces/serviceCapabilities'

export const KNOWN_SERVICES: KnownService[] = [
  { name: 'Interview', apiBasePath: 'interview/api', host: env.INTERVIEW_HOST },
  { name: 'Intro Course Developer', apiBasePath: 'intro-course/api', host: env.INTRO_COURSE_HOST },
  { name: 'Assessment', apiBasePath: 'assessment/api', host: env.ASSESSMENT_HOST },
  { name: 'Team Allocation', apiBasePath: 'team-allocation/api', host: env.TEAM_ALLOCATION_HOST },
  {
    name: 'Self Team Allocation',
    apiBasePath: 'self-team-allocation/api',
    host: env.SELF_TEAM_ALLOCATION_HOST,
  },
]
