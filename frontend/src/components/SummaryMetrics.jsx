import './SummaryMetrics.css'

function SummaryMetrics({ weeklyTrend, greenCount, yellowCount, redCount, vsLastWeek }) {
  return (
    <div className="summary-metrics">
      <div className="summary-card">
        <span className="summary-label">WEEKLY TRND</span>
        <span className={`summary-value trend-${weeklyTrend.direction}`}>
          {weeklyTrend.direction === 'up' ? '+' : ''}{weeklyTrend.change.toFixed(1)}%
        </span>
        <span className="summary-desc">week over week</span>
      </div>

      <div className="summary-card">
        <span className="summary-label">MEET TARGET</span>
        <span className="summary-value text-green">{greenCount}</span>
        <span className="summary-desc">metrics performing</span>
      </div>

      <div className="summary-card status-yellow-bg">
        <span className="summary-label">AT RISK</span>
        <span className="summary-value text-yellow">{yellowCount}</span>
        <span className="summary-desc">need attention</span>
      </div>

      <div className="summary-card status-red-bg">
        <span className="summary-label">OFF TRACK</span>
        <span className="summary-value text-red">{redCount}</span>
        <span className="summary-desc">underperforming</span>
      </div>

      {vsLastWeek !== undefined && (
        <div className="summary-card vs-last-week">
          <span className={`vs-badge ${vsLastWeek >= 0 ? 'positive' : 'negative'}`}>
            {vsLastWeek >= 0 ? '+' : ''}{vsLastWeek} vs last week
          </span>
        </div>
      )}
    </div>
  )
}

export default SummaryMetrics
