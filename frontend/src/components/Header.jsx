import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useAuth } from '../context/useAuth'
import ConfirmModal from './ConfirmModal'
import SettingsModal from './SettingsModal'
import './Header.css'

function Header() {
  const { user, logout } = useAuth()
  const [showLogoutModal, setShowLogoutModal] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const location = useLocation()

  const handleLogoutClick = () => {
    setShowLogoutModal(true)
  }

  const handleConfirmLogout = () => {
    setShowLogoutModal(false)
    logout()
  }

  const handleCancelLogout = () => {
    setShowLogoutModal(false)
  }

  return (
    <>
      <header className="header">
        <div className="header-container">
          <div className="header-left">
            <img src="/logo.png" alt="IDstar" className="header-logo" />
            <div className="header-title">
              <h1>Weekly Performance Score</h1>
              <span className="header-subtitle">CURRENT WEEK ANALYSIS</span>
            </div>
          </div>

          <nav className="header-nav">
            <Link 
              to="/dashboard" 
              className={`nav-link ${location.pathname === '/dashboard' ? 'active' : ''}`}
            >
              📋 Dashboard
            </Link>
            <Link 
              to="/charts" 
              className={`nav-link ${location.pathname === '/charts' ? 'active' : ''}`}
            >
              📊 Monthly Charts
            </Link>
            <button 
              className={`nav-link settings-btn`}
              onClick={() => setShowSettings(true)}
              title="Spreadsheet Settings"
            >
              ⚙️ Settings
            </button>
          </nav>
          
          <div className="header-right">
            {user && (
              <div className="user-menu">
                <img 
                  src={user.picture || '/default-avatar.png'} 
                  alt={user.name} 
                  className="user-avatar"
                />
                <div className="user-info">
                  <span className="user-name">{user.name}</span>
                  <span className="user-email">{user.email}</span>
                </div>
                <button className="btn btn-secondary btn-sm" onClick={handleLogoutClick}>
                  Logout
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      <ConfirmModal
        isOpen={showLogoutModal}
        title="Logout"
        message="Are you sure you want to logout from your account?"
        onConfirm={handleConfirmLogout}
        onCancel={handleCancelLogout}
        confirmText="Logout"
        cancelText="Cancel"
        variant="danger"
      />

      <SettingsModal
        isOpen={showSettings}
        onClose={() => setShowSettings(false)}
      />
    </>
  )
}

export default Header


