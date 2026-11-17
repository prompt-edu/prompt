import type { Team } from '@tumaet/prompt-shared-state'
import { getStudentName } from '@tumaet/prompt-ui-components'
import type { AssessmentType } from '../../../interfaces/assessmentType'
import { assessmentApi } from '../../../network/api'
import { PrintReport } from '../../components/PrintReport/PrintReport'
import { renderReportToPdf } from '../../utils/bulkExport/renderReportToPdf'
import {
  downloadZip,
  type ExportFile,
  exportDate,
  personFileBaseName,
} from '../../utils/bulkExport/zipFiles'
import { getTeamMemberName } from '../../utils/getTeamMemberName'
import type { EvaluationParticipantRow } from '../interfaces/evaluationParticipantRow'

// Participants without any evaluation are skipped; returns how many reports were exported.
export const exportEvaluationReportsAsPdf = async (
  phaseId: string,
  rows: EvaluationParticipantRow[],
  assessmentType: AssessmentType.SELF | AssessmentType.PEER,
  reportTitle: string,
  teams: Team[],
): Promise<number> => {
  const [categories, evaluations, feedbackItems] = await Promise.all([
    assessmentApi.categories.listWithCompetencies(phaseId, assessmentType),
    assessmentApi.evaluations.listInPhase(phaseId),
    assessmentApi.feedbackItems.listInPhase(phaseId),
  ])
  const ofRow = <T extends { type: AssessmentType; courseParticipationID: string }>(
    items: T[],
    row: EvaluationParticipantRow,
  ) => items.filter((item) => item.type === assessmentType && item.courseParticipationID === row.id)

  const files: ExportFile[] = []
  for (const row of rows) {
    const scores = ofRow(evaluations, row).map((evaluation) => ({
      ...evaluation,
      authorName: getTeamMemberName(teams, evaluation.authorCourseParticipationID),
    }))
    if (scores.length === 0) continue
    const pdf = await renderReportToPdf(
      <PrintReport
        className='block'
        title={`${reportTitle} for ${getStudentName(row)}`}
        subtitle={row.teamName || undefined}
        categories={categories}
        scores={scores}
        feedbackItems={ofRow(feedbackItems, row)}
      />,
    )
    files.push({ baseName: personFileBaseName(row), extension: 'pdf', content: pdf })
  }
  if (files.length > 0) {
    await downloadZip(files, `${assessmentType}_evaluations_pdf_${exportDate()}.zip`)
  }
  return files.length
}
