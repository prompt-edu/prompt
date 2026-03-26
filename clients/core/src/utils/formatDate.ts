export const formatDate = (value: string | Date): string => {
  const date = new Date(value)

  return date.toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

export function formatNoteDate(dateString: string): string {
  const date = new Date(dateString)

  return new Intl.DateTimeFormat('en-GB', {
    day: 'numeric',
    month: 'short',
    year: '2-digit',
    hour: 'numeric',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}
