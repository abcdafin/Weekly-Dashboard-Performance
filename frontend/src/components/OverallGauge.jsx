import './OverallGauge.css'

function OverallGauge({ percentage, status }) {
  // Calculate the stroke dashoffset for the gauge arc
  const radius = 90
  const circumference = 2 * Math.PI * radius
  const progress = Math.min(Math.max(percentage, 0), 100)
  const strokeDashoffset = circumference - (progress / 100) * circumference

  const statusColors = {
    green: { stroke: 'var(--color-green)', bg: 'var(--color-green-bg)' },
    supergreen: { stroke: 'var(--color-super-green)', bg: 'var(--color-super-green-bg)' },
    yellow: { stroke: 'var(--color-yellow)', bg: 'var(--color-yellow-bg)' },
    red: { stroke: 'var(--color-red)', bg: 'var(--color-red-bg)' },
  }

  const colors = statusColors[status] || statusColors.red

  return (
    <div className="overall-gauge">
      <svg className="gauge-svg" viewBox="0 0 200 200">
        {/* Background circle */}
        <circle
          className="gauge-bg"
          cx="100"
          cy="100"
          r={radius}
          fill="none"
          strokeWidth="16"
        />
        {/* Progress arc */}
        <circle
          className="gauge-progress"
          cx="100"
          cy="100"
          r={radius}
          fill="none"
          strokeWidth="16"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          strokeLinecap="round"
          style={{ stroke: colors.stroke }}
          transform="rotate(-90 100 100)"
        />
      </svg>
      <div className="gauge-center">
        <span className="gauge-value">{percentage.toFixed(2)}</span>
        <span className="gauge-unit">%</span>
      </div>
    </div>
  )
}

export default OverallGauge
