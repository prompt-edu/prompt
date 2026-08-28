import { format } from 'date-fns'

// Kept deliberately in step with the interview phase's slot creation
// (clients/interview_component/src/interview/pages/ScheduleManagement), so scheduling a
// presentation series behaves exactly like scheduling interviews.
export const MAX_SERIES_SLOTS = 100

export interface SlotFormData {
  // A full local date and time, as produced by an input of type datetime-local.
  startTime: string
  // A time of day (HH:mm) on the start date, as produced by an input of type time.
  endTime: string
  durationMinutes: number
  breakMinutes: number
  location: string
}

export interface SlotTimeRange {
  start: Date
  end: Date
}

export interface GeneratedSeries {
  slots: SlotTimeRange[]
  truncated: boolean
}

export const SERIES_DEFAULTS = { durationMinutes: 30, breakMinutes: 0 }

export const EMPTY_SERIES: GeneratedSeries = { slots: [], truncated: false }

export const emptySlotForm = (): SlotFormData => ({
  startTime: '',
  endTime: '',
  ...SERIES_DEFAULTS,
  location: '',
})

export const slotFormFromTimes = (
  startTime: string,
  endTime: string,
  location: string,
): SlotFormData => ({
  startTime: format(new Date(startTime), "yyyy-MM-dd'T'HH:mm"),
  endTime: format(new Date(endTime), 'HH:mm'),
  ...SERIES_DEFAULTS,
  location,
})

// endTime is a time-of-day (HH:mm); an end at or before the start means the slot runs past midnight.
export const buildSlotTimes = (startDateTime: string, endTimeOfDay: string): SlotTimeRange => {
  const start = new Date(startDateTime)
  const end = new Date(`${startDateTime.slice(0, 10)}T${endTimeOfDay}`)
  if (end <= start) {
    end.setDate(end.getDate() + 1)
  }
  return { start, end }
}

export const formatResolvedEnd = (range: SlotTimeRange): string => {
  const spansNextDay = range.end.getDate() !== range.start.getDate()
  return `Ends ${format(range.end, 'EEE, MMM d')} at ${format(range.end, 'HH:mm')}${
    spansNextDay ? ' (next day)' : ''
  }`
}

export const formatSlotCount = (count: number, capitalize = false): string => {
  const noun = capitalize ? 'Slot' : 'slot'
  return `${count} ${noun}${count === 1 ? '' : 's'}`
}

// Steps in wall-clock minutes rather than milliseconds, so a DST transition inside the range does
// not shift the generated slot times by an hour.
const addLocalMinutes = (base: Date, minutes: number) =>
  new Date(
    base.getFullYear(),
    base.getMonth(),
    base.getDate(),
    base.getHours(),
    base.getMinutes() + minutes,
  )

export const generateSlotSeries = (data: SlotFormData): GeneratedSeries => {
  const slots: SlotTimeRange[] = []
  if (!data.startTime || !data.endTime || data.durationMinutes <= 0) return EMPTY_SERIES

  const { start, end } = buildSlotTimes(data.startTime, data.endTime)
  const stepMinutes = data.durationMinutes + Math.max(0, data.breakMinutes)

  for (let offset = 0; ; offset += stepMinutes) {
    const slotEnd = addLocalMinutes(start, offset + data.durationMinutes)
    if (slotEnd > end) break
    if (slots.length === MAX_SERIES_SLOTS) return { slots, truncated: true }
    slots.push({ start: addLocalMinutes(start, offset), end: slotEnd })
  }

  return { slots, truncated: false }
}
