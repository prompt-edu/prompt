import { assessmentAxiosInstance } from '../assessmentServerConfig'

export type AssessmentExportFormat = 'json'

export const exportStudentAssessment = async (
  coursePhaseID: string,
  courseParticipationID: string,
  format: AssessmentExportFormat,
): Promise<unknown> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/${courseParticipationID}/export`,
        {
          params: { format },
        },
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const triggerTextDownload = (content: string, filename: string, type: string): void => {
  const blob = new Blob([content], { type })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)
}
