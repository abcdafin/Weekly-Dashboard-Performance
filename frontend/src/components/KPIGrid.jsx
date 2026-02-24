import KPICard from './KPICard'
import './KPIGrid.css'

function KPIGrid({ indicators }) {
  if (!indicators || indicators.length === 0) {
    return (
      <div className="kpi-grid-empty">
        <p>No KPI data available for the selected period.</p>
      </div>
    )
  }

  return (
    <div className="kpi-grid">
      {indicators.map((indicator) => (
        <KPICard key={indicator.code} indicator={indicator} />
      ))}
    </div>
  )
}

export default KPIGrid
