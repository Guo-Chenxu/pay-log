import type { BillType, Channel } from '../constants/bill'
import client from './client'

export interface ManualBillPayload {
  channel: Channel
  transaction_time: string
  counterparty: string
  description: string
  category: string
  bill_type: BillType
  amount: string
  payment_method: string
  remark: string
  is_investment: boolean
}

export const uploadBill = (file: File, channel: Channel) => {
  const form = new FormData()
  form.append('file', file)
  form.append('channel', String(channel))
  return client.post('/bill/upload', form)
}

export const addManual = (data: ManualBillPayload) => client.post('/bill/manual', data)

export const listBills = (year: number, month: number, page = 1, page_size = 20, sort_order = '') =>
  client.get('/bill/list', { params: { year, month, page, page_size, sort_order: sort_order || undefined } })

export const listRangeBills = (
  startYear: number, startMonth: number,
  endYear: number, endMonth: number,
  page = 1, page_size = 20, sort_order = ''
) =>
  client.get('/bill/list-range', {
    params: { start_year: startYear, start_month: startMonth, end_year: endYear, end_month: endMonth, page, page_size, sort_order: sort_order || undefined },
  })

export const deleteBill = (id: string) => client.delete(`/bill/${id}`)
