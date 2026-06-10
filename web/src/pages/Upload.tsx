import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { uploadBill } from '../api/bill'
import { CHANNEL, CHANNEL_ACCEPT, CHANNEL_LABEL, type Channel } from '../constants/bill'

const CHANNEL_OPTIONS = [CHANNEL.ALIPAY, CHANNEL.WECHAT] as const

export default function Upload() {
  const [channel, setChannel] = useState<Channel>(CHANNEL.ALIPAY)
  const [file, setFile] = useState<File | null>(null)
  const [result, setResult] = useState<{ success: number; skipped: number; failed: number } | null>(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleUpload = async () => {
    if (!file) return
    setLoading(true); setErr(''); setResult(null)
    try {
      const res = await uploadBill(file, channel)
      setResult(res.data.data)
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : '上传失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-start justify-center pt-16 px-4">
      <div className="bg-white rounded-xl shadow-md p-8 w-full max-w-md">
        <button onClick={() => navigate('/')} className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-flex items-center gap-1">
          ← 返回
        </button>
        <h2 className="text-xl font-semibold text-gray-700 mb-6">上传账单</h2>
        <div className="flex gap-6 mb-5">
          {CHANNEL_OPTIONS.map(v => (
            <label key={v} className="flex items-center gap-2 cursor-pointer text-sm text-gray-600">
              <input type="radio" value={v} checked={channel === v} onChange={() => setChannel(v)} className="accent-green-500" />
              {CHANNEL_LABEL[v]}
            </label>
          ))}
        </div>
        <input
          type="file"
          accept={CHANNEL_ACCEPT[channel]}
          onChange={e => setFile(e.target.files?.[0] || null)}
          className="block w-full text-sm text-gray-500 mb-5 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:bg-green-50 file:text-green-700 hover:file:bg-green-100"
        />
        <button
          onClick={handleUpload}
          disabled={!file || loading}
          className="w-full bg-green-500 hover:bg-green-600 disabled:bg-gray-300 text-white font-medium py-2.5 rounded-lg transition-colors"
        >
          {loading ? '上传中...' : '上传并解析'}
        </button>
        {err && <p className="text-red-500 text-sm mt-3">{err}</p>}
        {result && (
          <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-lg text-sm space-y-1">
            <p className="text-green-600">成功导入：{result.success} 条</p>
            <p className="text-gray-500">跳过重复：{result.skipped} 条</p>
            <p className={result.failed > 0 ? 'text-red-500' : 'text-gray-500'}>失败：{result.failed} 条</p>
          </div>
        )}
      </div>
    </div>
  )
}
