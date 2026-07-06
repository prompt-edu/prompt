import { ApplicationSidebar } from './ExternalSidebars/ApplicationSidebar'
import { AssessmentSidebar } from './ExternalSidebars/AssessmentSidebar'
import { CertificateSidebar } from './ExternalSidebars/CertificateSidebar'
import { DevOpsChallengeSidebar } from './ExternalSidebars/DevOpsChallengeSidebar'
import { InterviewSidebar } from './ExternalSidebars/InterviewSidebar'
import { IntroCourseDeveloperSidebar } from './ExternalSidebars/IntroCourseDeveloperSidebar'
import { MatchingSidebar } from './ExternalSidebars/MatchingSidebar'
import { SelfTeamAllocationSidebar } from './ExternalSidebars/SelfTeamAllocationSidebar'
import { TeamAllocationSidebar } from './ExternalSidebars/TeamAllocationSidebar'
import { TemplateSidebar } from './ExternalSidebars/TemplateSidebar'

export const PhaseSidebarMapping: {
  [key: string]: React.FC<{ rootPath: string; title: string; coursePhaseID: string }>
} = {
  template_component: TemplateSidebar,
  Application: ApplicationSidebar,
  Interview: InterviewSidebar,
  Matching: MatchingSidebar,
  'Intro Course Developer': IntroCourseDeveloperSidebar,
  Assessment: AssessmentSidebar,
  'DevOps Challenge': DevOpsChallengeSidebar,
  'Team Allocation': TeamAllocationSidebar,
  'Self Team Allocation': SelfTeamAllocationSidebar,
  Certificate: CertificateSidebar,
}
