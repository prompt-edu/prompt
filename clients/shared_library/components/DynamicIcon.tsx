import React, { lazy, Suspense } from 'react'
import { LucideProps } from 'lucide-react'
import dynamicIconImports from 'lucide-react/dynamicIconImports'

const fallback = <div style={{ width: 24, height: 24 }} />

interface IconProps extends Omit<LucideProps, 'ref'> {
  name: string
}

const iconCache = new Map<string, React.LazyExoticComponent<React.ComponentType<LucideProps>>>()

const DynamicIcon = ({ name, ...props }: IconProps) => {
  if (!dynamicIconImports[name]) {
    console.error(`Icon "${name}" does not exist in dynamicIconImports`)
    return fallback
  }

  let LucideIcon = iconCache.get(name)
  if (!LucideIcon) {
    LucideIcon = lazy(dynamicIconImports[name])
    iconCache.set(name, LucideIcon)
  }

  return (
    <Suspense fallback={fallback}>
      <LucideIcon {...props} />
    </Suspense>
  )
}

export default DynamicIcon
