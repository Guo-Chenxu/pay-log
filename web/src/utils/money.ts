export const formatYuan = (value?: string) => {
  const normalized = value && value.trim() ? value : '0.00'
  return `¥${normalized}`
}
