import { PropsWithChildren } from 'react'

type InsideSidebarVisualGroupProps = PropsWithChildren<{
  title: string
}>

export const InsideSidebarVisualGroup = ({ title, children }: InsideSidebarVisualGroupProps) => {
  return (
    <div className='flex flex-col gap-px'>
      <h3 className='uppercase text-xs mt-1 mb-1 ml-2 transform translate-y-1'>{title}</h3>
      {children}
    </div>
  )
}
