import type { CoursePhaseStudentIdentifierProps } from '@tumaet/prompt-shared-state'
import { ApplicationStudentDetailComponent } from './ExternalStudentDetailComponents/ApplicationStudentDetailComponent'
import { AssessmentStudentDetailComponent } from './ExternalStudentDetailComponents/AssessmentStudentDetailComponent'
import { DevOpsStudentDetailComponent } from './ExternalStudentDetailComponents/DevOpsStudentDetailComponent'
import { InterviewStudentDetailComponent } from './ExternalStudentDetailComponents/InterviewStudentDetailComponent'
import { IntroCourseDeveloperStudentDetailComponent } from './ExternalStudentDetailComponents/IntroCourseDeveloperStudentDetailComponent'
import { MatchingStudentDetailComponent } from './ExternalStudentDetailComponents/MatchingStudentDetailComponent'
import { SelfTeamAllocationStudentDetailComponent } from './ExternalStudentDetailComponents/SelfTeamAllocationStudentDetailComponent'
import { TeamAllocationStudentDetailComponent } from './ExternalStudentDetailComponents/TeamAllocationStudentDetailComponent'

function Fallback() {
  return <></>
}

export const PhaseStudentDetailMapping: {
  [key: string]: React.FC<CoursePhaseStudentIdentifierProps>
} = {
  example_component: Fallback,
  Application: ApplicationStudentDetailComponent,
  Interview: InterviewStudentDetailComponent,
  Matching: MatchingStudentDetailComponent,
  'Intro Course Developer': IntroCourseDeveloperStudentDetailComponent,
  Assessment: AssessmentStudentDetailComponent,
  'DevOps Challenge': DevOpsStudentDetailComponent,
  'Team Allocation': TeamAllocationStudentDetailComponent,
  'Self Team Allocation': SelfTeamAllocationStudentDetailComponent,
}
