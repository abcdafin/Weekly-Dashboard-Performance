import { useState, useEffect } from 'react'
import { screenshotApi, dashboardApi } from '../services/api'
import { useToast } from '../context/ToastContext'
import ConfirmModal from './ConfirmModal'
import './ScreenshotGallery.css'

function ScreenshotGallery({ month, year, refreshTrigger, onSnapshot, savingScreenshot }) {
  const [screenshots, setScreenshots] = useState([])
  const [loading, setLoading] = useState(false)
  const [previewImage, setPreviewImage] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [deleting, setDeleting] = useState(false)
  const { success: showSuccess, error: showError } = useToast()

  useEffect(() => {
    const fetchScreenshots = async () => {
      setLoading(true)
      try {
        const response = await screenshotApi.list(month, year)
        if (response.data.success) {
          setScreenshots(response.data.data || [])
        }
      } catch (err) {
        console.error('Failed to fetch screenshots:', err)
        setScreenshots([])
      } finally {
        setLoading(false)
      }
    }
    
    fetchScreenshots()
  }, [month, year, refreshTrigger])

  const handleDownload = async (screenshot) => {
    try {
      const response = await screenshotApi.getImage(screenshot.id)
      if (response.data.success) {
        const link = document.createElement('a')
        link.href = response.data.data.image
        link.download = screenshot.filename
        link.click()
      }
    } catch (err) {
      console.error('Download failed:', err)
    }
  }

  const handlePreview = (screenshot) => {
    setPreviewImage(screenshot)
  }

  const closePreview = () => {
    setPreviewImage(null)
  }

  const handleDeleteClick = (screenshot) => {
    setDeleteTarget(screenshot)
  }

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return
    
    setDeleting(true)
    try {
      await dashboardApi.deleteSnapshot(month, year, deleteTarget.week)
      showSuccess(`Week ${deleteTarget.week} snapshot deleted successfully!`)
      setDeleteTarget(null)
      // Refresh the gallery list
      const response = await screenshotApi.list(month, year)
      if (response.data.success) {
        setScreenshots(response.data.data || [])
      }
    } catch (err) {
      console.error('Delete failed:', err)
      showError('Failed to delete snapshot')
    } finally {
      setDeleting(false)
    }
  }

  const handleCancelDelete = () => {
    setDeleteTarget(null)
  }

  const monthName = new Date(year, month - 1).toLocaleString('en-US', { month: 'long' })

  if (loading) {
    return (
      <div className="screenshot-gallery">
        <div className="gallery-header">
          <h3>📷 Weekly Snapshots - {monthName} {year}</h3>
          {onSnapshot && (
            <button 
              className="btn btn-primary btn-snapshot"
              onClick={onSnapshot}
              disabled={savingScreenshot}
              title="Save dashboard screenshot and data snapshot"
            >
              {savingScreenshot ? '📸 Saving...' : '📸 Weekly Snapshot'}
            </button>
          )}
        </div>
        <div className="gallery-loading">Loading snapshots...</div>
      </div>
    )
  }

  if (screenshots.length === 0) {
    return (
      <div className="screenshot-gallery">
        <div className="gallery-header">
          <h3>📷 Weekly Snapshots - {monthName} {year}</h3>
          {onSnapshot && (
            <button 
              className="btn btn-primary btn-snapshot"
              onClick={onSnapshot}
              disabled={savingScreenshot}
              title="Save dashboard screenshot and data snapshot"
            >
              {savingScreenshot ? '📸 Saving...' : '📸 Weekly Snapshot'}
            </button>
          )}
        </div>
        <div className="gallery-empty">
          <p>No snapshots saved for this month yet.</p>
          <p className="hint">Click "📸 Weekly Snapshot" to save a screenshot.</p>
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="screenshot-gallery">
        <div className="gallery-header">
          <h3>📷 Weekly Snapshots - {monthName} {year}</h3>
          {onSnapshot && (
            <button 
              className="btn btn-primary btn-snapshot"
              onClick={onSnapshot}
              disabled={savingScreenshot}
              title="Save dashboard screenshot and data snapshot"
            >
              {savingScreenshot ? '📸 Saving...' : '📸 Weekly Snapshot'}
            </button>
          )}
        </div>
        <div className="gallery-grid">
          {screenshots.map((screenshot) => (
            <div key={screenshot.id} className="gallery-item">
              <div 
                className="gallery-preview"
                onClick={() => handlePreview(screenshot)}
              >
                <img 
                  src={screenshotApi.getImageUrl(screenshot.id)} 
                  alt={screenshot.filename}
                  loading="lazy"
                />
              </div>
              <div className="gallery-info">
                <span className="week-label">Week {screenshot.week}</span>
                <span className="saved-date">
                  {new Date(screenshot.saved_at).toLocaleDateString('en-US', {
                    day: 'numeric',
                    month: 'short',
                    hour: '2-digit',
                    minute: '2-digit'
                  })}
                </span>
              </div>
              <div className="gallery-actions">
                <button 
                  className="btn btn-download"
                  onClick={() => handleDownload(screenshot)}
                  title={`Download ${screenshot.filename}`}
                >
                  📥 Download
                </button>
                <button 
                  className="btn btn-delete-snapshot"
                  onClick={() => handleDeleteClick(screenshot)}
                  title={`Delete Week ${screenshot.week} snapshot`}
                >
                  🗑️ Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Preview Modal */}
      {previewImage && (
        <div className="preview-modal-overlay" onClick={closePreview}>
          <div className="preview-modal-content" onClick={e => e.stopPropagation()}>
            <div className="preview-modal-header">
              <span className="preview-modal-title">
                Week {previewImage.week} - {monthName} {year}
              </span>
              <button className="preview-modal-close" onClick={closePreview}>
                &times;
              </button>
            </div>
            <img 
              src={screenshotApi.getImageUrl(previewImage.id)} 
              alt={previewImage.filename}
            />
            <div className="preview-modal-actions">
              <button 
                className="btn btn-download"
                onClick={() => handleDownload(previewImage)}
                style={{ borderRadius: '8px' }}
              >
                📥 Download {previewImage.filename}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!deleteTarget}
        title="Delete Snapshot"
        message={deleteTarget ? `Are you sure you want to delete Week ${deleteTarget.week} snapshot? This will remove both the screenshot image and the data snapshot. This action cannot be undone.` : ''}
        onConfirm={handleConfirmDelete}
        onCancel={handleCancelDelete}
        confirmText={deleting ? 'Deleting...' : 'Delete'}
        cancelText="Cancel"
        variant="danger"
      />
    </>
  )
}

export default ScreenshotGallery
