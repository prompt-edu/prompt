import type { PropsWithChildren } from 'react'

type InsideSidebarVisualGroupProps = PropsWithChildren<{
  title: string
}>

export const InsideSidebarVisualGroup = ({ title, children }: InsideSidebarVisualGroupProps) => {
  return (
    <div className='flex flex-col gap-px'>
      <h3 className='uppercase text-xs mb-1 ml-2'>{title}</h3>
      {children}
    </div>
  )
}
