import type { Org, OrgMember } from '@core/interfaces/org'
import { staffMemberFullName } from '@core/managementConsole/courseUserManagement/interfaces/StaffMember'
import {
  useOrgMembers,
  useRemoveOrgMember,
  useSetOrgMember,
  useSetOrgMemberRole,
} from '@core/network/hooks/useOrgs'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  DeleteConfirmation,
  ProfilePicture,
  QueryGate,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
import { Loader2, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { AddOrgMemberForm } from './AddOrgMemberForm'
import { OrgRoleSelect } from './OrgRoleSelect'

interface OrgMembersSettingsProps {
  org: Org
}

// The login column shows the university_login attribute, or the username of users without it, as
// the server matches an added login against both. Role changes address the user by Keycloak ID.
const loginOf = (member: OrgMember): string => member.universityLogin || member.username

export const OrgMembersSettings = ({ org }: OrgMembersSettingsProps) => {
  const membersQuery = useOrgMembers(org.id)
  const setMember = useSetOrgMember(org.id)
  const setMemberRole = useSetOrgMemberRole(org.id)
  const removeMember = useRemoveOrgMember(org.id)
  const { user } = useAuthStore()
  const [pendingRemoval, setPendingRemoval] = useState<OrgMember | null>(null)

  const isBusy = setMember.isPending || setMemberRole.isPending || removeMember.isPending
  // Case-insensitive: the token may carry a display-case username while Keycloak stores it
  // lowercase. The server guards against own-membership changes by the token's user ID.
  const isSelf = (member: OrgMember): boolean =>
    user !== undefined && member.username.toLowerCase() === user.username.toLowerCase()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Members</CardTitle>
        <CardDescription>
          Admins manage the org&apos;s details and its members. Nobody can change their own
          membership.
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-6'>
        <AddOrgMemberForm
          disabled={isBusy}
          onAdd={(member, onAdded) => setMember.mutate(member, { onSuccess: onAdded })}
        />
        <QueryGate
          queries={[membersQuery]}
          loadingFallback={
            <div className='flex justify-center py-6'>
              <Loader2 className='h-5 w-5 animate-spin text-muted-foreground' />
            </div>
          }
          errorFallback={({ refetch }) => (
            <div className='flex flex-col items-center gap-2 py-6 text-center text-sm'>
              <p className='text-destructive'>The members could not be loaded.</p>
              <Button variant='outline' size='sm' onClick={() => refetch()}>
                Retry
              </Button>
            </div>
          )}
        >
          {() => {
            const members = membersQuery.data ?? []
            if (members.length === 0) {
              return (
                <p className='py-6 text-center text-sm text-muted-foreground'>No members yet.</p>
              )
            }
            return (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>University Login</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead className='w-40'>Role</TableHead>
                    <TableHead className='w-12 text-right'>
                      <span className='sr-only'>Actions</span>
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {members.map((member) => {
                    const self = isSelf(member)
                    const removeButton = (
                      <Button
                        variant='ghost'
                        size='icon'
                        aria-label={`Remove ${member.username} from the org`}
                        disabled={self || isBusy}
                        onClick={() => setPendingRemoval(member)}
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    )
                    return (
                      <TableRow key={member.keycloakUserID}>
                        <TableCell className='font-medium'>
                          <div className='flex items-center gap-2'>
                            <ProfilePicture
                              userId={member.keycloakUserID}
                              firstName={member.firstName}
                              lastName={member.lastName}
                              size='sm'
                            />
                            {staffMemberFullName(member)}
                          </div>
                        </TableCell>
                        <TableCell>{loginOf(member)}</TableCell>
                        <TableCell>{member.email}</TableCell>
                        <TableCell>
                          <OrgRoleSelect
                            value={member.role}
                            ariaLabel={`Role of ${member.username}`}
                            disabled={self || isBusy}
                            onChange={(role) => setMemberRole.mutate({ member, role })}
                          />
                        </TableCell>
                        <TableCell className='text-right'>
                          {self ? (
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <span>{removeButton}</span>
                                </TooltipTrigger>
                                <TooltipContent>
                                  You cannot change your own membership. Ask another Admin of the
                                  org or a PROMPT admin.
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          ) : (
                            removeButton
                          )}
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            )
          }}
        </QueryGate>
      </CardContent>

      <DeleteConfirmation
        isOpen={pendingRemoval !== null}
        setOpen={(open) => {
          if (!open) setPendingRemoval(null)
        }}
        deleteMessage={`Remove ${pendingRemoval ? staffMemberFullName(pendingRemoval) : ''} from ${org.name}?`}
        onClick={(confirmed) => {
          if (confirmed && pendingRemoval) removeMember.mutate(pendingRemoval.keycloakUserID)
          setPendingRemoval(null)
        }}
      />
    </Card>
  )
}
