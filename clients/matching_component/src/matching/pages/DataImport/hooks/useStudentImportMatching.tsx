import { type CoursePhaseParticipationWithStudent, PassStatus } from '@tumaet/prompt-shared-state'
import { useEffect, useState } from 'react'
import type { UploadedStudent } from '../../../interfaces/UploadedStudent'
import { useMatchingStore } from '../../../zustand/useMatchingStore'

export const useStudentImportMatching = () => {
  const { participations, uploadedData } = useMatchingStore()
  const [matchedStudents, setMatchedStudents] = useState<CoursePhaseParticipationWithStudent[]>([])
  const [unmatchedStudents, setUnmatchedStudents] = useState<UploadedStudent[]>([])

  useEffect(() => {
    if (uploadedData?.length > 0 && participations) {
      const data: UploadedStudent[] = uploadedData

      const matched: CoursePhaseParticipationWithStudent[] = []
      const unmatched: UploadedStudent[] = []

      data.forEach((student) => {
        const matchedParticipation = participations.find(
          (participation) =>
            participation.student.matriculationNumber === student.matriculationNumber,
        )
        if (matchedParticipation) {
          matched.push({
            ...matchedParticipation,
            passStatus: PassStatus.PASSED,
          })
        } else {
          const nameMatch = participations.find(
            (participation) =>
              participation.student.firstName.toLowerCase() === student.firstName.toLowerCase() &&
              participation.student.lastName.toLowerCase() === student.lastName.toLowerCase(),
          )
          if (nameMatch) {
            matched.push({
              ...nameMatch,
              passStatus: PassStatus.PASSED,
            })
          } else {
            unmatched.push(student)
          }
        }
      })

      setMatchedStudents(matched)
      setUnmatchedStudents(unmatched)
    }
  }, [uploadedData, participations])

  return { matchedStudents, unmatchedStudents }
}
