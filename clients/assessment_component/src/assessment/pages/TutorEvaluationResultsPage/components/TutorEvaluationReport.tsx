import type { Team } from '@tumaet/prompt-shared-state'
import { getStudentName } from '@tumaet/prompt-ui-components'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { Evaluation } from '../../../interfaces/evaluation'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { PrintReport } from '../../components/PrintReport/PrintReport'
import { getTeamMemberName } from '../../utils/getTeamMemberName'

interface TutorEvaluationReportProps {
  tutor: { firstName: string; lastName: string }
  subtitle: string
  tutorTitle: string
  categories: CategoryWithCompetencies[]
  evaluations: Evaluation[]
  feedbackItems: FeedbackItem[]
  teams: Team[]
  className?: string
}

export const TutorEvaluationReport = ({
  tutor,
  subtitle,
  tutorTitle,
  categories,
  evaluations,
  feedbackItems,
  teams,
  className,
}: TutorEvaluationReportProps) => {
  const scores = evaluations.map((evaluation) => ({
    ...evaluation,
    authorName: getTeamMemberName(teams, evaluation.authorCourseParticipationID),
  }))
  const evaluatorCount = new Set(
    evaluations.map((evaluation) => evaluation.authorCourseParticipationID),
  ).size

  return (
    <PrintReport
      className={className}
      title={`${tutorTitle} Evaluation Results for ${getStudentName(tutor)}`}
      subtitle={subtitle}
      meta={
        evaluatorCount > 1 ? (
          <span>
            <strong>Evaluators:</strong> {evaluatorCount}
          </span>
        ) : undefined
      }
      categories={categories}
      scores={scores}
      feedbackItems={feedbackItems}
    />
  )
}
