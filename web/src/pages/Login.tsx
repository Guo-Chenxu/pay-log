import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login } from '../api/auth'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const res = await login(username, password)
      localStorage.setItem('token', res.data.data.token)
      navigate('/')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '登录失败')
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="bg-white rounded-xl shadow-md p-10 w-full max-w-sm">
        <h2 className="text-2xl font-semibold text-center text-gray-700 mb-8">Pay Log</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            className="w-full border border-gray-300 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-green-400"
            placeholder="用户名"
            value={username}
            onChange={e => setUsername(e.target.value)}
            required
          />
          <input
            className="w-full border border-gray-300 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-green-400"
            type="password"
            placeholder="密码"
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
          />
          {error && <p className="text-red-500 text-sm">{error}</p>}
          <button
            type="submit"
            className="w-full bg-green-500 hover:bg-green-600 text-white font-medium py-2.5 rounded-lg transition-colors"
          >
            登录
          </button>
        </form>
      </div>
    </div>
  )
}
