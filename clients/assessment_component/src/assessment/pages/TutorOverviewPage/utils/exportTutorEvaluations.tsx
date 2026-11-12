import type { Team } from '@tumaet/prompt-shared-state'
import { AssessmentType } from '../../../interfaces/assessmentType'
import { assessmentApi } from '../../../network/api'
import { TutorEvaluationReport } from '../../TutorEvaluationResultsPage/components/TutorEvaluationReport'
import { renderReportToPdf } from '../../utils/bulkExport/renderReportToPdf'
import {
  downloadZip,
  type ExportFile,
  exportDate,
  personFileBaseName,
} from '../../utils/bulkExport/zipFiles'
import { getTeamNamesOfTutor, type TutorRow } from '../../utils/getTutorRows'

// A tutor leading several teams has one row per team but gets one file with all their teams.
const fetchTutorEvaluations = (phaseId: string, rows: TutorRow[], teams: Team[]) =>
  Promise.all(
    [...new Map(rows.map((row) => [row.tutorId, row])).values()].map(async (tutor) => {
      const [evaluations, feedbackItems] = await Promise.all([
        assessmentApi.evaluations.ofTutor(phaseId, tutor.tutorId),
        assessmentApi.feedbackItems.ofTutor(phaseId, tutor.tutorId),
      ])
      const teamNames = getTeamNamesOfTutor(teams, tutor.tutorId)
      return { tutor, teamNames, evaluations, feedbackItems }
    }),
  )

export const exportTutorEvaluationsAsJson = async (
  phaseId: string,
  tutors: TutorRow[],
  teams: Team[],
) => {
  const exportedAt = new Date().toISOString()
  const files = (await fetchTutorEvaluations(phaseId, tutors, teams)).map(
    ({ tutor, teamNames, evaluations, feedbackItems }) => ({
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
            teamNames,
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
  teams: Team[],
) => {
  const [categories, tutorEvaluations] = await Promise.all([
    assessmentApi.categories.listWithCompetencies(phaseId, AssessmentType.TUTOR),
    fetchTutorEvaluations(phaseId, tutors, teams),
  ])
  const files: ExportFile[] = []
  for (const { tutor, teamNames, evaluations, feedbackItems } of tutorEvaluations) {
    const pdf = await renderReportToPdf(
      <TutorEvaluationReport
        className='block'
        tutor={tutor}
        subtitle={teamNames.join(', ')}
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
