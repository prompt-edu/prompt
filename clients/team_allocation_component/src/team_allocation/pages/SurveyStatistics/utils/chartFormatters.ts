export const truncate = (value: string, maxLength = 12) =>
  value.length > maxLength ? `${value.slice(0, maxLength - 1)}…` : value

export const commonWordPrefix = (values: string[]): string => {
  if (values.length < 2) return ''
  let prefix = values[0]
  for (const value of values.slice(1)) {
    while (prefix && !value.startsWith(prefix)) {
      prefix = prefix.slice(0, -1)
    }
    if (!prefix) return ''
  }
  const lastSpace = prefix.lastIndexOf(' ')
  if (lastSpace < 0) return ''
  const wordPrefix = prefix.slice(0, lastSpace + 1)
  return values.every((value) => value.length > wordPrefix.length) ? wordPrefix : ''
}

export const ordinal = (n: number) => {
  const mod100 = n % 100
  if (mod100 >= 11 && mod100 <= 13) return `${n}th`
  switch (n % 10) {
    case 1:
      return `${n}st`
    case 2:
      return `${n}nd`
    case 3:
      return `${n}rd`
    default:
      return `${n}th`
  }
}
