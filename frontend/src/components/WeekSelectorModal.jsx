import { useState } from 'react'
import './WeekSelectorModal.css'

function WeekSelectorModal({ isOpen, onClose, onSave, month, year, saving }) {
  const [selectedWeek, setSelectedWeek] = useState(1)

  if (!isOpen) return null

  const handleSave = () => {
    onSave(selectedWeek)
  }

  const monthName = new Date(year, month - 1).toLocaleString('en-US', { month: 'long' })

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content week-selector-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>📸 Weekly Snapshot</h2>
          <button className="modal-close" onClick={onClose}>&times;</button>
        </div>
        
        <div className="modal-body">
          <p className="modal-description">
            Save dashboard screenshot for <strong>{monthName} {year}</strong>
          </p>
          
          <div className="week-selector">
            <label htmlFor="week-select">Select Week:</label>
            <select 
              id="week-select"
              value={selectedWeek} 
              onChange={(e) => setSelectedWeek(parseInt(e.target.value))}
              disabled={saving}
            >
              <option value={1}>Week 1</option>
              <option value={2}>Week 2</option>
              <option value={3}>Week 3</option>
              <option value={4}>Week 4</option>
              <option value={5}>Week 5</option>
            </select>
          </div>

          <p className="modal-hint">
            File will be saved as: <code>{monthName}_{year}_Week_{selectedWeek}.png</code>
          </p>
          <p className="modal-warning">
            ⚠️ If the same week already exists, the file will be overwritten.
          </p>
        </div>

        <div className="modal-footer">
          <button 
            className="btn btn-secondary" 
            onClick={onClose}
            disabled={saving}
          >
            Cancel
          </button>
          <button 
            className="btn btn-primary" 
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? '💾 Saving...' : '💾 Save Snapshot'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default WeekSelectorModal
