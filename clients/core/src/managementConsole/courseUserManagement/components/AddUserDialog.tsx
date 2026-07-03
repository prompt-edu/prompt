import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  Input,
} from '@tumaet/prompt-ui-components'
import { Loader2, Search } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useSearchKeycloakUsers } from '../hooks/useSearchKeycloakUsers'
import {
  type CourseGroupName,
  type StaffMember,
  staffMemberFullName,
} from '../interfaces/StaffMember'

interface AddUserDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  groupName: CourseGroupName
  existingLecturerIDs: Set<string>
  existingEditorIDs: Set<string>
  onSelect: (user: StaffMember) => void
  isAdding: boolean
}

const useDebouncedValue = <T,>(value: T, delayMs = 300): T => {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delayMs)
    return () => clearTimeout(id)
  }, [value, delayMs])
  return debounced
}

export const AddUserDialog = ({
  open,
  onOpenChange,
  groupName,
  existingLecturerIDs,
  existingEditorIDs,
  onSelect,
  isAdding,
}: AddUserDialogProps) => {
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebouncedValue(query, 300)
  const { data, isFetching, error } = useSearchKeycloakUsers(debouncedQuery)

  useEffect(() => {
    if (!open) setQuery('')
  }, [open])

  const existingForThisRole = useMemo(
    () => (groupName === 'Lecturer' ? existingLecturerIDs : existingEditorIDs),
    [groupName, existingLecturerIDs, existingEditorIDs],
  )

  const membershipLabel = (userID: string): string | null => {
    const inThis = existingForThisRole.has(userID)
    if (inThis) return `Already a ${groupName}`
    if (groupName === 'Lecturer' && existingEditorIDs.has(userID)) return 'Currently an Editor'
    if (groupName === 'Editor' && existingLecturerIDs.has(userID)) return 'Currently a Lecturer'
    return null
  }

  const trimmed = debouncedQuery.trim()
  const showHint = trimmed.length > 0 && trimmed.length < 2

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-lg'>
        <DialogHeader>
          <DialogTitle>Add {groupName}</DialogTitle>
          <DialogDescription>
            Search for a Keycloak user by name, email, or username.
          </DialogDescription>
        </DialogHeader>

        <div className='relative w-full'>
          <Search className='pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
          <Input
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder='Type at least 2 characters...'
            className='w-full pl-10'
          />
        </div>

        <div className='min-h-[12rem] max-h-80 overflow-y-auto rounded-md border'>
          {showHint && (
            <p className='p-4 text-sm text-muted-foreground'>
              Type at least 2 characters to search.
            </p>
          )}
          {!showHint && isFetching && (
            <div className='flex items-center justify-center p-6'>
              <Loader2 className='h-5 w-5 animate-spin text-muted-foreground' />
            </div>
          )}
          {error && <p className='p-4 text-sm text-destructive'>Search failed. Try again.</p>}
          {!isFetching && !error && data && data.results.length === 0 && trimmed.length >= 2 && (
            <p className='p-4 text-sm text-muted-foreground'>No users found.</p>
          )}
          {!isFetching && trimmed.length >= 2 && data && data.results.length > 0 && (
            <ul className='divide-y'>
              {data.results.map((user) => {
                const note = membershipLabel(user.keycloakUserID)
                const disabled = existingForThisRole.has(user.keycloakUserID) || isAdding
                return (
                  <li key={user.keycloakUserID} className='flex items-center justify-between p-3'>
                    <div className='min-w-0'>
                      <p className='truncate font-medium'>{staffMemberFullName(user)}</p>
                      <p className='truncate text-xs text-muted-foreground'>
                        {user.username}
                        {user.email ? ` · ${user.email}` : ''}
                      </p>
                      {note && <p className='mt-1 text-xs text-muted-foreground italic'>{note}</p>}
                    </div>
                    <Button size='sm' disabled={disabled} onClick={() => onSelect(user)}>
                      Add
                    </Button>
                  </li>
                )
              })}
            </ul>
          )}
        </div>

        {trimmed.length >= 2 && data?.truncated && (
          <p className='text-xs text-muted-foreground'>
            More results available - refine your search to narrow them down.
          </p>
        )}
      </DialogContent>
    </Dialog>
  )
}
