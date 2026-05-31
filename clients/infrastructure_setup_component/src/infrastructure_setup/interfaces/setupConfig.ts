export interface SetupConfig {
  coursePhaseId: string
  teamSourceCoursePhaseId?: string
  studentSourceCoursePhaseId?: string
  semesterTag: string
}

export interface UpsertSetupConfigRequest {
  teamSourceCoursePhaseId: string | null
  studentSourceCoursePhaseId: string | null
  semesterTag: string
}
