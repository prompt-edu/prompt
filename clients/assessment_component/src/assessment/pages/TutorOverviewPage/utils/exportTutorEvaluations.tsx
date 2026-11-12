import type { Team } from '@tumaet/prompt-shared-state'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import { assessmentApi } from '../../../network/api'
import { TutorEvaluationReport } from '../../TutorEvaluationResultsPage/components/TutorEvaluationReport'
import { renderReportToPdf } from '../../utils/bulkExport/renderReportToPdf'
import {
  downloadZip,
  type ExportFile,
  exportDate,
  personFileBaseName,
} from '../../utils/bulkExport/zipFiles'
import type { TutorRow } from '../../utils/getTutorRows'

const fetchTutorEvaluations = (phaseId: string, tutors: TutorRow[]) =>
  Promise.all(
    tutors.map(async (tutor) => {
      const [evaluations, feedbackItems] = await Promise.all([
        assessmentApi.evaluations.ofTutor(phaseId, tutor.tutorId),
        assessmentApi.feedbackItems.ofTutor(phaseId, tutor.tutorId),
      ])
      return { tutor, evaluations, feedbackItems }
    }),
  )

export const exportTutorEvaluationsAsJson = async (phaseId: string, tutors: TutorRow[]) => {
  const exportedAt = new Date().toISOString()
  const files = (await fetchTutorEvaluations(phaseId, tutors)).map(
    ({ tutor, evaluations, feedbackItems }) => ({
      baseName: personFileBaseName(tutor),
      extension: 'json',
      content: JSON.stringify(
        {
          exportedAt,
          coursePhaseID: phaseId,
          tutor: {
            courseParticipationID: tutor.tutorId,
            firstName: tutor.firstName,
            lastName: tutor.lastName,
            teamName: tutor.teamName,
          },
          evaluations,
          feedbackItems,
        },
        null,
        2,
      ),
    }),
  )
  await downloadZip(files, `tutor_evaluations_json_${exportDate()}.zip`)
}

export const exportTutorEvaluationsAsPdf = async (
  phaseId: string,
  tutors: TutorRow[],
  tutorTitle: string,
  categories: CategoryWithCompetencies[],
  teams: Team[],
) => {
  const files: ExportFile[] = []
  for (const { tutor, evaluations, feedbackItems } of await fetchTutorEvaluations(
    phaseId,
    tutors,
  )) {
    const pdf = await renderReportToPdf(
      <TutorEvaluationReport
        className='block'
        tutor={tutor}
        subtitle={tutor.teamName}
        tutorTitle={tutorTitle}
        categories={categories}
        evaluations={evaluations}
        feedbackItems={feedbackItems}
        teams={teams}
      />,
    )
    files.push({ baseName: personFileBaseName(tutor), extension: 'pdf', content: pdf })
  }
  await downloadZip(files, `tutor_evaluations_pdf_${exportDate()}.zip`)
}
