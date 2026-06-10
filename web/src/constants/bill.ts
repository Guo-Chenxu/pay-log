export const CHANNEL = {
  ALIPAY: 1,
  WECHAT: 2,
} as const

export type Channel = typeof CHANNEL[keyof typeof CHANNEL]

export const BILL_TYPE = {
  INCOME: 1,
  EXPENSE: 2,
  NEUTRAL: 3,
} as const

export type BillType = typeof BILL_TYPE[keyof typeof BILL_TYPE]

export const CHANNEL_LABEL: Record<number, string> = {
  [CHANNEL.ALIPAY]: '支付宝',
  [CHANNEL.WECHAT]: '微信',
}

export const BILL_TYPE_LABEL: Record<number, string> = {
  [BILL_TYPE.INCOME]: '收入',
  [BILL_TYPE.EXPENSE]: '支出',
  [BILL_TYPE.NEUTRAL]: '不计收支',
}

export const CHANNEL_ACCEPT: Record<number, string> = {
  [CHANNEL.ALIPAY]: '.csv',
  [CHANNEL.WECHAT]: '.xlsx',
}

export const billTypeAmountClass = (billType: number) => {
  if (billType === BILL_TYPE.EXPENSE) return 'text-red-500'
  if (billType === BILL_TYPE.INCOME) return 'text-green-600'
  return 'text-gray-400'
}
