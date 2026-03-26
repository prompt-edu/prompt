export interface NoteVersion {
  id: string
  content: string
  dateCreated: string
  versionNumber: number
}

export type NoteTagColor = 'blue' | 'green' | 'red' | 'yellow' | 'orange' | 'pink'

export interface NoteTag {
  id: string
  name: string
  color: NoteTagColor
}

export interface CreateNoteTag {
  name: string
  color: NoteTagColor
}

export interface UpdateNoteTag {
  name: string
  color: NoteTagColor
}

export interface InstructorNote {
  id: string
  author: string
  authorName: string
  authorEmail: string
  forStudent: string
  dateCreated: string
  dateDeleted: string | null
  deletedBy: string | null
  versions: NoteVersion[]
  tags: NoteTag[]
}

export interface CreateInstructorNote {
  content: string
  new: boolean
  forNote?: string
  tags?: string[]
}
