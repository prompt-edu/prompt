import type { ParticipantRow } from '@tumaet/prompt-ui-components'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import { assessmentApi } from '../../../network/api'
import { AssessmentReport } from '../../components/AssessmentPrintReport/AssessmentPrintReport'
import { renderReportToPdf } from '../../utils/bulkExport/renderReportToPdf'
import {
  downloadZip,
  type ExportFile,
  exportDate,
  personFileBaseName,
} from '../../utils/bulkExport/zipFiles'

const fetchExport = (phaseId: string, row: ParticipantRow) =>
  assessmentApi.assessments.export(phaseId, row.courseParticipationID, 'json')

export const exportAssessmentsAsJson = async (phaseId: string, rows: ParticipantRow[]) => {
  const files = await Promise.all(
    rows.map(async (row) => ({
      baseName: personFileBaseName(row),
      extension: 'json',
      content: JSON.stringify(await fetchExport(phaseId, row), null, 2),
    })),
  )
  await downloadZip(files, `assessments_json_${exportDate()}.zip`)
}

export const exportAssessmentsAsPdf = async (
  phaseId: string,
  rows: ParticipantRow[],
  categories: CategoryWithCompetencies[],
  includeFeedback: boolean,
) => {
  const reports = await Promise.all(
    rows.map(async (row) => {
      const [exported, feedbackItems] = await Promise.all([
        fetchExport(phaseId, row),
        includeFeedback
          ? assessmentApi.feedbackItems.ofStudent(phaseId, row.courseParticipationID)
          : [],
      ])
      return { row, exported, feedbackItems: feedbackItems.filter((item) => item.feedbackText) }
    }),
  )

  const files: ExportFile[] = []
  for (const { row, exported, feedbackItems } of reports) {
    const { studentAssessment } = exported
    const pdf = await renderReportToPdf(
      <AssessmentReport
        className='block'
        student={row.student}
        categories={categories}
        assessments={studentAssessment.assessments}
        categoryAssessments={studentAssessment.categoryAssessments ?? []}
        assessmentCompletion={studentAssessment.assessmentCompletion}
        studentScore={studentAssessment.studentScore}
        actionItems={exported.actionItems}
        feedbackItems={feedbackItems}
      />,
    )
    files.push({ baseName: personFileBaseName(row), extension: 'pdf', content: pdf })
  }
  await downloadZip(files, `assessments_pdf_${exportDate()}.zip`)
}
