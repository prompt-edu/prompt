import { useCallback, useMemo, useState } from 'react'
import type { Tutor, TutorCandidate } from '../../../interfaces/tutor'
import { buildTutorImport, isSelectableTutor } from '../utils/tutorImport'

// Selection state shared by every source of the tutor import dialog: who is selected, and which
// teams each selected person leads. A person starts from the teams they lead today (currentTeams),
// since an import replaces them.
export const useTutorSelection = (currentTeams: Record<string, string[]>) => {
  const [selectedTutors, setSelectedTutors] = useState<string[]>([])
  const [editedTeams, setEditedTeams] = useState<Record<string, string[]>>({})
  const teamAssignments = useMemo(
    () => ({ ...currentTeams, ...editedTeams }),
    [currentTeams, editedTeams],
  )

  const toggleTutor = useCallback((tutorId: string) => {
    setSelectedTutors((prev) =>
      prev.includes(tutorId) ? prev.filter((id) => id !== tutorId) : [...prev, tutorId],
    )
  }, [])

  const toggleAll = useCallback((candidates: TutorCandidate[]) => {
    const selectableIds = candidates.filter(isSelectableTutor).map((candidate) => candidate.id)
    setSelectedTutors((prev) => (prev.length === selectableIds.length ? [] : selectableIds))
  }, [])

  const setTeams = useCallback((tutorId: string, teamIds: string[]) => {
    setEditedTeams((prev) => ({ ...prev, [tutorId]: teamIds }))
  }, [])

  const reset = useCallback(() => {
    setSelectedTutors([])
    setEditedTeams({})
  }, [])

  const buildImport = useCallback(
    (candidates: TutorCandidate[]): Tutor[] =>
      buildTutorImport(candidates, selectedTutors, teamAssignments),
    [selectedTutors, teamAssignments],
  )

  return { selectedTutors, teamAssignments, toggleTutor, toggleAll, setTeams, reset, buildImport }
}
