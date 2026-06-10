import client from './client'

export const getOverview = (page = 1, page_size = 12) =>
  client.get('/summary/overview', { params: { page, page_size } })

export const getMonthDetail = (year: number, month: number) =>
  client.get('/summary/month', { params: { year, month } })

export const getRangeDetail = (
  startYear: number, startMonth: number,
  endYear: number, endMonth: number
) =>
  client.get('/summary/range', {
    params: { start_year: startYear, start_month: startMonth, end_year: endYear, end_month: endMonth },
  })
