function splitCsvLine(line) {
  const cells = []
  let current = ''
  let inQuote = false
  for (let i = 0; i < line.length; i += 1) {
    const char = line[i]
    const next = line[i + 1]
    if (char === '"' && inQuote && next === '"') {
      current += '"'
      i += 1
    } else if (char === '"') {
      inQuote = !inQuote
    } else if (char === ',' && !inQuote) {
      cells.push(current.trim())
      current = ''
    } else {
      current += char
    }
  }
  cells.push(current.trim())
  return cells
}

function parseCsv(text) {
  return String(text || '')
    .split(/\r?\n/)
    .map(line => line.trim())
    .filter(Boolean)
    .map(splitCsvLine)
}

function findIndex(headers, names) {
  return headers.findIndex(header => names.some(name => header.includes(name)))
}

function parseAmountFen(value) {
  const match = String(value || '').replace(/[￥¥,\s]/g, '').match(/-?\d+(\.\d{1,2})?/)
  if (!match) return 0
  return Math.round(Math.abs(Number(match[0])) * 100)
}

function detectType(row, headers, amountIndex, explicitTypeIndex) {
  if (explicitTypeIndex >= 0) {
    const value = row[explicitTypeIndex] || ''
    if (value.includes('收入') || value.includes('入账') || value.includes('收款')) return 'income'
    if (value.includes('支出') || value.includes('付款') || value.includes('消费')) return 'expense'
  }
  const amount = String(row[amountIndex] || '')
  return amount.trim().startsWith('-') ? 'expense' : 'income'
}

function normalizeDate(value) {
  const match = String(value || '').match(/(20\d{2})[-\/年.](\d{1,2})[-\/月.](\d{1,2})/)
  if (!match) return ''
  const year = match[1]
  const month = match[2].padStart(2, '0')
  const day = match[3].padStart(2, '0')
  return `${year}-${month}-${day}`
}

function csvToBills(text) {
  const rows = parseCsv(text)
  if (rows.length < 2) return []
  const headers = rows[0]
  const amountIndex = findIndex(headers, ['金额', '交易金额', '金额(元)', 'amount'])
  const typeIndex = findIndex(headers, ['类型', '收支', '交易类型', '收/支'])
  const dateIndex = findIndex(headers, ['时间', '日期', '交易时间', '创建时间'])
  const categoryIndex = findIndex(headers, ['分类', '交易分类'])
  const noteIndex = findIndex(headers, ['备注', '说明', '商品', '交易对方', '名称'])
  if (amountIndex < 0 || dateIndex < 0) return []
  const now = Date.now()
  return rows.slice(1).map((row, index) => {
    const amountFen = parseAmountFen(row[amountIndex])
    const date = normalizeDate(row[dateIndex])
    if (!amountFen || !date) return null
    const type = detectType(row, headers, amountIndex, typeIndex)
    return {
      id: `csv_${now}_${index}`,
      type,
      amountFen,
      category: row[categoryIndex] || '导入',
      account: '导入账单',
      date,
      note: row[noteIndex] || '',
      createdAt: now + index,
      updatedAt: now + index
    }
  }).filter(Boolean)
}

module.exports = {
  csvToBills
}


