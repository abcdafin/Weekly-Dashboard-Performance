import { useState, useRef, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import html2canvas from 'html2canvas'
import Header from '../components/Header'
import MonthSelector from '../components/MonthSelector'
import OverallGauge from '../components/OverallGauge'
import SummaryMetrics from '../components/SummaryMetrics'
import KPIGrid from '../components/KPIGrid'
import LoadingSpinner from '../components/LoadingSpinner'
import WeekSelectorModal from '../components/WeekSelectorModal'
import ScreenshotGallery from '../components/ScreenshotGallery'
import { useDashboard } from '../hooks/useDashboard'
import { useToast } from '../context/ToastContext'
import { screenshotApi, dashboardApi } from '../services/api'
import './DashboardPage.css'

function DashboardPage() {
  const [searchParams] = useSearchParams()
  const initialMonth = parseInt(searchParams.get('month')) || new Date().getMonth() + 1
  const initialYear = parseInt(searchParams.get('year')) || new Date().getFullYear()
  const dashboardRef = useRef(null)
  const [showWeekModal, setShowWeekModal] = useState(false)
  const [savingScreenshot, setSavingScreenshot] = useState(false)
  const [refreshGallery, setRefreshGallery] = useState(0)
  const { success: showSuccess, error: showError } = useToast()
  
  const { 
    data, 
    loading, 
    error, 
    month, 
    year, 
    changeMonth,
    refresh,
    lastUpdated,
    autoRefresh,
    toggleAutoRefresh
  } = useDashboard(initialMonth, initialYear)



  const handleWeeklySnapshot = useCallback(async (week) => {
    if (!dashboardRef.current) return
    
    setSavingScreenshot(true)
    try {
      // Generate PNG
      const canvas = await html2canvas(dashboardRef.current, {
        backgroundColor: '#1a1a2e',
        scale: 2,
        useCORS: true,
        logging: false,
      })
      
      // Convert to blob
      const blob = await new Promise(resolve => canvas.toBlob(resolve, 'image/png'))
      
      // Create form data
      const formData = new FormData()
      formData.append('file', blob, 'screenshot.png')
      formData.append('month', month.toString())
      formData.append('year', year.toString())
      formData.append('week', week.toString())
      
      // Upload screenshot to server
      await screenshotApi.upload(formData)
      
      // Also save data snapshot for the same week
      await dashboardApi.saveSnapshot(month, year, week)
      
      showSuccess(`Week ${week} snapshot saved successfully!`)
      setShowWeekModal(false)
      setRefreshGallery(prev => prev + 1) // Trigger gallery refresh
    } catch (err) {
      console.error('Save snapshot failed:', err)
      showError('Failed to save snapshot')
    } finally {
      setSavingScreenshot(false)
    }
  }, [month, year, showSuccess, showError])

  if (loading) {
    return (
      <div className="dashboard-layout">
        <Header />
        <main className="dashboard-main">
          <div className="container">
            <div className="loading-state">
              <LoadingSpinner size="large" text="Loading dashboard data..." />
            </div>
          </div>
        </main>
      </div>
    )
  }

  if (error) {
    return (
      <div className="dashboard-layout">
        <Header />
        <main className="dashboard-main">
          <div className="container">
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h2>Unable to load dashboard</h2>
              <p>{error}</p>
              <button className="btn btn-primary" onClick={() => window.location.reload()}>
                Try Again
              </button>
            </div>
          </div>
        </main>
      </div>
    )
  }

  const { overall_performance, weekly_trend, indicators } = data || {}

  return (
    <div className="dashboard-layout">
      <Header />
      
      <main className="dashboard-main">
        <div className="container" ref={dashboardRef}>
          {/* Month Selector */}
          <MonthSelector 
            month={month} 
            year={year} 
            onChange={changeMonth} 
          />

          {/* Top Section: Overall Performance + Summary */}
          <section className="dashboard-overview">
            <div className="overview-left">
              {overall_performance && (
                <OverallGauge 
                  percentage={overall_performance.percentage} 
                  status={overall_performance.status} 
                />
              )}
            </div>
            
            <div className="overview-right">
              {overall_performance && weekly_trend && (
                <SummaryMetrics
                  weeklyTrend={weekly_trend}
                  greenCount={overall_performance.green_count}
                  yellowCount={overall_performance.yellow_count}
                  redCount={overall_performance.red_count}
                  vsLastWeek={weekly_trend.green_count_change}
                />
              )}
            </div>
          </section>

          {/* Analysis Status */}
          <div className="analysis-status">
            <span className="status-indicator complete">◀ ANALYSIS COMPLETE</span>
            <div className="refresh-section">
              {lastUpdated && (
                <span className="last-updated">
                  Last updated: {lastUpdated.toLocaleDateString('en-US', { 
                    weekday: 'short'
                  })}, {lastUpdated.getDate()} {lastUpdated.toLocaleDateString('en-US', { 
                    month: 'short'
                  })} {lastUpdated.getFullYear()} {lastUpdated.toLocaleTimeString('en-US', { 
                    hour: '2-digit', 
                    minute: '2-digit',
                    second: '2-digit',
                    hour12: false
                  })} WIB
                </span>
              )}
              <button 
                className="btn btn-refresh" 
                onClick={refresh}
                disabled={loading}
                title="Refresh data"
              >
                🔄 Refresh
              </button>
              <button 
                className={`btn btn-auto-refresh ${autoRefresh ? 'active' : ''}`}
                onClick={toggleAutoRefresh}
                title={autoRefresh ? 'Disable auto-refresh' : 'Enable auto-refresh (every 5 min)'}
              >
                {autoRefresh ? '⏸️ Auto' : '▶️ Auto'}
              </button>
            </div>
          </div>

          {/* KPI Grid */}
          <section className="dashboard-kpis">
            <KPIGrid indicators={indicators || []} />
          </section>
        </div>

        {/* Screenshot Gallery with Snapshot Button */}
        <div className="container">
          <ScreenshotGallery 
            month={month} 
            year={year} 
            refreshTrigger={refreshGallery}
            onSnapshot={() => setShowWeekModal(true)}
            savingScreenshot={savingScreenshot}
          />
        </div>
      </main>

      {/* Week Selector Modal */}
      <WeekSelectorModal
        isOpen={showWeekModal}
        onClose={() => setShowWeekModal(false)}
        onSave={handleWeeklySnapshot}
        month={month}
        year={year}
        saving={savingScreenshot}
      />

      <footer className="dashboard-footer">
        <div className="container">
          <p>© {new Date().getFullYear()} IDstar - Weekly Performance Dashboard</p>
        </div>
      </footer>
    </div>
  )
}

export default DashboardPage

