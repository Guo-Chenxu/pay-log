import client from './client'

export const triggerAI = (startYear: number, startMonth: number, endYear: number, endMonth: number) =>
  client.post('/ai/trigger', {
    start_year: startYear,
    start_month: startMonth,
    end_year: endYear,
    end_month: endMonth,
  })

export const getAISummary = (startYear: number, startMonth: number, endYear: number, endMonth: number) =>
  client.get('/ai/summary', {
    params: { start_year: startYear, start_month: startMonth, end_year: endYear, end_month: endMonth },
  })
