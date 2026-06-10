import axios from 'axios'

const client = axios.create({ baseURL: '/api/v1' })

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers['Authorization'] = token
  return config
})

client.interceptors.response.use(
  (res) => {
    if (res.data.code === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
      return Promise.reject(new Error(res.data.message || '未授权'))
    }
    if (res.data.code !== 0) return Promise.reject(new Error(res.data.message))
    return res
  },
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default client
