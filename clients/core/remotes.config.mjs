export const REMOTES = [
  {
    name: 'example_component',
    phaseTypeName: 'example_component',
    devPort: 3001,
    prodPath: '/example',
  },
  {
    name: 'interview_component',
    phaseTypeName: 'Interview',
    devPort: 3002,
    prodPath: '/interview',
  },
  {
    name: 'matching_component',
    phaseTypeName: 'Matching',
    devPort: 3003,
    prodPath: '/matching',
  },
  {
    name: 'intro_course_developer_component',
    phaseTypeName: 'Intro Course Developer',
    devPort: 3005,
    prodPath: '/intro-course-developer',
  },
  {
    name: 'github_challenge_component',
    phaseTypeName: 'DevOps Challenge',
    devPort: 3006,
    prodPath: '/github-challenge',
  },
  {
    name: 'assessment_component',
    phaseTypeName: 'Assessment',
    devPort: 3007,
    prodPath: '/assessment',
  },
  {
    name: 'team_allocation_component',
    phaseTypeName: 'Team Allocation',
    devPort: 3008,
    prodPath: '/team-allocation',
  },
  {
    name: 'self_team_allocation_component',
    phaseTypeName: 'Self Team Allocation',
    devPort: 3009,
    prodPath: '/self-team-allocation',
  },
  {
    name: 'certificate_component',
    phaseTypeName: 'Certificate',
    devPort: 3010,
    prodPath: '/certificate',
  },
  {
    name: 'presentation_component',
    phaseTypeName: 'Presentation',
    devPort: 3011,
    prodPath: '/presentation',
  },
  {
    name: 'infrastructure_setup_component',
    phaseTypeName: 'Infrastructure Setup',
    devPort: 3012,
    prodPath: '/infrastructure-setup',
  },
]

export const resolveRemotes = (isDev) =>
  REMOTES.map(({ name, phaseTypeName, devPort, prodPath }) => ({
    name,
    phaseTypeName,
    url: isDev ? `http://localhost:${devPort}` : prodPath,
  }))
