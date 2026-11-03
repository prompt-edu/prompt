import { useCallback, useState } from 'react'
import type { Tutor, TutorCandidate } from '../../../interfaces/tutor'
import { buildTutorImport, isSelectableTutor } from '../utils/tutorImport'

// Selection state shared by every source of the tutor import dialog: who is selected, and which
// teams each selected person leads.
export const useTutorSelection = () => {
  const [selectedTutors, setSelectedTutors] = useState<string[]>([])
  const [teamAssignments, setTeamAssignments] = useState<Record<string, string[]>>({})

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
    setTeamAssignments((prev) => ({ ...prev, [tutorId]: teamIds }))
  }, [])

  const reset = useCallback(() => {
    setSelectedTutors([])
    setTeamAssignments({})
  }, [])

  const buildImport = useCallback(
    (candidates: TutorCandidate[]): Tutor[] =>
      buildTutorImport(candidates, selectedTutors, teamAssignments),
    [selectedTutors, teamAssignments],
  )

  return { selectedTutors, teamAssignments, toggleTutor, toggleAll, setTeams, reset, buildImport }
}
