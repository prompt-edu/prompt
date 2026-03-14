import { formatNoteDate } from '@core/utils/formatDate'

interface NoteVersionHistoryItemProps {
  content: string
  dateCreated: string
}

export function NoteVersionHistoryItem({ content, dateCreated }: NoteVersionHistoryItemProps) {
  return (
    <div className='text-sm text-gray-600'>
      <span className='block text-xs text-gray-400'>{formatNoteDate(dateCreated)}</span>
      <p className='whitespace-pre-wrap'>{content}</p>
    </div>
  )
}
