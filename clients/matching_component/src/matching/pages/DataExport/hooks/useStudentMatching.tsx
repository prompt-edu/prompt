import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { useEffect, useState } from 'react'
import type { UploadedStudent } from '../../../interfaces/UploadedStudent'
import { useMatchingStore } from '../../../zustand/useMatchingStore'

export const useStudentMatching = () => {
  const { participations, uploadedData } = useMatchingStore()
  const [matchedByMatriculation, setMatchedByMatriculation] = useState<UploadedStudent[]>([])
  const [matchedByName, setMatchedByName] = useState<UploadedStudent[]>([])
  const [unmatchedApplications, setUnmatchedApplications] = useState<
    CoursePhaseParticipationWithStudent[]
  >([])
  const [unmatchedStudents, setUnmatchedStudents] = useState<UploadedStudent[]>([])

  useEffect(() => {
    if (uploadedData?.length > 0 && participations) {
      const students: UploadedStudent[] = uploadedData
      const applications: CoursePhaseParticipationWithStudent[] = participations

      const matched: UploadedStudent[] = []
      const matchedByNameTemp: UploadedStudent[] = []
      const unmatchedApps: CoursePhaseParticipationWithStudent[] = []
      const unmatchedStuds: UploadedStudent[] = []

      students.forEach((student) => {
        const matchedApp = applications.find(
          (app) => app.student.matriculationNumber === student.matriculationNumber,
        )
        if (matchedApp) {
          matched.push({
            ...student,
            rank: matchedApp.prevData.score,
          })
        } else {
          const nameMatch = applications.find(
            (app) =>
              app.student.firstName.toLowerCase() === student.firstName.toLowerCase() &&
              app.student.lastName.toLowerCase() === student.lastName.toLowerCase(),
          )
          if (nameMatch) {
            matchedByNameTemp.push({
              ...student,
              rank: nameMatch.prevData.score,
            })
          } else {
            unmatchedStuds.push(student)
          }
        }
      })

      applications.forEach((app) => {
        if (
          !matched.some((s) => s.matriculationNumber === app.student.matriculationNumber) &&
          !matchedByNameTemp.some(
            (s) =>
              s.firstName.toLowerCase() === app.student.firstName.toLowerCase() &&
              s.lastName.toLowerCase() === app.student.lastName.toLowerCase(),
          )
        ) {
          unmatchedApps.push(app)
        }
      })

      setMatchedByMatriculation(matched)
      setMatchedByName(matchedByNameTemp)
      setUnmatchedApplications(unmatchedApps)
      setUnmatchedStudents(unmatchedStuds)
    }
  }, [uploadedData, participations])

  return { matchedByMatriculation, matchedByName, unmatchedApplications, unmatchedStudents }
}
