import { useState, useRef, useEffect } from 'react'
import { InstructorNoteTag } from './InstructorNoteTag'
import { NoteTag } from '../../interfaces/InstructorNote'
import { InstructorNoteComposerTagPicker } from './InstructorNoteFormElements/InstructorNoteComposerTagPicker'
import { InstructorNoteComposerSubmitButton } from './InstructorNoteFormElements/InstructorNoteComposerSubmitButton'

interface NoteComposerProps {
  onSubmit: (content: string, tagIds: string[]) => Promise<void>
  isPending: boolean
  onCancel?: () => void
  initialContent?: string
  initialTags?: NoteTag[]
  autoFocus?: boolean
}

export function NoteComposer({
  onSubmit,
  isPending,
  onCancel,
  initialContent = '',
  initialTags = [],
  autoFocus,
}: NoteComposerProps) {
  const [content, setContent] = useState(initialContent)
  const [selectedTags, setSelectedTags] = useState<NoteTag[]>(initialTags)
  const [isMultiLine, setIsMultiLine] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const singleLineHeightRef = useRef(0)

  useEffect(() => {
    const el = textareaRef.current
    if (!el) return
    // Measure natural single-line height
    const saved = el.value
    el.value = 'x'
    el.style.height = 'auto'
    singleLineHeightRef.current = el.scrollHeight
    el.value = saved
    // Resize for initial content (edit mode)
    if (initialContent) {
      el.style.height = 'auto'
      el.style.height = `${el.scrollHeight}px`
      setIsMultiLine(el.scrollHeight > singleLineHeightRef.current)
    }
  }, [initialContent])

  const sendUnavailable = !content.trim() || isPending
  const isEditMode = onCancel !== undefined

  const toggleTag = (tag: NoteTag) => {
    setSelectedTags((prev) =>
      prev.some((t) => t.id === tag.id) ? prev.filter((t) => t.id !== tag.id) : [...prev, tag],
    )
  }

  const handleSubmit = async () => {
    if (!content.trim()) return
    try {
      await onSubmit(
        content.trim(),
        selectedTags.map((t) => t.id),
      )
      setContent('')
      setSelectedTags([])
      setIsMultiLine(false)
      if (textareaRef.current) textareaRef.current.style.height = 'auto'
    } catch {
      // error handled upstream; keep form open
    }
  }

  return (
    <div className='space-y-1.5'>
      {selectedTags.length > 0 && (
        <div className='flex items-center flex-wrap gap-1'>
          {selectedTags.map((tag) => (
            <InstructorNoteTag key={tag.id} tag={tag} />
          ))}
        </div>
      )}

      <div
        className={
          `flex ${isMultiLine ? 'items-end' : 'items-start'} ` +
          'gap-1 rounded-md border border-input bg-background focus-within:ring-1 focus-within:ring-ring px-2 py-1.5'
        }
      >
        <textarea
          ref={textareaRef}
          placeholder={isEditMode ? undefined : 'Leave an instructor note'}
          className={
            `ml-1 flex-1 min-w-0 bg-transparent pt-0.5 ${isMultiLine ? 'pb-0.5' : ''} ` +
            'text-sm placeholder:text-muted-foreground focus:outline-hidden resize-none overflow-hidden placeholder:text-nowrap'
          }
          value={content}
          rows={1}
          onChange={(e) => {
            setContent(e.target.value)
            e.target.style.height = 'auto'
            const newHeight = e.target.scrollHeight
            e.target.style.height = `${newHeight}px`
            setIsMultiLine(newHeight > singleLineHeightRef.current)
          }}
          autoFocus={autoFocus}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handleSubmit()
            if (e.key === 'Escape' && isEditMode) onCancel?.()
          }}
        />
        <InstructorNoteComposerTagPicker selectedTags={selectedTags} onToggle={toggleTag} />
        <InstructorNoteComposerSubmitButton
          onClick={handleSubmit}
          disabled={sendUnavailable}
          isEditMode={isEditMode}
        />
      </div>
    </div>
  )
}
