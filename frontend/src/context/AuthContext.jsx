import { useState, useEffect, useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import api from '../services/api'
import { AuthContext, TOKEN_KEY } from './authContextValue'

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  // Logout function - defined first so it can be used in fetchUser
  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
    setIsAuthenticated(false)
    navigate('/login')
  }, [navigate])

  // Fetch user data from API
  const fetchUser = useCallback(async () => {
    try {
      const response = await api.get('/auth/me')
      if (response.data.success) {
        setUser(response.data.data)
        setIsAuthenticated(true)
      }
    } catch (error) {
      console.error('Failed to fetch user:', error)
      logout()
    } finally {
      setLoading(false)
    }
  }, [logout])

  // Initialize auth state from token
  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      fetchUser()
    } else {
      setLoading(false)
    }
  }, [fetchUser])

  // Handle OAuth callback token
  useEffect(() => {
    const params = new URLSearchParams(location.search)
    const token = params.get('token')
    
    if (token && location.pathname === '/auth/callback') {
      localStorage.setItem(TOKEN_KEY, token)
      // Clear the token from URL
      window.history.replaceState({}, '', '/dashboard')
      fetchUser().then(() => {
        navigate('/dashboard', { replace: true })
      })
    }
  }, [location, navigate, fetchUser])

  const login = () => {
    // Redirect to Google OAuth
    window.location.href = `${import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'}/auth/google`
  }

  // logout is now defined above fetchUser

  const getToken = () => localStorage.getItem(TOKEN_KEY)

  return (
    <AuthContext.Provider value={{ 
      user, 
      loading, 
      isAuthenticated, 
      login, 
      logout, 
      getToken 
    }}>
      {children}
    </AuthContext.Provider>
  )
}
