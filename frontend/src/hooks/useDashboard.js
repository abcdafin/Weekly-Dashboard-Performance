import { useState, useEffect, useCallback, useRef } from 'react'
import { dashboardApi } from '../services/api'
import { useToast } from '../context/ToastContext'

export function useDashboard(initialMonth, initialYear) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [month, setMonth] = useState(initialMonth || new Date().getMonth() + 1)
  const [year, setYear] = useState(initialYear || new Date().getFullYear())
  const [lastUpdated, setLastUpdated] = useState(null)
  const [savingSnapshot, setSavingSnapshot] = useState(false)
  const { success: showSuccess, error: showError } = useToast()
  const [autoRefresh, setAutoRefresh] = useState(false)
  const intervalRef = useRef(null)

  const fetchDashboard = useCallback(async (forceRefresh = false) => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await dashboardApi.getDashboard(month, year, forceRefresh)
      if (response.data.success) {
        setData(response.data.data)
        setLastUpdated(new Date())
      } else {
        throw new Error(response.data.error || 'Failed to fetch dashboard data')
      }
    } catch (err) {
      const message = err.response?.data?.error || err.message || 'Failed to load dashboard'
      setError(message)
      showError(message)
    } finally {
      setLoading(false)
    }
  }, [month, year, showError])

  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  const changeMonth = useCallback((newMonth, newYear) => {
    setMonth(newMonth)
    setYear(newYear)
    // Update URL params
    const url = new URL(window.location.href)
    url.searchParams.set('month', newMonth)
    url.searchParams.set('year', newYear)
    window.history.pushState({}, '', url)
  }, [])

  const refresh = useCallback(() => {
    fetchDashboard(true) // force refresh to re-discover layout
  }, [fetchDashboard])

  // Auto-refresh polling (every 5 minutes)
  useEffect(() => {
    if (autoRefresh) {
      intervalRef.current = setInterval(() => {
        fetchDashboard()
      }, 5 * 60 * 1000) // 5 minutes
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [autoRefresh, fetchDashboard])

  const toggleAutoRefresh = useCallback(() => {
    setAutoRefresh(prev => !prev)
  }, [])

  const saveSnapshot = useCallback(async () => {
    setSavingSnapshot(true)
    try {
      const response = await dashboardApi.saveSnapshot(month, year)
      if (response.data.success) {
        showSuccess('Weekly snapshot saved successfully!')
        return true
      } else {
        throw new Error(response.data.error || 'Failed to save snapshot')
      }
    } catch (err) {
      const message = err.response?.data?.error || err.message || 'Failed to save snapshot'
      showError(message)
      return false
    } finally {
      setSavingSnapshot(false)
    }
  }, [month, year, showSuccess, showError])

  return {
    data,
    loading,
    error,
    month,
    year,
    changeMonth,
    refresh,
    lastUpdated,
    saveSnapshot,
    savingSnapshot,
    autoRefresh,
    toggleAutoRefresh,
  }
}
