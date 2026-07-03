import { useCourseStore } from '@tumaet/prompt-shared-state'
import { Suspense } from 'react'
import { useParams } from 'react-router-dom'
import { ApplicationRoutes } from './ExternalRoutes/ApplicationRoutes'
import { AssessmentRoutes } from './ExternalRoutes/AssessmentRoutes'
import { CertificateRoutes } from './ExternalRoutes/CertificateRoutes'
import { DevOpsChallengeRoutes } from './ExternalRoutes/DevOpsChallengeRoutes'
import { InterviewRoutes } from './ExternalRoutes/InterviewRoutes'
import { IntroCourseDeveloperRoutes } from './ExternalRoutes/IntroCourseDeveloperRoutes'
import { MatchingRoutes } from './ExternalRoutes/MatchingRoutes'
import { SelfTeamAllocationRoutes } from './ExternalRoutes/SelfTeamAllocationRoutes'
import { TeamAllocationRoutes } from './ExternalRoutes/TeamAllocationRoutes'
import { TemplateRoutes } from './ExternalRoutes/TemplateRoutes'

const PhaseRouter: { [key: string]: React.FC } = {
  template_component: TemplateRoutes,
  Application: ApplicationRoutes,
  Interview: InterviewRoutes,
  Matching: MatchingRoutes,
  'Intro Course Developer': IntroCourseDeveloperRoutes,
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
