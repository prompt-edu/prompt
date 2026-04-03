import { differenceInDays, differenceInHours, differenceInMinutes } from 'date-fns'

/**
 * Returns a human-readable duration string from now until the given date,
 * e.g. "3 days", "1 hour", "42 minutes". Falls back to days → hours → minutes.
 */
export function formatTimeRemaining(date: Date, now: Date): string {
  const days = differenceInDays(date, now)
  if (days >= 1) return `${days} ${days === 1 ? 'day' : 'days'}`
  const hours = differenceInHours(date, now)
  if (hours >= 1) return `${hours} ${hours === 1 ? 'hour' : 'hours'}`
  const minutes = Math.max(differenceInMinutes(date, now), 0)
  return `${minutes} ${minutes === 1 ? 'minute' : 'minutes'}`
}
