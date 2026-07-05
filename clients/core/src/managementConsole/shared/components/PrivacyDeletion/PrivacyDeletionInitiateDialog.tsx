import {
  adminInitiateDataDeletions,
  DeletionRequestStatus,
  getDataDeletionsStatus,
  type PrivacyDeletionRequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'
import { useQueryClient } from '@tanstack/react-query'
import { AlertCircle, CheckCircle2, Loader2, Users, XCircle } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import {
  PrivacyDeletionHowItWorks,
  PrivacyDeletionWhatGetsDeleted,
} from './PrivacyDeletionExplainerContent'

const BATCH_SIZE = 10
const WAIT_SECONDS = 5
const POLL_INTERVAL_MS = 3000
const POLL_TIMEOUT_MS = 35 * 60 * 1000

interface PrivacyDeletionInitiateDialogProps {
  studentIDs: string[]
  open: boolean
  onClose: () => void
}

type Phase = 'idle' | 'running' | 'done' | 'error'

function isTerminal(status: DeletionRequestStatus): boolean {
  return (
    status === DeletionRequestStatus.succeeded ||
    status === DeletionRequestStatus.failed ||
    status === DeletionRequestStatus.rejected
  )
}

function chunk<T>(arr: T[], size: number): T[][] {
  const out: T[][] = []
  for (let i = 0; i < arr.length; i += size) out.push(arr.slice(i, i + size))
  return out
}

export function PrivacyDeletionInitiateDialog({
  studentIDs,
  open,
  onClose,
}: PrivacyDeletionInitiateDialogProps) {
  const queryClient = useQueryClient()
  const batches = useMemo(() => chunk(studentIDs, BATCH_SIZE), [studentIDs])

  const [triggerCooldownRemaining, setTriggerCooldownRemaining] = useState(WAIT_SECONDS)
  const [phase, setPhase] = useState<Phase>('idle')
  const [currentBatch, setCurrentBatch] = useState(0)
  const [batchTerminalCount, setBatchTerminalCount] = useState(0)
  const [results, setResults] = useState<PrivacyDeletionRequest[]>([])
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setTriggerCooldownRemaining(WAIT_SECONDS)
    setPhase('idle')
    setCurrentBatch(0)
    setBatchTerminalCount(0)
    setResults([])
    setErrorMsg(null)
    const interval = setInterval(() => {
      setTriggerCooldownRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval)
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(interval)
  }, [open])

  const cancelledRef = useRef(false)
  useEffect(() => {
    if (!open) cancelledRef.current = true
    else cancelledRef.current = false
  }, [open])

  const start = async () => {
    setPhase('running')
    setErrorMsg(null)

    try {
      for (let i = 0; i < batches.length; i++) {
        if (cancelledRef.current) return
        setCurrentBatch(i)
        setBatchTerminalCount(0)

        const created = await adminInitiateDataDeletions(batches[i])
        const ids = created.map((r) => r.id)

        const deadline = Date.now() + POLL_TIMEOUT_MS
        while (true) {
          if (cancelledRef.current) return
          await new Promise((resolve) => setTimeout(resolve, POLL_INTERVAL_MS))
          if (cancelledRef.current) return

          const statuses = await getDataDeletionsStatus(ids)
          const terminal = statuses.filter((r) => isTerminal(r.status))
          setBatchTerminalCount(terminal.length)
          if (terminal.length === ids.length) {
            setResults((prev) => [...prev, ...statuses])
            break
          }
          if (Date.now() >= deadline) {
            throw new Error('Timed out waiting for deletion to complete.')
          }
        }
      }

      setPhase('done')
      queryClient.invalidateQueries({ queryKey: ['privacy', 'admin', 'deletions'] })
    } catch (err) {
      setPhase('error')
      setErrorMsg(err instanceof Error ? err.message : String(err))
    }
  }

  const succeededCount = results.filter((r) => r.status === DeletionRequestStatus.succeeded).length
  const failedCount = results.filter((r) => r.status === DeletionRequestStatus.failed).length

  const handleOpenChange = (next: boolean) => {
    if (!next && (phase === 'idle' || phase === 'done' || phase === 'error')) {
      onClose()
    }
  }

  const canClose = phase === 'idle' || phase === 'done' || phase === 'error'

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='max-w-3xl'>
        <DialogHeader>
          <DialogTitle>Delete Student Data</DialogTitle>
        </DialogHeader>

        {phase === 'idle' && (
          <div className='grid gap-6 lg:grid-cols-2 text-sm'>
            <div className='flex flex-col gap-4'>
              <div className='flex items-center gap-3'>
                <Users className='h-5 w-5 text-muted-foreground' />
                <div>
                  <p className='font-medium'>
                    {studentIDs.length} student{studentIDs.length === 1 ? '' : 's'} selected
                  </p>
                  {studentIDs.length > BATCH_SIZE && (
                    <p className='text-muted-foreground'>
                      Processed in {batches.length} batch{batches.length === 1 ? '' : 'es'} of up to{' '}
                      {BATCH_SIZE}.
                    </p>
                  )}
                </div>
              </div>

              <hr className='border-border' />

              <PrivacyDeletionHowItWorks variant='initiate' />
            </div>

            <div className='flex flex-col gap-4'>
              <PrivacyDeletionWhatGetsDeleted variant='initiate' />
            </div>
          </div>
        )}

        {phase === 'running' && (
          <div className='flex items-center justify-center gap-3 py-12 text-sm text-muted-foreground'>
            <Loader2 className='h-5 w-5 animate-spin' />
            <span>
              Batch {currentBatch + 1} of {batches.length} · {batchTerminalCount} of{' '}
              {batches[currentBatch]?.length ?? 0} done
            </span>
          </div>
        )}

        {phase === 'done' &&
          (failedCount === 0 ? (
            <div className='flex flex-col items-center justify-center gap-2 py-12 text-sm'>
              <CheckCircle2 className='h-10 w-10 text-muted-foreground' />
              <p className='font-medium'>Deletion complete</p>
              <p className='text-muted-foreground'>
                {succeededCount} student{succeededCount === 1 ? '' : 's'} deleted.
              </p>
            </div>
          ) : (
            <div className='flex flex-col items-center justify-center gap-2 py-12 text-sm'>
              <AlertCircle className='h-10 w-10 text-amber-600 dark:text-amber-400' />
              <p className='font-medium'>Deletion completed with errors</p>
              <p className='text-muted-foreground'>
                {succeededCount} succeeded · {failedCount} failed
              </p>
            </div>
          ))}

        {phase === 'error' && (
          <div className='flex flex-col items-center justify-center gap-2 py-12 text-sm'>
            <XCircle className='h-10 w-10 text-destructive' />
            <p className='font-medium'>Deletion failed</p>
            {errorMsg && <p className='text-destructive'>{errorMsg}</p>}
            {(succeededCount > 0 || failedCount > 0) && (
              <p className='text-muted-foreground'>
                {succeededCount} succeeded · {failedCount} failed before the error
              </p>
            )}
          </div>
        )}

        <DialogFooter>
          <Button variant='outline' onClick={onClose} disabled={!canClose}>
            {phase === 'done' || phase === 'error' ? 'Close' : 'Cancel'}
          </Button>
          {phase === 'idle' && (
            <Button
              variant='destructive'
              onClick={start}
              disabled={triggerCooldownRemaining > 0 || studentIDs.length === 0}
            >
              {triggerCooldownRemaining > 0
                ? `Start Deletion (${triggerCooldownRemaining}s)`
                : `Start Deletion`}
            </Button>
          )}
          {phase === 'running' && (
            <Button variant='destructive' disabled>
              Running…
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
