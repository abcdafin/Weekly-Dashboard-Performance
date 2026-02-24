import './MonthSelector.css'

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
]

function MonthSelector({ month, year, onChange }) {
  const currentYear = new Date().getFullYear()
  const years = [currentYear, currentYear + 1]

  const handlePrev = () => {
    let newMonth = month - 1
    let newYear = year
    if (newMonth < 1) {
      newMonth = 12
      newYear--
    }
    onChange(newMonth, newYear)
  }

  const handleNext = () => {
    let newMonth = month + 1
    let newYear = year
    if (newMonth > 12) {
      newMonth = 1
      newYear++
    }
    onChange(newMonth, newYear)
  }

  return (
    <div className="month-selector">
      <button 
        className="month-nav-btn" 
        onClick={handlePrev}
        aria-label="Previous month"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>

      <div className="month-dropdowns">
        <select 
          value={month} 
          onChange={(e) => onChange(parseInt(e.target.value), year)}
          className="month-select"
        >
          {MONTHS.map((name, idx) => (
            <option key={idx} value={idx + 1}>{name}</option>
          ))}
        </select>

        <select 
          value={year} 
          onChange={(e) => onChange(month, parseInt(e.target.value))}
          className="year-select"
        >
          {years.map(y => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>
      </div>

      <button 
        className="month-nav-btn" 
        onClick={handleNext}
        aria-label="Next month"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M9 18l6-6-6-6"/>
        </svg>
      </button>
    </div>
  )
}

export default MonthSelector
