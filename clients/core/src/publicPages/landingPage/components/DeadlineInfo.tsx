import { format } from 'date-fns'
import { toZonedTime } from 'date-fns-tz'
import { Clock } from 'lucide-react'

export const DeadlineInfo = ({ deadline }: { deadline: Date | string }) => {
  // Ensure deadline is a Date object (if it's a string, this converts it)
  const deadlineDate = new Date(deadline)

  const userTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone // Get the user's current timezone
  const berlinTimeZone = 'Europe/Berlin'
  const now = new Date()

  // Convert the deadline to Berlin time if needed
  const deadlineInBerlin = toZonedTime(deadlineDate, berlinTimeZone)

  // Format the deadline for display
  const formattedDeadline =
    userTimeZone === berlinTimeZone
      ? format(deadlineDate, "MMMM d, yyyy 'at' HH:mm")
      : format(deadlineInBerlin, "MMMM d, yyyy 'at' HH:mm '(Europe/Berlin)'")

  // Calculate the time remaining in milliseconds
  const timeRemainingMs = deadlineDate.getTime() - now.getTime()
  const isDeadlinePassed = timeRemainingMs <= 0

  // Use the actual remaining time (in ms) to decide if the deadline is "close" (within 3 days)
  const isDeadlineClose = timeRemainingMs < 72 * 3600 * 1000

  let timeRemainingText = ''

  if (isDeadlinePassed) {
    timeRemainingText = 'Deadline passed'
  } else if (timeRemainingMs < 3600 * 1000) {
    // Less than one hour remaining – display minutes (or "less than a minute")
    const minutesUntilDeadline = Math.floor(timeRemainingMs / (1000 * 60))
    timeRemainingText =
      timeRemainingMs < 60000
        ? 'less than a minute remaining'
        : `${minutesUntilDeadline} minute${minutesUntilDeadline !== 1 ? 's' : ''} remaining`
  } else {
    // At least one hour remains. We display days and hours.
    const totalHours = Math.floor(timeRemainingMs / (1000 * 60 * 60))
    const days = Math.floor(totalHours / 24)
    const hours = totalHours % 24

    const daysString = days > 0 ? `${days} day${days !== 1 ? 's' : ''}` : ''
    const hoursString = hours > 0 ? `${hours} hour${hours !== 1 ? 's' : ''}` : ''

    // Join the two parts if both are present.
    timeRemainingText = [daysString, hoursString].filter(Boolean).join(' and ') + ' remaining'
  }

  return (
    <div className={`space-y-1 ${isDeadlineClose ? 'text-red-600' : 'text-gray-600'}`}>
      <div className='flex items-center space-x-2'>
        <Clock className='h-4 w-4 shrink-0' />
        <p className='text-sm font-medium'>Apply by {formattedDeadline}</p>
      </div>
      <p className={`text-xs ${isDeadlineClose ? 'font-semibold' : ''} pl-6`}>
        {timeRemainingText}
      </p>
    </div>
  )
}
