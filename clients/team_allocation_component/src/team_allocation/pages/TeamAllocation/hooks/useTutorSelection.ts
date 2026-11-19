import { useCallback, useState } from 'react'
import type { Tutor, TutorCandidate } from '../../../interfaces/tutor'
import {
  buildTutorImport,
  type CurrentTutorTeams,
  currentTeamsOf,
  isSelectableTutor,
  tutorKey,
} from '../utils/tutorImport'

// Selection state shared by every source of the tutor import dialog: who is selected, and which
// teams each selected person leads, both keyed by normalized login. A person starts from the teams
// they lead today (currentTeams), since an import replaces them.
export const useTutorSelection = (currentTeams: CurrentTutorTeams) => {
  const [selectedTutors, setSelectedTutors] = useState<string[]>([])
  const [editedTeams, setEditedTeams] = useState<Record<string, string[]>>({})

  const teamsOf = useCallback(
    (candidate: TutorCandidate): string[] =>
      editedTeams[tutorKey(candidate)] ?? currentTeamsOf(currentTeams, candidate),
    [currentTeams, editedTeams],
  )

  const toggleTutor = useCallback((key: string) => {
    setSelectedTutors((prev) =>
      prev.includes(key) ? prev.filter((selected) => selected !== key) : [...prev, key],
    )
  }, [])

  const toggleAll = useCallback((candidates: TutorCandidate[]) => {
    const selectableKeys = [...new Set(candidates.filter(isSelectableTutor).map(tutorKey))]
    setSelectedTutors((prev) => (prev.length === selectableKeys.length ? [] : selectableKeys))
  }, [])

  const setTeams = useCallback((key: string, teamIds: string[]) => {
    setEditedTeams((prev) => ({ ...prev, [key]: teamIds }))
  }, [])

  const reset = useCallback(() => {
    setSelectedTutors([])
    setEditedTeams({})
  }, [])

  const buildImport = useCallback(
    (candidates: TutorCandidate[]): Tutor[] =>
      buildTutorImport(candidates, selectedTutors, teamsOf),
    [selectedTutors, teamsOf],
  )

  return { selectedTutors, teamsOf, toggleTutor, toggleAll, setTeams, reset, buildImport }
}
