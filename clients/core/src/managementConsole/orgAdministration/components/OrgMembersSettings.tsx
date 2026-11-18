import { ORG_ROLE_LABELS, type Org, type OrgMember, type OrgRole } from '@core/interfaces/org'
import { staffMemberFullName } from '@core/managementConsole/courseUserManagement/interfaces/StaffMember'
import {
  useAddOrgMember,
  useOrgMembers,
  useRemoveOrgMember,
  useSetOrgMemberRole,
} from '@core/network/hooks/useOrgs'
import { Role, useAuthStore } from '@tumaet/prompt-shared-state'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Label,
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
import { Loader2, Trash2, UserPlus } from 'lucide-react'
import { useState } from 'react'
import { AddUserDialog } from '../../shared/components/AddUserDialog'
import { OrgRoleSelect } from './OrgRoleSelect'

interface OrgMembersSettingsProps {
  org: Org
}

// The login column shows the university_login attribute, or the username of users without it.
const loginOf = (member: OrgMember): string => member.universityLogin || member.username

export const OrgMembersSettings = ({ org }: OrgMembersSettingsProps) => {
  const membersQuery = useOrgMembers(org.id)
  const addMember = useAddOrgMember(org.id)
  const setMemberRole = useSetOrgMemberRole(org.id)
  const removeMember = useRemoveOrgMember(org.id)
  const { user, permissions } = useAuthStore()
  const [isAdding, setIsAdding] = useState(false)
  const [newMemberRole, setNewMemberRole] = useState<OrgRole>('Member')
  const [pendingRemoval, setPendingRemoval] = useState<OrgMember | null>(null)

  const isBusy = addMember.isPending || setMemberRole.isPending || removeMember.isPending
  // PROMPT admins manage every org anyway, so only they may change their own membership. The
  // username is compared case-insensitively: the token may carry a display-case username while
  // Keycloak stores it lowercase. The server guards by the token's user ID.
  const isLockedSelf = (member: OrgMember): boolean =>
    !permissions.includes(Role.PROMPT_ADMIN) &&
    user !== undefined &&
    member.username.toLowerCase() === user.username.toLowerCase()
  const roleOf = (userID: string): OrgRole | undefined =>
    membersQuery.data?.find((member) => member.keycloakUserID === userID)?.role

  const closeAddDialog = () => {
    setIsAdding(false)
    // Back to Member, so an Admin grant is never carried over to the next user
    setNewMemberRole('Member')
  }

  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between space-y-0'>
        <div>
          <CardTitle>Members</CardTitle>
          <CardDescription>
            Admins manage the org&apos;s details and its members. Only PROMPT admins can change
            their own membership.
          </CardDescription>
        </div>
        {/* Existing members are marked in the search, so it waits for the member list */}
        <Button size='sm' disabled={!membersQuery.isSuccess} onClick={() => setIsAdding(true)}>
          <UserPlus className='mr-2 h-4 w-4' />
          Add Member
        </Button>
      </CardHeader>
      <CardContent>
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
              <div className='overflow-x-auto'>
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
                      const lockedSelf = isLockedSelf(member)
                      const removeButton = (
                        <Button
                          variant='ghost'
                          size='icon'
                          aria-label={`Remove ${member.username} from the org`}
                          disabled={lockedSelf || isBusy}
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
                              disabled={lockedSelf || isBusy}
                              onChange={(role) => setMemberRole.mutate({ user: member, role })}
                            />
                          </TableCell>
                          <TableCell className='text-right'>
                            {lockedSelf ? (
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
              </div>
            )
          }}
        </QueryGate>
      </CardContent>

      <AddUserDialog
        open={isAdding}
        onOpenChange={(open) => {
          if (!open) closeAddDialog()
        }}
        title='Add Member'
        noteFor={(candidate) => {
          const role = roleOf(candidate.keycloakUserID)
          return role ? `Already in the org as ${ORG_ROLE_LABELS[role]}` : null
        }}
        // A member's role changes in the table, so adding never changes it by accident
        isDisabled={(candidate) => roleOf(candidate.keycloakUserID) !== undefined}
        onSelect={(candidate) =>
          addMember.mutate({ user: candidate, role: newMemberRole }, { onSuccess: closeAddDialog })
        }
        isAdding={addMember.isPending}
      >
        <div className='flex items-center gap-3'>
          <Label htmlFor='org-new-member-role'>Add as</Label>
          <OrgRoleSelect
            id='org-new-member-role'
            value={newMemberRole}
            ariaLabel='Role of the new member'
            className='w-36'
            onChange={setNewMemberRole}
          />
        </div>
      </AddUserDialog>

      <AlertDialog
        open={pendingRemoval !== null}
        onOpenChange={(open) => {
          if (!open) setPendingRemoval(null)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove member?</AlertDialogTitle>
            <AlertDialogDescription>
              Remove {pendingRemoval ? staffMemberFullName(pendingRemoval) : ''} from {org.name}?
              They lose their org role once their session renews, at the latest at their next login.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (pendingRemoval) removeMember.mutate(pendingRemoval.keycloakUserID)
              }}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  )
}
