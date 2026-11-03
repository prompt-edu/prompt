import { useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  getStudentName,
  useOwnProfilePicture,
} from '@tumaet/prompt-ui-components'

interface NavAvatarProps {
  avatarOnly?: boolean
}

export function NavAvatar({ avatarOnly = false }: NavAvatarProps) {
  const { user } = useAuthStore()
  const { data: ownPicture } = useOwnProfilePicture()

  const userName = getStudentName(user ?? {}) || 'Unknown User'
  const userEmail = user?.email || 'Unknown Email'
  const initials =
    [user?.firstName, user?.lastName]
      .map((part) => part?.trim().charAt(0))
      .filter(Boolean)
      .join('') || '??'

  return (
    <>
      <Avatar className='h-10 w-10 rounded-lg'>
        {ownPicture && <AvatarImage src={ownPicture.url} alt={userName} className='object-cover' />}
        <AvatarFallback className='rounded-lg'>{initials}</AvatarFallback>
      </Avatar>
      {!avatarOnly && (
        <div className='grid flex-1 text-left text-sm leading-tight'>
          <span className='truncate font-semibold'>{userName}</span>
          <span className='truncate text-xs'>{userEmail}</span>
        </div>
      )}
    </>
  )
}
