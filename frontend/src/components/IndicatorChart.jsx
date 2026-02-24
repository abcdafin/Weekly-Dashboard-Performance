import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell, ReferenceLine, LabelList } from 'recharts'
import './IndicatorChart.css'

const STATUS_COLORS = {
  green: '#16a34a',
  yellow: '#ca8a04',
  red: '#dc2626',
}

function getStatusColor(percentage) {
  if (percentage > 85) return STATUS_COLORS.green
  if (percentage > 55) return STATUS_COLORS.yellow
  return STATUS_COLORS.red
}

const CustomTooltip = ({ active, payload }) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload
    return (
      <div className="chart-tooltip">
        <p className="tooltip-week">Week {data.week}</p>
        <p className="tooltip-percentage" style={{ color: getStatusColor(data.percentage) }}>
          {data.percentage.toFixed(1)}%
        </p>
      </div>
    )
  }
  return null
}

function IndicatorChart({ indicator }) {
  const { code, department, name, weeks } = indicator

  const chartData = weeks.map(w => ({
    week: w.week,
    weekLabel: `W${w.week}`,
    percentage: w.percentage,
  }))

  return (
    <div className="indicator-chart-card">
      <div className="chart-header">
        <span className="chart-department">{department}</span>
        <h4 className="chart-title">{name}</h4>
        <span className="chart-code">{code}</span>
      </div>
      <div className="chart-body">
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={chartData} margin={{ top: 20, right: 10, left: -10, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(0,0,0,0.08)" />
            <XAxis 
              dataKey="weekLabel" 
              tick={{ fill: '#475569', fontSize: 12 }} 
              axisLine={{ stroke: 'rgba(0,0,0,0.15)' }}
            />
            <YAxis 
              domain={[0, 120]} 
              tick={{ fill: '#475569', fontSize: 11 }} 
              axisLine={{ stroke: 'rgba(0,0,0,0.15)' }}
              tickFormatter={(v) => `${v}%`}
            />
            <Tooltip content={<CustomTooltip />} cursor={{ fill: 'rgba(0,0,0,0.04)' }} />
            <ReferenceLine y={85} stroke="rgba(22,163,74,0.4)" strokeDasharray="4 4" label={{ value: '85%', fill: '#6b7280', fontSize: 10, position: 'right' }} />
            <Bar dataKey="percentage" radius={[6, 6, 0, 0]} maxBarSize={50}>
              <LabelList dataKey="percentage" position="top" formatter={(v) => `${v.toFixed(1)}%`} style={{ fill: '#334155', fontSize: 11, fontWeight: 600 }} />
              {chartData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={getStatusColor(entry.percentage)} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

export default IndicatorChart
