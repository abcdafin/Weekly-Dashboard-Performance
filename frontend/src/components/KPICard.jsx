import "./KPICard.css";

function KPICard({ indicator }) {
  const {
    department,
    name,
    performance,
    percentage,
    status,
    is_inverse,
    wow_change,
    wow_direction,
  } = indicator;

  // Mini gauge calculations
  const radius = 45;
  const circumference = 2 * Math.PI * radius;
  const progress = Math.min(Math.max(percentage, 0), 150); // Allow up to 150%
  const normalizedProgress = (progress / 100) * 100; // Cap at 100% for visual
  const strokeDashoffset =
    circumference - (Math.min(normalizedProgress, 100) / 100) * circumference;

  const statusColors = {
    green: "var(--color-green)",
    supergreen: "var(--color-super-green)",
    yellow: "var(--color-yellow)",
    red: "var(--color-red)",
  };

  const formatValue = (val) => {
    if (typeof val !== "number" || isNaN(val)) return "0";
    if (Math.abs(val) >= 1000000000) return (val / 1000000000).toFixed(1) + "B";
    if (Math.abs(val) >= 1000000) return (val / 1000000).toFixed(1) + "M";
    if (Math.abs(val) >= 1000) return (val / 1000).toFixed(1) + "K";
    if (Number.isInteger(val)) return val.toString();
    return val.toFixed(1);
  };

  // Determine badge text and class based on status and inverse
  const getBadge = () => {
    if (is_inverse) {
      // Inverse metrics: lower is better
      switch (status) {
        case "supergreen":
          return { text: "EXCEEDED", className: "goal-exceeded" };
        case "green":
          return { text: "ON TRACK", className: "goal-met" };
        case "yellow":
          return { text: "AT RISK", className: "at-risk" };
        case "red":
          return { text: "OFF TRACK", className: "off-track" };
        default:
          return null;
      }
    } else {
      // Normal metrics: higher is better
      switch (status) {
        case "supergreen":
          return { text: "EXCEEDED", className: "goal-exceeded" };
        case "green":
          return { text: "GOAL MET", className: "goal-met" };
        case "yellow":
          return { text: "AT RISK", className: "at-risk" };
        case "red":
          return { text: "OFF TRACK", className: "off-track" };
        default:
          return null;
      }
    }
  };

  const badge = getBadge();

  return (
    <div className={`kpi-card kpi-${status}`}>
      <div className="kpi-header">
        <span className="kpi-department">{department}</span>
        <span className="kpi-name">{name}</span>
      </div>

      <div className="kpi-gauge-container">
        <svg className="kpi-gauge" viewBox="0 0 100 100">
          <circle
            className="kpi-gauge-bg"
            cx="50"
            cy="50"
            r={radius}
            fill="none"
            strokeWidth="8"
          />
          <circle
            className="kpi-gauge-progress"
            cx="50"
            cy="50"
            r={radius}
            fill="none"
            strokeWidth="8"
            strokeDasharray={circumference}
            strokeDashoffset={strokeDashoffset}
            strokeLinecap="round"
            style={{ stroke: statusColors[status] }}
            transform="rotate(-90 50 50)"
          />
        </svg>
        <div className="kpi-value">
          <span className="kpi-value-number">{formatValue(performance)}</span>
        </div>
      </div>

      <div className="kpi-footer">
        <div className="kpi-status-wrapper">
          {badge ? (
            <span className={`kpi-status ${badge.className}`}>
              {badge.text}
            </span>
          ) : (
            <span className="kpi-status placeholder">&nbsp;</span>
          )}
        </div>
        {/* Hidden per mentor feedback
        <div className="kpi-schedule-wrapper">
          <span className={`kpi-schedule-badge ${scheduleInfo.className}`}>
            {scheduleInfo.icon} {scheduleInfo.text}
          </span>
          {variance !== 0 && (
            <span className={`kpi-variance ${variance > 0 ? 'positive' : 'negative'}`}>
              {variance > 0 ? '+' : ''}{variance.toFixed(1)}%
            </span>
          )}
        </div>
        */}
        <div className="kpi-metrics">
          <span className="kpi-percentage">
            {percentage.toFixed(0)}% of goal
          </span>
          <span className={`kpi-trend trend-${wow_direction}`}>
            {wow_direction === "up" && "↑"}
            {wow_direction === "down" && "↓"}
            {wow_direction === "neutral" && "→"}
            {wow_change !== 0
              ? `${wow_change > 0 ? "+" : ""}${wow_change.toFixed(1)}%`
              : "0%"}
          </span>
        </div>
        {/* Hidden per mentor feedback
        {expected_progress > 0 && (
          <div className="kpi-expected">
            <span className="kpi-expected-label">Expected:</span>
            <span className="kpi-expected-value">{formatValue(expected_progress)}</span>
          </div>
        )}
        */}
      </div>
    </div>
  );
}

export default KPICard;
