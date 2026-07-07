import type { Team } from '@tumaet/prompt-shared-state'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { GraduationCap, LogOut, Trash2, UserPlus, Users } from 'lucide-react'
import type React from 'react'

export interface TeamCardProps {
  team: Team
  isMember: boolean
  full: boolean
  isLecturer: boolean
  joiningDisabled: boolean
  disabled: boolean
  courseParticipationID?: string
  onJoin: (teamId: string) => void
  onLeave: (teamId: string) => void
  onDelete: (team: Team) => void
}

export const TeamCard: React.FC<TeamCardProps> = ({
  team,
  isMember,
  full,
  isLecturer,
  joiningDisabled,
  disabled,
  courseParticipationID,
  onJoin,
  onLeave,
  onDelete,
}) => (
  <Card
    data-testid={`team-card-${team.name}`}
    className={`overflow-hidden transition-all duration-200 flex flex-col ${
      isMember ? 'ring-2 ring-primary shadow-md' : ''
    }`}
  >
    <CardHeader className='pb-3'>
      <CardTitle
        className={`font-semibold truncate pb-2 ${team.name.length > 20 ? 'text-base' : ''}`}
      >
        {team.name}
      </CardTitle>
      <CardDescription className='flex items-center'>
        <Badge variant={full ? 'destructive' : 'secondary'} className='whitespace-nowrap text-sm'>
          {team.members.length}/3 Members
        </Badge>
      </CardDescription>
    </CardHeader>

    <CardContent className='pb-2 flex-1 space-y-3'>
      {/* Members Section */}
      <div>
        <p className='text-sm font-medium mb-2'>Members:</p>
        {team.members.length > 0 ? (
          <ul className='space-y-1.5 max-h-[100px] overflow-y-auto'>
            {team.members.map((m, idx) => {
              const isCurrent = courseParticipationID && m.id === courseParticipationID
              return (
                <li
                  key={idx}
                  className={`flex items-center gap-2 p-1 rounded-md text-sm ${
                    isCurrent ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'
                  }`}
                >
                  <Users size={14} className={isCurrent ? 'text-primary' : ''} />
                  <span className='truncate'>
                    {m.firstName} {m.lastName}
                  </span>
                  {isCurrent && (
                    <Badge variant='outline' className='ml-auto text-xs'>
                      You
                    </Badge>
                  )}
                </li>
              )
            })}
          </ul>
        ) : (
          <p className='text-sm text-muted-foreground italic'>No members yet</p>
        )}
      </div>

      {/* Tutors Section */}
      {team.tutors && team.tutors.length > 0 && (
        <div>
          <p className='text-sm font-medium mb-2'>Tutors:</p>
          <ul className='space-y-1.5 max-h-[80px] overflow-y-auto'>
            {team.tutors.map((tutor, idx) => (
              <li
                key={idx}
                className='flex items-center gap-2 p-1 rounded-md text-sm text-muted-foreground'
              >
                <GraduationCap size={14} />
                <span className='truncate'>
                  {tutor.firstName} {tutor.lastName}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </CardContent>

    <CardFooter className='pt-2 pb-4 space-y-2'>
      {!isLecturer &&
        (isMember ? (
          <Button
            variant='destructive'
            className='w-full'
            onClick={() => onLeave(team.id)}
            disabled={disabled}
          >
            <LogOut className='mr-2 h-4 w-4' />
            Leave Team
          </Button>
        ) : (
          <Button
            className='w-full'
            onClick={() => onJoin(team.id)}
            disabled={joiningDisabled || full || disabled}
            variant={full ? 'outline' : 'default'}
          >
            <UserPlus className='mr-2 h-4 w-4' />
            {full ? 'Team Full' : 'Join Team'}
          </Button>
        ))}

      {isLecturer && (
        <Button
          variant='destructive'
          className='w-full flex items-center justify-center gap-2'
          onClick={() => onDelete(team)}
        >
          <Trash2 className='h-4 w-4' />
          Delete Team
        </Button>
      )}
    </CardFooter>
  </Card>
)
