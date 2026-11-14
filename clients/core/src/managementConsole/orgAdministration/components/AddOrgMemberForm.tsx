import type { OrgRole, SetOrgMember } from '@core/interfaces/org'
import { Button, Input, Label } from '@tumaet/prompt-ui-components'
import { UserPlus } from 'lucide-react'
import { useState } from 'react'
import { validateUniversityLogin } from '../utils/validateUniversityLogin'
import { OrgRoleSelect } from './OrgRoleSelect'

interface AddOrgMemberFormProps {
  disabled: boolean
  onAdd: (member: SetOrgMember, onAdded: () => void) => void
}

export const AddOrgMemberForm = ({ disabled, onAdd }: AddOrgMemberFormProps) => {
  const [universityLogin, setUniversityLogin] = useState('')
  const [role, setRole] = useState<OrgRole>('Member')

  const login = universityLogin.trim()
  // An empty field is the resting state, not an error
  const error = login === '' ? undefined : validateUniversityLogin(login)
  const canAdd = login !== '' && error === undefined && !disabled

  return (
    <form
      className='flex flex-col gap-1.5'
      onSubmit={(event) => {
        event.preventDefault()
        if (canAdd) {
          // Back to Member after each add, so an Admin grant is never carried over to the next user
          onAdd({ universityLogin: login, role }, () => {
            setUniversityLogin('')
            setRole('Member')
          })
        }
      }}
    >
      <Label htmlFor='org-member-login'>Add a member</Label>
      <div className='flex flex-col gap-2 sm:flex-row'>
        <Input
          id='org-member-login'
          placeholder='University login, e.g. ab12cde'
          value={universityLogin}
          disabled={disabled}
          aria-invalid={error !== undefined}
          onChange={(event) => setUniversityLogin(event.target.value)}
        />
        <OrgRoleSelect
          value={role}
          ariaLabel='Role of the new member'
          disabled={disabled}
          className='sm:w-36'
          onChange={setRole}
        />
        <Button type='submit' disabled={!canAdd}>
          <UserPlus className='mr-2 h-4 w-4' />
          Add
        </Button>
      </div>
      {error && <p className='text-sm text-destructive'>{error}</p>}
      <p className='text-sm text-muted-foreground'>
        Adding an existing member changes their role. A change applies once the user&apos;s session
        renews, at the latest at their next login.
      </p>
    </form>
  )
}
