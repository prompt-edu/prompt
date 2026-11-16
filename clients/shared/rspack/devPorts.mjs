// Dev server ports of a local stack. The configs name every port as it is in
// stack 0 (core 3000, each remote its own 30xx). A parallel stack shifts all of
// them by the same amount: the Makefile hands the client its base port as
// PROMPT_CLIENT_PORT, and every port moves by its distance from 3000.
// Docs: docs/contributor/setup.md ("Run several stacks side by side").

export const DEFAULT_CLIENT_PORT = 3000

/** How far this stack's ports sit from stack 0's. */
export const devPortOffset = (env = process.env) => {
  const raw = env.PROMPT_CLIENT_PORT
  if (!raw) return 0
  const base = Number(raw)
  if (!Number.isInteger(base) || base <= 0) {
    throw new Error(`PROMPT_CLIENT_PORT must be a port number (got "${raw}")`)
  }
  return base - DEFAULT_CLIENT_PORT
}

/** The port that stack 0's `port` has in this stack. */
export const devPort = (port, env = process.env) => port + devPortOffset(env)

/**
 * Shifts every `localhost:<port>` in source by offset. The core client's
 * public/env.js points at the APIs and Keycloak of stack 0, so the dev server
 * serves it shifted to this stack.
 */
export const shiftLocalhostPorts = (source, offset) =>
  source.replace(/localhost:(\d+)/g, (_, port) => `localhost:${Number(port) + offset}`)
