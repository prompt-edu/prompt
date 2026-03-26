import { NoteTag, NoteTagColor } from '../../interfaces/InstructorNote'

const tagColorClass: Record<NoteTagColor, string> = {
  green: 'bg-green-200 border-green-600 dark:bg-green-800',
  blue: 'bg-blue-200 border-blue-600 dark:bg-blue-800',
  red: 'bg-red-200 border-red-600 dark:bg-red-800',
  yellow: 'bg-yellow-200 border-yellow-600 dark:bg-yellow-800',
  orange: 'bg-orange-200 border-orange-600 dark:bg-orange-800',
  pink: 'bg-pink-200 border-pink-600 dark:bg-pink-800',
}

const bgFromTagColor: Record<NoteTagColor, string> = {
  blue: 'bg-blue-500',
  green: 'bg-green-500',
  red: 'bg-red-500',
  yellow: 'bg-yellow-500',
  orange: 'bg-orange-500',
  pink: 'bg-pink-500',
}

export function InstructorNoteTag({ tag }: { tag: NoteTag }) {
  return (
    <div
      className={
        'text-xs border-l-2 px-1 text-nowrap whitespace-nowrap inline-block text-black dark:text-white ' +
        tagColorClass[tag.color]
      }
    >
      {tag.name}
    </div>
  )
}

export function InstructorNoteTagColor({ color }: { color: NoteTagColor }) {
  return (
    <div className='flex items-center gap-2'>
      <span className={`inline-block h-3 w-3 rounded-full ${bgFromTagColor[color]}`} />
      <span className='capitalize'>{color}</span>
    </div>
  )
}

export function InstructorNoteTags({ tags }: { tags: NoteTag[] }) {
  return (
    <div className='flex gap-2'>
      {tags.map((tag: NoteTag) => (
        <InstructorNoteTag key={tag.id} tag={tag} />
      ))}
    </div>
  )
}
