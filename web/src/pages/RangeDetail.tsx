import { useState, useCallback, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { getRangeDetail } from '../api/summary'
import { listRangeBills } from '../api/bill'
import { getAISummary, triggerAI } from '../api/ai'
import ManualAddModal from '../components/ManualAddModal'
import { BILL_TYPE_LABEL, CHANNEL_LABEL, billTypeAmountClass } from '../constants/bill'
import { formatYuan } from '../utils/money'
import { formatRFC3339Minute } from '../utils/time'

interface PeriodAgg {
  total_income: string; total_expense: string
  alipay_income: string; alipay_expense: string
  wechat_income: string; wechat_expense: string; investment_amount: string
}
interface MonthSummary {
  id: string; year: number; month: number
  total_income: string; total_expense: string
  alipay_income: string; alipay_expense: string
  wechat_income: string; wechat_expense: string; investment_amount: string
}
interface BillRecord {
  id: string; channel: number; transaction_time: string
  category: string; description: string; bill_type: number; amount: string
}

const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1)
const currentYear = new Date().getFullYear()
const YEARS = Array.from({ length: 5 }, (_, i) => currentYear - i)

export default function RangeDetail() {
  const navigate = useNavigate()
  const [startYear, setStartYear] = useState(currentYear)
  const [startMonth, setStartMonth] = useState(1)
  const [endYear, setEndYear] = useState(currentYear)
  const [endMonth, setEndMonth] = useState(new Date().getMonth() + 1)
  const [page, setPage] = useState(1)
  const [amountSort, setAmountSort] = useState<'asc' | 'desc' | null>(null)
  const pageSize = 20

  const [agg, setAgg] = useState<PeriodAgg | null>(null)
  const [months, setMonths] = useState<MonthSummary[]>([])
  const [bills, setBills] = useState<BillRecord[]>([])
  const [total, setTotal] = useState(0)
  const [aiText, setAiText] = useState('')
  const [triggering, setTriggering] = useState(false)
  const [showModal, setShowModal] = useState(false)
  const [loaded, setLoaded] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  // track current query range so loadBills can reference it independently
  const rangeRef = useRef({ startYear, startMonth, endYear, endMonth })

  const stopPolling = () => {
    if (pollRef.current) { clearInterval(pollRef.current); pollRef.current = null }
  }

  const loadMeta = useCallback((sy: number, sm: number, ey: number, em: number) => {
    getRangeDetail(sy, sm, ey, em)
      .then(r => {
        const d = r.data.data
        setAgg(d.agg)
        setMonths(d.months || [])
      })
      .catch(() => {})
    getAISummary(sy, sm, ey, em)
      .then(r => setAiText(r.data.data?.ai_analysis || ''))
      .catch(() => setAiText(''))
  }, [])

  const loadBills = useCallback((sy: number, sm: number, ey: number, em: number, pg: number, sortOrder: string) => {
    listRangeBills(sy, sm, ey, em, pg, pageSize, sortOrder)
      .then(r => {
        setBills(r.data.data?.items || [])
        setTotal(Number(r.data.data?.total || 0))
      })
      .catch(() => {})
  }, [])

  const handleQuery = () => {
    rangeRef.current = { startYear, startMonth, endYear, endMonth }
    setPage(1)
    setAmountSort(null)
    setLoaded(true)
    stopPolling()
    loadMeta(startYear, startMonth, endYear, endMonth)
    loadBills(startYear, startMonth, endYear, endMonth, 1, '')
  }

  // Only reload bills when page or sort changes (not meta)
  useEffect(() => {
    if (!loaded) return
    const { startYear: sy, startMonth: sm, endYear: ey, endMonth: em } = rangeRef.current
    loadBills(sy, sm, ey, em, page, amountSort || '')
  }, [page, amountSort, loaded, loadBills])

  useEffect(() => () => stopPolling(), [])

  const handleTriggerAI = async () => {
    setTriggering(true)
    const prevText = aiText
    const { startYear: sy, startMonth: sm, endYear: ey, endMonth: em } = rangeRef.current
    try {
      await triggerAI(sy, sm, ey, em)
    } catch (e: unknown) { setAiText('AI 分析失败：' + (e instanceof Error ? e.message : '')); setTriggering(false); return }
    let elapsed = 0
    stopPolling()
    pollRef.current = setInterval(async () => {
      elapsed += 10
      try {
        const r = await getAISummary(sy, sm, ey, em)
        const text = r.data.data?.ai_analysis || ''
        if (text && text !== prevText) { setAiText(text); setTriggering(false); stopPolling(); return }
      } catch { /* ignore polling errors */ }
      if (elapsed >= 60) { setTriggering(false); stopPolling() }
    }, 10000)
  }

  const aggCards = agg ? [
    { label: '总支出', val: agg.total_expense, color: 'text-red-500' },
    { label: '总收入', val: agg.total_income, color: 'text-green-600' },
    { label: '支付宝支出', val: agg.alipay_expense, color: 'text-orange-500' },
    { label: '支付宝收入', val: agg.alipay_income, color: 'text-orange-400' },
    { label: '微信支出', val: agg.wechat_expense, color: 'text-gray-700' },
    { label: '微信收入', val: agg.wechat_income, color: 'text-gray-500' },
    { label: '理财支出', val: agg.investment_amount, color: 'text-purple-600' },
  ] : []

  const selectCls = "border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-400"

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center gap-4">
          <button onClick={() => navigate('/')} className="text-sm text-gray-500 hover:text-gray-700">← 返回</button>
          <h2 className="font-semibold text-gray-700">多月账单汇总</h2>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-6 py-6 space-y-6">

        {/* Date range selector */}
        <div className="bg-white rounded-xl border border-gray-200 p-5 flex flex-wrap items-center gap-3">
          <span className="text-sm text-gray-500">起始：</span>
          <select className={selectCls} value={startYear} onChange={e => setStartYear(Number(e.target.value))}>
            {YEARS.map(y => <option key={y} value={y}>{y}年</option>)}
          </select>
          <select className={selectCls} value={startMonth} onChange={e => setStartMonth(Number(e.target.value))}>
            {MONTHS.map(m => <option key={m} value={m}>{m}月</option>)}
          </select>
          <span className="text-sm text-gray-500 ml-2">结束：</span>
          <select className={selectCls} value={endYear} onChange={e => setEndYear(Number(e.target.value))}>
            {YEARS.map(y => <option key={y} value={y}>{y}年</option>)}
          </select>
          <select className={selectCls} value={endMonth} onChange={e => setEndMonth(Number(e.target.value))}>
            {MONTHS.map(m => <option key={m} value={m}>{m}月</option>)}
          </select>
          <button onClick={handleQuery} className="ml-2 px-5 py-2 text-sm bg-green-500 hover:bg-green-600 text-white rounded-lg transition-colors">查询</button>
        </div>

        {/* Aggregate stat cards */}
        {loaded && aggCards.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3">
            {aggCards.map(({ label, val, color }) => (
              <div key={label} className="bg-white rounded-xl border border-gray-200 p-4">
                <p className="text-xs text-gray-400 mb-1">{label}</p>
                <p className={`text-base font-bold ${color}`}>{formatYuan(val)}</p>
              </div>
            ))}
          </div>
        )}

        {/* Per-month breakdown table */}
        {loaded && months.length > 0 && (
          <div className="bg-white rounded-xl border border-gray-200 overflow-x-auto">
            <div className="px-5 py-3 border-b border-gray-100">
              <span className="font-medium text-gray-700 text-sm">月度明细</span>
            </div>
            <table className="w-full text-sm border-collapse">
              <thead className="bg-gray-50 text-xs text-gray-500">
                <tr>
                  {['月份','总支出','总收入','支付宝支出','支付宝收入','微信支出','微信收入','理财支出'].map(h => (
                    <th key={h} className="px-4 py-3 text-left font-medium whitespace-nowrap border border-gray-200">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {months.map(mo => (
                  <tr key={mo.id || `${mo.year}-${mo.month}`} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/month/${mo.year}/${mo.month}`)}>
                    <td className="px-4 py-3 font-medium text-gray-700 whitespace-nowrap border border-gray-200">{mo.year}年{mo.month}月</td>
                    <td className="px-4 py-3 text-red-500 border border-gray-200">{formatYuan(mo.total_expense)}</td>
                    <td className="px-4 py-3 text-green-600 border border-gray-200">{formatYuan(mo.total_income)}</td>
                    <td className="px-4 py-3 text-orange-500 border border-gray-200">{formatYuan(mo.alipay_expense)}</td>
                    <td className="px-4 py-3 text-orange-400 border border-gray-200">{formatYuan(mo.alipay_income)}</td>
                    <td className="px-4 py-3 text-gray-700 border border-gray-200">{formatYuan(mo.wechat_expense)}</td>
                    <td className="px-4 py-3 text-gray-500 border border-gray-200">{formatYuan(mo.wechat_income)}</td>
                    <td className="px-4 py-3 text-purple-600 border border-gray-200">{formatYuan(mo.investment_amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {loaded && months.length === 0 && (
          <div className="bg-white rounded-xl border border-gray-200 p-8 text-center text-gray-400 text-sm">暂无月度汇总数据（请先上传账单）</div>
        )}

        {/* AI Analysis */}
        {loaded && (
          <div className="bg-white rounded-xl border border-gray-200 p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="font-medium text-gray-700 text-sm">AI 分析</span>
              <button onClick={handleTriggerAI} disabled={triggering} className="text-xs px-3 py-1.5 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50">
                {triggering ? '分析中，每10秒轮询...' : '触发 AI 分析'}
              </button>
            </div>
            {aiText
              ? <div className="prose prose-sm max-w-none text-gray-600"><ReactMarkdown remarkPlugins={[remarkGfm]}>{aiText}</ReactMarkdown></div>
              : <p className="text-sm text-gray-400">暂无 AI 分析，点击右侧按钮生成</p>
            }
          </div>
        )}

        {/* Bill list */}
        {loaded && (
          <div>
            <div className="flex items-center justify-between mb-3">
              <span className="font-medium text-gray-700 text-sm">明细列表（共 {total} 条）</span>
              <button onClick={() => setShowModal(true)} className="text-xs px-3 py-1.5 bg-green-500 text-white rounded-lg hover:bg-green-600">+ 手动添加</button>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-xs text-gray-500 border-b border-gray-200">
                  <tr>
                    {['时间','渠道','类别','说明','类型'].map(h => <th key={h} className="px-4 py-3 text-left font-medium">{h}</th>)}
                    <th className="px-4 py-3 text-left font-medium">
                      <button onClick={() => { setAmountSort(s => s === 'desc' ? 'asc' : 'desc'); setPage(1) }} className="flex items-center gap-1 hover:text-gray-700">
                        金额 {amountSort === 'desc' ? '↓' : amountSort === 'asc' ? '↑' : '↕'}
                      </button>
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {bills.map(b => (
                    <tr key={b.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-gray-600 whitespace-nowrap">{formatRFC3339Minute(b.transaction_time)}</td>
                      <td className="px-4 py-3 text-gray-600">{CHANNEL_LABEL[b.channel] || '未知渠道'}</td>
                      <td className="px-4 py-3 text-gray-600">{b.category}</td>
                      <td className="px-4 py-3 text-gray-600 max-w-xs truncate">{b.description}</td>
                      <td className="px-4 py-3 text-gray-600">{BILL_TYPE_LABEL[b.bill_type] || '-'}</td>
                      <td className={`px-4 py-3 font-medium ${billTypeAmountClass(b.bill_type)}`}>{formatYuan(b.amount)}</td>
                    </tr>
                  ))}
                  {bills.length === 0 && <tr><td colSpan={6} className="text-center py-10 text-gray-400">暂无明细</td></tr>}
                </tbody>
              </table>
            </div>
            {total > pageSize && (
              <div className="mt-4 flex items-center justify-center gap-4">
                <button disabled={page<=1} onClick={() => setPage(p=>p-1)} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-40 hover:bg-gray-50">上一页</button>
                <span className="text-sm text-gray-500">{page} / {Math.ceil(total/pageSize)}</span>
                <button disabled={page*pageSize>=total} onClick={() => setPage(p=>p+1)} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-40 hover:bg-gray-50">下一页</button>
              </div>
            )}
          </div>
        )}
      </main>

      {showModal && (
        <ManualAddModal onClose={() => setShowModal(false)} onSuccess={() => {
          setShowModal(false)
          const { startYear: sy, startMonth: sm, endYear: ey, endMonth: em } = rangeRef.current
          loadBills(sy, sm, ey, em, page, amountSort || '')
        }} />
      )}
    </div>
  )
}
