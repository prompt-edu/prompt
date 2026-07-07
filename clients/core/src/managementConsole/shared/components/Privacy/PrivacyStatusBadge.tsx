interface PrivacyStatusBadgeProps {
  label: string
  icon: React.ReactNode
  colorClass: string
}

export function PrivacyStatusBadge({ label, icon, colorClass }: PrivacyStatusBadgeProps) {
  return (
    <span
      className={
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ' +
        colorClass
      }
    >
      {icon}
      {label}
    </span>
  )
}
