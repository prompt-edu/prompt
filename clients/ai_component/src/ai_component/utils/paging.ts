// The call list is keyset-paginated: a page only knows the cursor of the page after it, so walking
// back needs the cursors already visited. The stack holds the cursor of every page beyond the
// newest one; an empty stack is the newest page.

export const currentCursor = <C>(stack: C[]): C | undefined => stack[stack.length - 1]

export const hasNewerPage = <C>(stack: C[]): boolean => stack.length > 0

export const pushOlderPage = <C>(stack: C[], next: C | null): C[] =>
  next ? [...stack, next] : stack

export const popNewerPage = <C>(stack: C[]): C[] => stack.slice(0, -1)
