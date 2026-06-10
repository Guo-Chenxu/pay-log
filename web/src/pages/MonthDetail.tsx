import { useEffect, useState, useCallback, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { listBills, deleteBill } from '../api/bill'
import { getMonthDetail } from '../api/summary'
import { triggerAI, getAISummary } from '../api/ai'
import ManualAddModal from '../components/ManualAddModal'
import { BILL_TYPE_LABEL, CHANNEL_LABEL, billTypeAmountClass } from '../constants/bill'
import { formatYuan } from '../utils/money'
import { formatRFC3339Minute } from '../utils/time'

interface MonthSummary {
  total_income: string; total_expense: string
  alipay_income: string; alipay_expense: string
  wechat_income: string; wechat_expense: string; investment_amount: string
}
interface BillRecord {
  id: string; channel: number; transaction_time: string
  category: string; description: string; bill_type: number; amount: string
}

export default function MonthDetail() {
  const { year, month } = useParams<{ year: string; month: string }>()
  const y = Number(year), m = Number(month)
  const navigate = useNavigate()
  const [summary, setSummary] = useState<MonthSummary | null>(null)
  const [aiText, setAiText] = useState('')
  const [bills, setBills] = useState<BillRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [amountSort, setAmountSort] = useState<'asc' | 'desc' | null>(null)
  const [showModal, setShowModal] = useState(false)
  const [triggering, setTriggering] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const pageSize = 20

  const stopPolling = () => {
    if (pollRef.current) { clearInterval(pollRef.current); pollRef.current = null }
  }

  // loadMeta: summary + AI — only on mount
  const loadMeta = useCallback(() => {
    getMonthDetail(y, m).then(r => setSummary(r.data.data)).catch(() => {})
    getAISummary(y, m, y, m).then(r => setAiText(r.data.data?.ai_analysis || '')).catch(() => {})
  }, [y, m])

  // loadBills: only bills — on page/sort change
  const loadBills = useCallback((pg: number, sortOrder: string) => {
    listBills(y, m, pg, pageSize, sortOrder)
      .then(r => {
        setBills(r.data.data?.items || [])
        setTotal(Number(r.data.data?.total || 0))
      })
      .catch(() => {})
  }, [y, m])

  useEffect(() => { loadMeta() }, [loadMeta])
  useEffect(() => { loadBills(page, amountSort || '') }, [loadBills, page, amountSort])
  useEffect(() => () => stopPolling(), [])

  const handleTriggerAI = async () => {
    setTriggering(true)
    const prevText = aiText
    try {
      await triggerAI(y, m, y, m)
    } catch (e: unknown) {
      setAiText('AI 分析失败：' + (e instanceof Error ? e.message : ''))
      setTriggering(false)
      return
    }
    let elapsed = 0
    stopPolling()
    pollRef.current = setInterval(async () => {
      elapsed += 10
      try {
        const r = await getAISummary(y, m, y, m)
        const text = r.data.data?.ai_analysis || ''
        if (text && text !== prevText) {
          setAiText(text); setTriggering(false); stopPolling(); return
        }
      } catch { /* ignore polling errors */ }
      if (elapsed >= 60) { setTriggering(false); stopPolling() }
    }, 10000)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确认删除这条记录？')) return
    try { await deleteBill(id); loadBills(page, amountSort || '') } catch { /* ignore delete errors */ }
  }

  const statCards = summary ? [
    { label: '总支出', val: summary.total_expense, color: 'text-red-500' },
    { label: '总收入', val: summary.total_income, color: 'text-green-600' },
    { label: '支付宝支出', val: summary.alipay_expense, color: 'text-orange-500' },
    { label: '支付宝收入', val: summary.alipay_income, color: 'text-orange-400' },
    { label: '微信支出', val: summary.wechat_expense, color: 'text-gray-700' },
    { label: '微信收入', val: summary.wechat_income, color: 'text-gray-500' },
    { label: '理财支出', val: summary.investment_amount, color: 'text-purple-600' },
  ] : []

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center gap-4">
          <button onClick={() => navigate('/')} className="text-sm text-gray-500 hover:text-gray-700">← 返回</button>
          <h2 className="font-semibold text-gray-700">{y}年{m}月 账单详情</h2>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-6 py-6 space-y-6">
        {statCards.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3">
            {statCards.map(({ label, val, color }) => (
              <div key={label} className="bg-white rounded-xl border border-gray-200 p-4">
                <p className="text-xs text-gray-400 mb-1">{label}</p>
                <p className={`text-base font-bold ${color}`}>{formatYuan(val)}</p>
              </div>
            ))}
          </div>
        )}

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
                  <th className="px-4 py-3 text-left font-medium">操作</th>
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
                    <td className="px-4 py-3"><button onClick={() => handleDelete(b.id)} className="text-red-400 hover:text-red-600 text-xs">删除</button></td>
                  </tr>
                ))}
                {bills.length === 0 && <tr><td colSpan={7} className="text-center py-10 text-gray-400">暂无明细</td></tr>}
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
      </main>
      {showModal && <ManualAddModal onClose={() => setShowModal(false)} onSuccess={() => { setShowModal(false); loadBills(page, amountSort || '') }} />}
    </div>
  )
}
