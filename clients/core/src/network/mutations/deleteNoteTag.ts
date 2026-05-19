import { axiosInstance } from '@tumaet/prompt-shared-state'

export const deleteNoteTag = async (tagId: string): Promise<void> => {
  try {
    await axiosInstance.delete(`/api/instructor-notes/tags/${tagId}`, {
      headers: {
        'Content-Type': 'application/json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
