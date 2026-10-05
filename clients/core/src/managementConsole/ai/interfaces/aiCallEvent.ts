export enum AICallEventType {
  SHOWN = 'shown',
  ACCEPTED = 'accepted',
  EDITED = 'edited',
  REJECTED = 'rejected',
  CONTENT_VIEWED = 'content_viewed',
}

export interface AICallEvent {
  id: string
  actorId: string
  type: AICallEventType
  data: Record<string, unknown>
  createdAt: string
}
