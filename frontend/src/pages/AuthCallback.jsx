import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import LoadingSpinner from '../components/LoadingSpinner'

function AuthCallback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  useEffect(() => {
    const token = searchParams.get('token')
    
    if (token) {
      localStorage.setItem('auth_token', token)
      navigate('/dashboard', { replace: true })
    } else {
      navigate('/login', { replace: true })
    }
  }, [searchParams, navigate])

  return (
    <div className="flex-center" style={{ minHeight: '100vh' }}>
      <LoadingSpinner size="large" text="Authenticating..." />
    </div>
  )
}

export default AuthCallback
