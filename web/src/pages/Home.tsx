import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getOverview } from '../api/summary'
import { logout } from '../api/auth'
import ManualAddModal from '../components/ManualAddModal'
import { formatYuan } from '../utils/money'

interface MonthSummary {
  id: string; year: number; month: number
  total_income: string; total_expense: string
  alipay_expense: string; wechat_expense: string; investment_amount: string
}

export default function Home() {
  const [items, setItems] = useState<MonthSummary[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [showModal, setShowModal] = useState(false)
  const navigate = useNavigate()
  const pageSize = 12

  const loadOverview = useCallback(() => {
    getOverview(page, pageSize).then(res => {
      setItems(res.data.data.items || [])
      setTotal(Number(res.data.data.total || 0))
    }).catch(() => {})
  }, [page, pageSize])

  useEffect(() => { loadOverview() }, [loadOverview])

  const handleLogout = async () => {
    try { await logout() } catch { /* ignore logout errors */ }
    localStorage.removeItem('token')
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <h1 className="text-lg font-semibold text-gray-800">账单概览</h1>
          <div className="flex gap-2">
            <button onClick={() => setShowModal(true)} className="px-4 py-1.5 text-sm bg-green-500 text-white rounded-lg hover:bg-green-600">+ 手动添加</button>
            <button onClick={() => navigate('/upload')} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-gray-50">上传账单</button>
            <button onClick={() => navigate('/range')} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-gray-50">多月汇总</button>
            <button onClick={handleLogout} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 text-red-500">退出</button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {items.length === 0 && (
          <div className="text-center text-gray-400 py-24">暂无数据，请先上传账单或手动添加</div>
        )}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map(item => (
            <div
              key={item.id}
              onClick={() => navigate(`/month/${item.year}/${item.month}`)}
              className="bg-white rounded-xl border border-gray-200 p-5 cursor-pointer hover:shadow-md transition-shadow"
            >
              <h3 className="font-semibold text-gray-700 mb-3">{item.year}年{item.month}月</h3>
              <div className="flex gap-2 mb-3">
                <span className="text-xs px-2 py-1 bg-red-50 text-red-500 rounded-full">支出 {formatYuan(item.total_expense)}</span>
                <span className="text-xs px-2 py-1 bg-green-50 text-green-600 rounded-full">收入 {formatYuan(item.total_income)}</span>
              </div>
              <p className="text-xs text-gray-400">支付宝 {formatYuan(item.alipay_expense)} | 微信 {formatYuan(item.wechat_expense)}</p>
              <p className="text-xs text-purple-500 mt-1">理财 {formatYuan(item.investment_amount)}</p>
            </div>
          ))}
        </div>

        {total > pageSize && (
          <div className="mt-8 flex items-center justify-center gap-4">
            <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-40 hover:bg-gray-50">上一页</button>
            <span className="text-sm text-gray-500">第 {page} 页 / 共 {Math.ceil(total / pageSize)} 页</span>
            <button disabled={page * pageSize >= total} onClick={() => setPage(p => p + 1)} className="px-4 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-40 hover:bg-gray-50">下一页</button>
          </div>
        )}
      </main>
      {showModal && <ManualAddModal onClose={() => setShowModal(false)} onSuccess={() => { setShowModal(false); loadOverview() }} />}
    </div>
  )
}
