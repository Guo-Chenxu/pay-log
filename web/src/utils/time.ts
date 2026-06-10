const pad2 = (value: number) => String(value).padStart(2, '0')

export const localDateTimeToRFC3339 = (value: string) => {
  if (!value) return value

  const [datePart, timePart = '00:00'] = value.split('T')
  const [year, month, day] = datePart.split('-').map(Number)
  const [hour = 0, minute = 0, second = 0] = timePart.split(':').map(Number)
  const date = new Date(year, month - 1, day, hour, minute, second, 0)

  const offsetMinutes = -date.getTimezoneOffset()
  const sign = offsetMinutes >= 0 ? '+' : '-'
  const absOffset = Math.abs(offsetMinutes)
  const offsetHour = Math.floor(absOffset / 60)
  const offsetMinute = absOffset % 60

  return `${datePart}T${pad2(hour)}:${pad2(minute)}:${pad2(second)}${sign}${pad2(offsetHour)}:${pad2(offsetMinute)}`
}

export const formatRFC3339Minute = (value?: string) => {
  if (!value) return ''

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value.replace('T', ' ').slice(0, 16)
  }

  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}`
}
