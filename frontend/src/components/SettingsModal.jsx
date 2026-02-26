import { useState, useEffect, useCallback } from 'react'
import { settingsApi } from '../services/api'
import { useToast } from '../context/ToastContext'
import './SettingsModal.css'

function SettingsModal({ isOpen, onClose }) {
  const [spreadsheetId, setSpreadsheetId] = useState('')
  const [sheetName, setSheetName] = useState('')
  const [spreadsheetYear, setSpreadsheetYear] = useState(2026)
  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(false)
  const { success: showSuccess, error: showError } = useToast()

  const fetchCurrentSettings = useCallback(async () => {
    setFetching(true)
    try {
      const response = await settingsApi.getSpreadsheet()
      const data = response.data.data
      setSpreadsheetId(data.spreadsheet_id || '')
      setSheetName(data.sheet_name || '')
      setSpreadsheetYear(data.spreadsheet_year || 2026)
    } catch (err) {
      console.error('Failed to fetch settings:', err)
      showError('Failed to load current settings')
    } finally {
      setFetching(false)
    }
  }, [showError])

  useEffect(() => {
    if (isOpen) {
      fetchCurrentSettings()
    }
  }, [isOpen, fetchCurrentSettings])

  const handleSave = async () => {
    if (!spreadsheetId.trim()) {
      showError('Spreadsheet ID or URL is required')
      return
    }

    setLoading(true)
    try {
      await settingsApi.updateSpreadsheet({
        spreadsheet_id: spreadsheetId.trim(),
        sheet_name: sheetName.trim(),
        spreadsheet_year: spreadsheetYear
      })
      showSuccess('Spreadsheet settings updated successfully!')
      onClose()
      // Reload page to apply new year setting
      window.location.reload()
    } catch (err) {
      console.error('Failed to update settings:', err)
      showError('Failed to update settings')
    } finally {
      setLoading(false)
    }
  }

  // Generate year options from 2026 to 2030
  const yearOptions = []
  for (let y = 2026; y <= 2030; y++) {
    yearOptions.push(y)
  }

  if (!isOpen) return null

  return (
    <div className="settings-overlay" onClick={onClose}>
      <div className="settings-modal" onClick={e => e.stopPropagation()}>
        <div className="settings-header">
          <h2>⚙️ Spreadsheet Settings</h2>
          <button className="settings-close" onClick={onClose}>×</button>
        </div>

        {fetching ? (
          <div className="settings-loading">Loading settings...</div>
        ) : (
          <div className="settings-body">
            <div className="settings-field">
              <label htmlFor="spreadsheet-id">Google Spreadsheet URL or ID</label>
              <input
                id="spreadsheet-id"
                type="text"
                value={spreadsheetId}
                onChange={e => setSpreadsheetId(e.target.value)}
                placeholder="Paste Google Sheets URL or Spreadsheet ID"
                className="settings-input"
              />
              <span className="settings-hint">
                You can paste the full Google Sheets URL or just the Spreadsheet ID
              </span>
            </div>

            <div className="settings-field">
              <label htmlFor="sheet-name">Sheet Name</label>
              <input
                id="sheet-name"
                type="text"
                value={sheetName}
                onChange={e => setSheetName(e.target.value)}
                placeholder="e.g. DashboardTemplate"
                className="settings-input"
              />
              <span className="settings-hint">
                The name of the sheet tab containing KPI data
              </span>
            </div>

            <div className="settings-field">
              <label htmlFor="spreadsheet-year">Spreadsheet Year</label>
              <select
                id="spreadsheet-year"
                value={spreadsheetYear}
                onChange={e => setSpreadsheetYear(parseInt(e.target.value))}
                className="settings-input"
              >
                {yearOptions.map(y => (
                  <option key={y} value={y}>{y}</option>
                ))}
              </select>
              <span className="settings-hint">
                The year that this spreadsheet data belongs to. The dashboard will display data for this year only.
              </span>
            </div>
          </div>
        )}

        <div className="settings-footer">
          <button
            className="btn btn-secondary"
            onClick={onClose}
            disabled={loading}
          >
            Cancel
          </button>
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={loading || fetching}
          >
            {loading ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default SettingsModal
