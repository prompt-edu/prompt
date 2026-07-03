import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  DeleteConfirmation,
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
import { Trash2, UserPlus } from 'lucide-react'
import { useState } from 'react'
import {
  type CourseGroupName,
  type StaffMember,
  staffMemberFullName,
} from '../interfaces/StaffMember'

interface StaffMemberTableProps {
  title: string
  description: string
  groupName: CourseGroupName
  members: StaffMember[]
  currentUsername: string | undefined
  onAdd: () => void
  onRemove: (member: StaffMember) => void
  isRemoving: boolean
}

export const StaffMemberTable = ({
  title,
  description,
  groupName,
  members,
  currentUsername,
  onAdd,
  onRemove,
  isRemoving,
}: StaffMemberTableProps) => {
  const [pendingRemoval, setPendingRemoval] = useState<StaffMember | null>(null)

  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between space-y-0'>
        <div>
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </div>
        <Button onClick={onAdd} size='sm'>
          <UserPlus className='mr-2 h-4 w-4' />
          Add {groupName}
        </Button>
      </CardHeader>
      <CardContent>
        {members.length === 0 ? (
          <p className='py-6 text-center text-sm text-muted-foreground'>
            No {groupName.toLowerCase()}s yet.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Username</TableHead>
                <TableHead>Email</TableHead>
                <TableHead className='w-12 text-right'>
                  <span className='sr-only'>Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.map((member) => {
                // Case-insensitive: the JWT may carry a display-case username while
                // Keycloak stores usernames lowercase. The server still defends in
                // depth using the JWT sub, so this is just for the UI guard.
                const isSelf =
                  currentUsername !== undefined &&
                  member.username.toLowerCase() === currentUsername.toLowerCase()
                const removeButton = (
                  <Button
                    variant='ghost'
                    size='icon'
                    aria-label={`Remove ${member.username} from ${groupName}`}
                    disabled={isSelf || isRemoving}
                    onClick={() => setPendingRemoval(member)}
                  >
                    <Trash2 className='h-4 w-4' />
                  </Button>
                )
                return (
                  <TableRow key={member.keycloakUserID}>
                    <TableCell className='font-medium'>{staffMemberFullName(member)}</TableCell>
                    <TableCell>{member.username}</TableCell>
                    <TableCell>{member.email}</TableCell>
                    <TableCell className='text-right'>
                      {isSelf ? (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span>{removeButton}</span>
                            </TooltipTrigger>
                            <TooltipContent>
                              You cannot remove yourself. Ask another instructor or use the Keycloak
                              admin console.
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
        )}
      </CardContent>

      <DeleteConfirmation
        isOpen={pendingRemoval !== null}
        setOpen={(open) => {
          if (!open) setPendingRemoval(null)
        }}
        deleteMessage={`Remove ${pendingRemoval ? staffMemberFullName(pendingRemoval) : ''} from ${groupName}?`}
        onClick={(confirmed) => {
          if (confirmed && pendingRemoval) onRemove(pendingRemoval)
          setPendingRemoval(null)
        }}
      />
    </Card>
  )
}
