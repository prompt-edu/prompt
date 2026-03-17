import { useUpdateCoursePhaseParticipationBatch } from '@/hooks/useUpdateCoursePhaseParticipationBatch'
import { PassStatus } from '@tumaet/prompt-shared-state'
import { ParticipantRow } from '../table/participationRow'

export function useParticipantBatchActions(phaseId: string) {
  const { mutate: updateBatch } = useUpdateCoursePhaseParticipationBatch()

  const setStatus = (status: PassStatus, rows: ParticipantRow[]) => {
    updateBatch(
      rows.map((r) => ({
        coursePhaseID: phaseId,
        courseParticipationID: r.courseParticipationID,
        passStatus: status,
        restrictedData: {},
        studentReadableData: {},
      })),
    )
  }

  return {
    setPassed: (rows: ParticipantRow[]) => setStatus(PassStatus.PASSED, rows),
    setFailed: (rows: ParticipantRow[]) => setStatus(PassStatus.FAILED, rows),
  }
}
