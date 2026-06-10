import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Home from './pages/Home'
import MonthDetail from './pages/MonthDetail'
import Upload from './pages/Upload'
import RangeDetail from './pages/RangeDetail'

function PrivateRoute({ children }: { children: React.ReactElement }) {
  return localStorage.getItem('token') ? children : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<PrivateRoute><Home /></PrivateRoute>} />
        <Route path="/month/:year/:month" element={<PrivateRoute><MonthDetail /></PrivateRoute>} />
        <Route path="/range" element={<PrivateRoute><RangeDetail /></PrivateRoute>} />
        <Route path="/upload" element={<PrivateRoute><Upload /></PrivateRoute>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
