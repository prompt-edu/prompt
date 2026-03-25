import React from 'react'
import { LucideProps } from 'lucide-react'
import * as LucideIcons from 'lucide-react'

const fallback = <div style={{ width: 24, height: 24 }} />

interface IconProps extends Omit<LucideProps, 'ref'> {
  name: string
}

type LucideIconComponent = React.ComponentType<LucideProps>

const iconRegistry = new Map<string, LucideIconComponent>()

const registerIcon = (name: string, icon: LucideIconComponent) => {
  iconRegistry.set(name, icon)
  iconRegistry.set(name.toLowerCase(), icon)
}

const toKebabCase = (value: string) => value.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase()

const toPascalCase = (value: string) =>
  value
    .split(/[-_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join('')

for (const [exportName, icon] of Object.entries(LucideIcons)) {
  if (typeof icon !== 'function') {
    continue
  }

  registerIcon(exportName, icon as LucideIconComponent)
  registerIcon(toKebabCase(exportName), icon as LucideIconComponent)
}

const DynamicIcon = ({ name, ...props }: IconProps) => {
  const normalizedName = name.trim()

  const LucideIcon =
    iconRegistry.get(normalizedName) ??
    iconRegistry.get(normalizedName.toLowerCase()) ??
    iconRegistry.get(toPascalCase(normalizedName))

  if (!LucideIcon) {
    console.error(`Icon "${name}" does not exist in lucide-react`)
    return fallback
  }

  return <LucideIcon {...props} />
}

export default DynamicIcon
