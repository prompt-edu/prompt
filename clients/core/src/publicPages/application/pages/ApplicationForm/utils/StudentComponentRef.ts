import type { Student } from '@tumaet/prompt-shared-state'

export interface StudentComponentRef {
  validate: () => Promise<boolean>
  rerender: (updatedStudent: Student) => void
}
