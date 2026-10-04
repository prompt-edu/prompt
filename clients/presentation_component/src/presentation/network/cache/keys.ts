type Id = string | undefined

export const presentationKeys = {
  presentations: {
    inPhase: (phaseId: Id) => ['presentations', phaseId] as const,
    own: (phaseId: Id) => ['presentations', phaseId, 'own'] as const,
  },
  slots: (phaseId: Id) => ['presentation-slots', phaseId] as const,
  targets: (phaseId: Id) => ['presentation-targets', phaseId] as const,
  config: (phaseId: Id) => ['presentation-config', phaseId] as const,
  categories: (phaseId: Id) => ['presentation-categories', phaseId] as const,
  materials: {
    inPhase: (phaseId: Id) => ['presentation-materials', phaseId] as const,
    ofPresentation: (phaseId: Id, presentationId: Id) =>
      ['presentation-materials', phaseId, presentationId] as const,
  },
  feedback: {
    inPhase: (phaseId: Id) => ['presentation-feedback', phaseId] as const,
    ofPresentation: (phaseId: Id, presentationId: Id) =>
      ['presentation-feedback', phaseId, presentationId] as const,
  },
}
