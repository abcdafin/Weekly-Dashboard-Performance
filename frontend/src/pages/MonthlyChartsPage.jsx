import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import Header from '../components/Header'
import MonthSelector from '../components/MonthSelector'
import IndicatorChart from '../components/IndicatorChart'
import LoadingSpinner from '../components/LoadingSpinner'
import { dashboardApi } from '../services/api'
import './MonthlyChartsPage.css'

function MonthlyChartsPage() {
  const [searchParams] = useSearchParams()
  const initialMonth = parseInt(searchParams.get('month')) || new Date().getMonth() + 1
  const initialYear = parseInt(searchParams.get('year')) || new Date().getFullYear()
  
  const [month, setMonth] = useState(initialMonth)
  const [year, setYear] = useState(initialYear)
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const fetchSnapshots = async () => {
      setLoading(true)
      setError(null)
      try {
        const response = await dashboardApi.getSnapshots(month, year)
        if (response.data.success) {
          setData(response.data.data)
        }
      } catch (err) {
        console.error('Failed to fetch snapshots:', err)
        const message = err.response?.data?.error || 'Failed to load snapshot data'
        setError(message)
      } finally {
        setLoading(false)
      }
    }

    fetchSnapshots()
  }, [month, year])

  const handleMonthChange = (newMonth, newYear) => {
    setMonth(newMonth)
    setYear(newYear)
  }

  const monthName = new Date(year, month - 1).toLocaleString('en-US', { month: 'long' })

  return (
    <div className="dashboard-layout">
      <Header />
      
      <main className="dashboard-main">
        <div className="container">
          <MonthSelector 
            month={month} 
            year={year} 
            onChange={handleMonthChange} 
          />

          <div className="charts-page-header">
            <h2>📊 Monthly Performance Charts</h2>
            <p className="charts-subtitle">
              Weekly performance data for <strong>{monthName} {year}</strong>
              {data && data.available_weeks && data.available_weeks.length > 0 && (
                <span className="weeks-badge">
                  {data.available_weeks.length} week{data.available_weeks.length !== 1 ? 's' : ''} recorded
                </span>
              )}
            </p>
          </div>

          {loading && (
            <div className="loading-state">
              <LoadingSpinner size="large" text="Loading chart data..." />
            </div>
          )}

          {error && (
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h3>Unable to load charts</h3>
              <p>{error}</p>
            </div>
          )}

          {!loading && !error && (!data || !data.indicators || data.indicators.length === 0) && (
            <div className="empty-charts-state">
              <div className="empty-icon">📈</div>
              <h3>No snapshot data available</h3>
              <p>Save weekly snapshots from the Dashboard to see charts here.</p>
              <p className="hint">Go to Dashboard → Click "📸 Weekly Snapshot" to save data.</p>
            </div>
          )}

          {!loading && !error && data && data.indicators && data.indicators.length > 0 && (
            <div className="charts-grid">
              {data.indicators.map((indicator) => (
                <IndicatorChart 
                  key={indicator.code} 
                  indicator={indicator} 
                />
              ))}
            </div>
          )}
        </div>
      </main>

      <footer className="dashboard-footer">
        <div className="container">
          <p>© {new Date().getFullYear()} IDstar - Weekly Performance Dashboard</p>
        </div>
      </footer>
    </div>
  )
}

export default MonthlyChartsPage
