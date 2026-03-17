import DynamicIcon from '@/components/DynamicIcon'
import { useMemo } from 'react'

export const CourseAvatar = ({
  bgColor,
  iconName,
  className,
  isActive = true,
}: {
  bgColor: string
  iconName: string
  className?: string
  isActive?: boolean
}) => {
  const memoizedIcon = useMemo(() => <DynamicIcon name={iconName} className='size-6' />, [iconName])

  return (
    <div
      className={`
            relative flex aspect-square items-center justify-center rounded-lg text-gray-800
            ${isActive ? 'size-12' : 'size-10'} ${bgColor} ${className}
          `}
    >
      <div className='size-6'>{memoizedIcon}</div>
    </div>
  )
}
