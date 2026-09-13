function pad(n) {
  return n < 10 ? `0${n}` : `${n}`
}

function toDateKey(date = new Date()) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function toMonthKey(date = new Date()) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}`
}

function fenToYuan(fen) {
  return (Number(fen || 0) / 100).toFixed(2)
}

function yuanToFen(value) {
  const normalized = String(value || '').trim()
  if (!/^\d+(\.\d{0,2})?$/.test(normalized)) return null
  const [yuan, cent = ''] = normalized.split('.')
  return Number(yuan) * 100 + Number((cent + '00').slice(0, 2))
}

function getMonthLabel(monthKey) {
  const [year, month] = String(monthKey).split('-')
  return `${year}年${Number(month)}月`
}

module.exports = {
  toDateKey,
  toMonthKey,
  fenToYuan,
  yuanToFen,
  getMonthLabel
}
