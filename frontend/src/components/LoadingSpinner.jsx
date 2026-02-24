import './LoadingSpinner.css'

function LoadingSpinner({ size = 'medium', text = '' }) {
  const sizeClass = {
    small: 'spinner-sm',
    medium: '',
    large: 'spinner-lg'
  }[size]

  return (
    <div className="loading-container">
      <div className={`spinner ${sizeClass}`}></div>
      {text && <p className="loading-text">{text}</p>}
    </div>
  )
}

export default LoadingSpinner
