interface ServerErrorPayload {
  response?: { status?: unknown; data?: { error?: unknown } }
}

/**
 * Turns a failed request into a message worth showing.
 *
 * Every endpoint of this service reports a rejected request as `{"error": "..."}`, and
 * that text is the only thing that says which field was wrong. Falling straight through
 * to the Error message would show "Request failed with status code 400" instead.
 */
export const describeError = (err: unknown, fallback = 'Unknown error'): string => {
  const serverError = (err as ServerErrorPayload | null | undefined)?.response?.data?.error
  if (typeof serverError === 'string' && serverError.trim() !== '') {
    return serverError
  }
  if (err instanceof Error && err.message !== '') {
    return err.message
  }
  return fallback
}

/**
 * Reports whether a failed request came back with the given HTTP status. Keeps the
 * status check in one place instead of importing axios into a component for its type
 * guard.
 */
export const hasStatus = (err: unknown, status: number): boolean =>
  (err as ServerErrorPayload | null | undefined)?.response?.status === status
