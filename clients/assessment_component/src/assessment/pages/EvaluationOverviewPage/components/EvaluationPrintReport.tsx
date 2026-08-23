import type { CategoryWithCompetencies } from '../../../interfaces/category'
import { PrintReport, type PrintReportScore } from '../../components/PrintReport/PrintReport'

export interface EvaluationPrintSection {
  categories: CategoryWithCompetencies[]
  scores: PrintReportScore[]
}

interface EvaluationPrintReportProps {
  self: EvaluationPrintSection
  peer: EvaluationPrintSection
}

export const EvaluationPrintReport = ({ self, peer }: EvaluationPrintReportProps) => {
  const hasSelf = self.scores.length > 0
  const hasPeer = peer.scores.length > 0

  return (
    <>
      {hasSelf && (
        <PrintReport
          title='Evaluation Results'
          subtitle='Your self-evaluation'
          categories={self.categories}
          scores={self.scores}
        />
      )}
      {hasPeer && (
        <PrintReport
          title='Peer Feedback'
          subtitle='Averaged across your teammates'
          categories={peer.categories}
          scores={peer.scores}
          className={hasSelf ? 'break-before-page' : undefined}
        />
      )}
    </>
  )
}
