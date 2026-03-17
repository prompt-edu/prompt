import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import { TemplateRoutes } from './ExternalRoutes/TemplateRoutes'
import { ApplicationRoutes } from './ExternalRoutes/ApplicationRoutes'
import { InterviewRoutes } from './ExternalRoutes/InterviewRoutes'
import { Suspense } from 'react'
import { MatchingRoutes } from './ExternalRoutes/MatchingRoutes'
import { IntroCourseDeveloperRoutes } from './ExternalRoutes/IntroCourseDeveloperRoutes'
import { IntroCourseTutorRoutes } from './ExternalRoutes/IntroCourseTutorRoutes'
import { AssessmentRoutes } from './ExternalRoutes/AssessmentRoutes'
import { DevOpsChallengeRoutes } from './ExternalRoutes/DevOpsChallengeRoutes'
import { TeamAllocationRoutes } from './ExternalRoutes/TeamAllocationRoutes'
import { SelfTeamAllocationRoutes } from './ExternalRoutes/SelfTeamAllocationRoutes'
import { CertificateRoutes } from './ExternalRoutes/CertificateRoutes'

const PhaseRouter: { [key: string]: React.FC } = {
  template_component: TemplateRoutes,
  Application: ApplicationRoutes,
  Interview: InterviewRoutes,
  Matching: MatchingRoutes,
  'Intro Course Developer': IntroCourseDeveloperRoutes,
  IntroCourseTutor: IntroCourseTutorRoutes,
  Assessment: AssessmentRoutes,
  'DevOps Challenge': DevOpsChallengeRoutes,
  'Team Allocation': TeamAllocationRoutes,
  'Self Team Allocation': SelfTeamAllocationRoutes,
  Certificate: CertificateRoutes,
}

export const PhaseRouterMapping = () => {
  const phaseId = useParams<{ phaseId: string }>().phaseId
  const courseId = useParams<{ courseId: string }>().courseId
  const { courses } = useCourseStore()

  const selectedPhase = courses
    .find((c) => c.id === courseId)
    ?.coursePhases.find((p) => p.id === phaseId)

  if (!selectedPhase) {
    return <div>Phase not found</div>
  }

  const PhaseComponent = PhaseRouter[selectedPhase.coursePhaseType]

  if (!PhaseComponent) {
    return <div>Phase Module not found</div>
  }

  return (
    <Suspense fallback={<div>Loading...</div>}>
      <PhaseComponent />
    </Suspense>
  )
}
