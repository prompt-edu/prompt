// Core interview slot type used by Zustand store
export interface InterviewSlot {
  id: string
  index?: number
  startTime?: string
  endTime?: string
  courseParticipationID?: string
}

// Assignment information returned by the server
export interface AssignmentInfo {
  id: string
  courseParticipationId: string
  assignedAt: string
  student?: {
    id: string
    firstName: string
    lastName: string
    email: string
  }
}

// Full interview slot with assignments (used by API responses)
export interface InterviewSlotWithAssignments {
  id: string
  coursePhaseId?: string
  startTime: string
  endTime: string
  location: string | null
  capacity?: number
  assignedCount?: number
  assignments: AssignmentInfo[]
  createdAt?: string
  updatedAt?: string
}
