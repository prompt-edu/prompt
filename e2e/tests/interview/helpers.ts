import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { BASE_URL, INTERVIEW_API } from '../../src/env'
import { FULL_COURSE_PHASES } from '../../src/data/constants'

export interface InterviewSlot {
  id: string
  coursePhaseId: string
  startTime: string
  endTime: string
  location: string | null
  capacity: number
  assignedCount: number
}

export function slotsUrl(phaseId = FULL_COURSE_PHASES.interview.id): string {
  return `${BASE_URL}${INTERVIEW_API}/course_phase/${phaseId}/interview-slots`
}

// Tomorrow at 10:00, so the UI treats the slot as bookable (past slots are
// disabled) and the 30-min slot never crosses midnight (the dialog's End Time
// is a time-of-day on the start's date).
export function futureSlotTimes(): { start: Date; end: Date } {
  const start = new Date(Date.now() + 24 * 60 * 60 * 1000)
  start.setHours(10, 0, 0, 0)
  const end = new Date(start.getTime() + 30 * 60 * 1000)
  return { start, end }
}

// Format for the schedule dialog's datetime-local inputs.
export function toDatetimeLocal(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}`
  )
}

// Format for the schedule dialog's time-of-day input (End Time).
export function toTimeOfDay(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export async function createSlot(
  api: APIRequestContext,
  location: string,
  capacity = 1,
): Promise<InterviewSlot> {
  const { start, end } = futureSlotTimes()
  const res = await api.post(slotsUrl(), {
    data: {
      startTime: start.toISOString(),
      endTime: end.toISOString(),
      location,
      capacity,
    },
  })
  if (!res.ok()) {
    throw new Error(`POST slot failed: ${res.status()} ${await res.text()}`)
  }
  return (await res.json()) as InterviewSlot
}

// Removes test-created slots by their FIXED per-spec location (deleting a slot
// cascades its assignments), so specs stay independent and re-runs after a
// failure or in UI watch mode start clean. Admin passes the staff-only delete.
export async function deleteSlotsByLocation(location: string) {
  const admin = await apiContextFor('admin')
  try {
    const res = await admin.get(slotsUrl())
    if (!res.ok()) {
      throw new Error(`GET slots failed: ${res.status()} ${await res.text()}`)
    }
    const slots = (await res.json()) as InterviewSlot[]
    for (const slot of slots.filter((s) => s.location === location)) {
      await admin.delete(`${slotsUrl()}/${slot.id}`)
    }
  } finally {
    await admin.dispose()
  }
}
