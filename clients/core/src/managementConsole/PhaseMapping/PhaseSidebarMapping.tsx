import { ApplicationSidebar } from './ExternalSidebars/ApplicationSidebar'
import { AssessmentSidebar } from './ExternalSidebars/AssessmentSidebar'
import { CertificateSidebar } from './ExternalSidebars/CertificateSidebar'
import { DevOpsChallengeSidebar } from './ExternalSidebars/DevOpsChallengeSidebar'
import { ExampleSidebar } from './ExternalSidebars/ExampleSidebar'
import { InterviewSidebar } from './ExternalSidebars/InterviewSidebar'
import { IntroCourseDeveloperSidebar } from './ExternalSidebars/IntroCourseDeveloperSidebar'
import { MatchingSidebar } from './ExternalSidebars/MatchingSidebar'
import { PresentationSidebar } from './ExternalSidebars/PresentationSidebar'
import { SelfTeamAllocationSidebar } from './ExternalSidebars/SelfTeamAllocationSidebar'
import { TeamAllocationSidebar } from './ExternalSidebars/TeamAllocationSidebar'

export const PhaseSidebarMapping: {
  [key: string]: React.FC<{ rootPath: string; title: string; coursePhaseID: string }>
} = {
  example_component: ExampleSidebar,
  Application: ApplicationSidebar,
  Interview: InterviewSidebar,
  Matching: MatchingSidebar,
  'Intro Course Developer': IntroCourseDeveloperSidebar,
  Assessment: AssessmentSidebar,
  'DevOps Challenge': DevOpsChallengeSidebar,
  'Team Allocation': TeamAllocationSidebar,
  'Self Team Allocation': SelfTeamAllocationSidebar,
  Certificate: CertificateSidebar,
  Presentation: PresentationSidebar,
}
