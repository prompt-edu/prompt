import { request } from '@playwright/test'
import { MAILPIT_URL } from '../env'

interface MailpitMessage {
  ID: string
  Subject: string
  To: { Address: string; Name: string }[]
}

// Deletes all captured messages so a spec starts from a clean mailbox.
export async function clearMailpit(): Promise<void> {
  const ctx = await request.newContext()
  try {
    await ctx.delete(`${MAILPIT_URL}/api/v1/messages`)
  } finally {
    await ctx.dispose()
  }
}

// Returns all captured messages (most recent first).
export async function getMailpitMessages(): Promise<MailpitMessage[]> {
  const ctx = await request.newContext()
  try {
    const response = await ctx.get(`${MAILPIT_URL}/api/v1/messages`)
    const body = await response.json()
    return (body.messages ?? []) as MailpitMessage[]
  } finally {
    await ctx.dispose()
  }
}
