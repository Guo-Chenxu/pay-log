import { useState, type FormEvent } from 'react'
import { addManual } from '../api/bill'
import { BILL_TYPE, BILL_TYPE_LABEL, CHANNEL, CHANNEL_LABEL, type BillType, type Channel } from '../constants/bill'
import { localDateTimeToRFC3339 } from '../utils/time'

interface Props { onClose: () => void; onSuccess: () => void }

interface FormState {
  channel: string
  transaction_time: string
  counterparty: string
  description: string
  category: string
  bill_type: string
  amount: string
  payment_method: string
  remark: string
  is_investment: boolean
}

type TextFieldKey = Exclude<keyof FormState, 'is_investment'>

interface Field {
  label: string
  key: TextFieldKey
  type: string
  options?: [string, string][]
  placeholder?: string
  inputMode?: 'decimal'
}

export default function ManualAddModal({ onClose, onSuccess }: Props) {
  const [form, setForm] = useState<FormState>({
    channel: String(CHANNEL.ALIPAY), transaction_time: '', counterparty: '', description: '',
    category: '', bill_type: String(BILL_TYPE.EXPENSE), amount: '', payment_method: '', remark: '',
    is_investment: false,
  })
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  const set = <K extends keyof FormState>(k: K, v: FormState[K]) => setForm(f => ({ ...f, [k]: v }))

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setErr(''); setLoading(true)
    try {
      await addManual({
        channel: Number(form.channel) as Channel,
        transaction_time: localDateTimeToRFC3339(form.transaction_time),
        counterparty: form.counterparty,
        description: form.description,
        category: form.category,
        bill_type: Number(form.bill_type) as BillType,
        amount: form.amount,
        payment_method: form.payment_method,
        remark: form.remark,
        is_investment: form.is_investment,
      })
      onSuccess()
    } catch (e: unknown) { setErr(e instanceof Error ? e.message : '添加失败') } finally { setLoading(false) }
  }

  const fields: Field[] = [
    { label: '渠道', key: 'channel', type: 'select', options: [[String(CHANNEL.ALIPAY), CHANNEL_LABEL[CHANNEL.ALIPAY]], [String(CHANNEL.WECHAT), CHANNEL_LABEL[CHANNEL.WECHAT]]] },
    { label: '类型', key: 'bill_type', type: 'select', options: [[String(BILL_TYPE.INCOME), BILL_TYPE_LABEL[BILL_TYPE.INCOME]], [String(BILL_TYPE.EXPENSE), BILL_TYPE_LABEL[BILL_TYPE.EXPENSE]], [String(BILL_TYPE.NEUTRAL), BILL_TYPE_LABEL[BILL_TYPE.NEUTRAL]]] },
    { label: '交易时间', key: 'transaction_time', type: 'datetime-local', placeholder: '' },
    { label: '交易对方', key: 'counterparty', type: 'text', placeholder: '' },
    { label: '商品说明', key: 'description', type: 'text', placeholder: '' },
    { label: '商品类别', key: 'category', type: 'text', placeholder: '' },
    { label: '金额', key: 'amount', type: 'text', placeholder: '0.00', inputMode: 'decimal' },
    { label: '支付方式', key: 'payment_method', type: 'text', placeholder: '' },
    { label: '备注', key: 'remark', type: 'text', placeholder: '' },
  ]

  const inputCls = "w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-400"

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-md max-h-[85vh] overflow-y-auto shadow-xl">
        <h3 className="font-semibold text-gray-700 mb-5">手动添加账单</h3>
        <form onSubmit={handleSubmit} className="space-y-3">
          {fields.map(({ label, key, type, options, placeholder, inputMode }) => (
            <div key={key}>
              <label className="block text-xs text-gray-500 mb-1">{label}</label>
              {type === 'select' ? (
                <select className={inputCls} value={form[key]} onChange={e => set(key, e.target.value)}>
                  {(options || []).map(([v, t]) => <option key={v} value={v}>{t}</option>)}
                </select>
              ) : (
                <input className={inputCls} type={type} inputMode={inputMode} step={type === 'datetime-local' ? 60 : undefined} placeholder={placeholder} value={form[key]} onChange={e => set(key, e.target.value)} />
              )}
            </div>
          ))}
          <label className="flex items-start gap-2 text-sm text-gray-600">
            <input type="checkbox" checked={form.is_investment} onChange={e => set('is_investment', e.target.checked)} className="mt-1 accent-green-500" />
            <span>
              理财支出
              <span className="block text-xs text-gray-400 mt-0.5">勾选后会单独计入理财金额，不计入普通支出。</span>
            </span>
          </label>
          {err && <p className="text-red-500 text-xs">{err}</p>}
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50">取消</button>
            <button type="submit" disabled={loading} className="px-4 py-2 text-sm bg-green-500 hover:bg-green-600 text-white rounded-lg disabled:opacity-50">
              {loading ? '保存中...' : '保存'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
