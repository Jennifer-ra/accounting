const DEFAULT_CATEGORIES = {
  expense: ['餐饮', '交通', '购物', '住房', '娱乐', '医疗', '学习', '人情', '其他'],
  income: ['工资', '奖金', '副业', '理财', '报销', '其他']
}

const DEFAULT_ACCOUNTS = ['默认账户', '微信', '支付宝', '银行卡', '现金']
const STORAGE_KEY = 'wx_accounting_mvp_bills'
const CONFIG_KEY = 'wx_accounting_mvp_config'
const VERSION_KEY = 'wx_accounting_mvp_version'

function cloneDefaultCategories() {
  return {
    expense: DEFAULT_CATEGORIES.expense.slice(),
    income: DEFAULT_CATEGORIES.income.slice()
  }
}

function normalizeList(list, fallback) {
  const source = Array.isArray(list) && list.length ? list : fallback
  return Array.from(new Set(source.map(item => String(item || '').trim()).filter(Boolean)))
}

function readConfig() {
  const saved = wx.getStorageSync(CONFIG_KEY) || {}
  const categories = Object.assign(cloneDefaultCategories(), saved.categories || {})
  return {
    categories: {
      expense: normalizeList(categories.expense, DEFAULT_CATEGORIES.expense),
      income: normalizeList(categories.income, DEFAULT_CATEGORIES.income)
    },
    accounts: normalizeList(saved.accounts, DEFAULT_ACCOUNTS),
    monthlyBudgetFen: Number.isInteger(saved.monthlyBudgetFen) ? saved.monthlyBudgetFen : 0
  }
}

function writeConfig(config) {
  wx.setStorageSync(CONFIG_KEY, config)
  wx.setStorageSync(VERSION_KEY, Date.now())
}

function replaceConfig(config) {
  const current = readConfig()
  const next = Object.assign({}, current)
  if (config && config.categories) next.categories = Object.assign({}, current.categories, config.categories)
  if (config && Array.isArray(config.accounts)) next.accounts = normalizeList(config.accounts, DEFAULT_ACCOUNTS)
  if (config && Number.isInteger(config.monthlyBudgetFen)) next.monthlyBudgetFen = config.monthlyBudgetFen
  writeConfig(next)
}

function getCategories(type) {
  return readConfig().categories[type] || readConfig().categories.expense
}

function getAccounts() {
  return readConfig().accounts
}

function addCategory(type, name) {
  const cleanName = String(name || '').trim()
  if (!cleanName) return false
  const config = readConfig()
  const key = type === 'income' ? 'income' : 'expense'
  if (!config.categories[key].includes(cleanName)) config.categories[key].push(cleanName)
  writeConfig(config)
  return true
}

function addAccount(name) {
  const cleanName = String(name || '').trim()
  if (!cleanName) return false
  const config = readConfig()
  if (!config.accounts.includes(cleanName)) config.accounts.push(cleanName)
  writeConfig(config)
  return true
}

function setMonthlyBudget(amountFen) {
  const config = readConfig()
  config.monthlyBudgetFen = Math.max(0, amountFen || 0)
  writeConfig(config)
}

function normalizeBill(item) {
  const id = String(item.id || `${Date.now()}_${Math.random().toString(16).slice(2)}`)
  return {
    id,
    type: item.type === 'income' ? 'income' : 'expense',
    amountFen: Number(item.amountFen) || 0,
    category: item.category || '其他',
    account: item.account || '默认账户',
    date: item.date || '',
    note: item.note || '',
    serverId: item.serverId ? String(item.serverId) : '',
    clientId: item.clientId || id,
    source: item.source || 'manual',
    createdAt: item.createdAt || Date.now(),
    updatedAt: item.updatedAt || item.createdAt || Date.now()
  }
}

function sortBills(bills) {
  return bills.slice().sort((a, b) => {
    if (a.date === b.date) return b.createdAt - a.createdAt
    return a.date < b.date ? 1 : -1
  })
}

function readBills() {
  const bills = wx.getStorageSync(STORAGE_KEY)
  return Array.isArray(bills) ? sortBills(bills.map(normalizeBill)) : []
}

function writeBills(bills) {
  wx.setStorageSync(STORAGE_KEY, sortBills(bills.map(normalizeBill)))
  wx.setStorageSync(VERSION_KEY, Date.now())
}

function addBill(input) {
  const now = Date.now()
  const bill = normalizeBill(Object.assign({}, input, {
    id: `${now}_${Math.random().toString(16).slice(2)}`,
    createdAt: now,
    updatedAt: now
  }))
  writeBills([bill].concat(readBills()))
  return bill
}

function findBill(id) {
  return readBills().find(item => item.id === id)
}

function upsertBill(bill) {
  const next = normalizeBill(bill)
  const bills = readBills().filter(item => item.id !== next.id && item.clientId !== next.clientId)
  writeBills([next].concat(bills))
  return next
}

function updateBill(id, patch) {
  const now = Date.now()
  const bills = readBills().map(item => item.id === id ? normalizeBill(Object.assign({}, item, patch, { updatedAt: now })) : item)
  writeBills(bills)
}

function removeBill(id) {
  writeBills(readBills().filter(item => item.id !== id && item.serverId !== String(id)))
}

function clearBills() {
  writeBills([])
}

function replaceBills(bills) {
  const cleanBills = bills
    .filter(item => item && item.id && item.type && Number.isInteger(Number(item.amountFen)) && item.date)
    .map(normalizeBill)
  writeBills(cleanBills)
}

function mergeBills(bills) {
  const map = {}
  readBills().concat(bills.map(normalizeBill)).forEach(item => {
    map[item.clientId || item.id] = item
  })
  writeBills(Object.keys(map).map(id => map[id]))
}

function getBillsByMonth(monthKey) {
  return readBills().filter(item => item.date && item.date.startsWith(monthKey))
}

function getSummary(monthKey) {
  const monthBills = getBillsByMonth(monthKey)
  const summary = monthBills.reduce((acc, item) => {
    if (item.type === 'income') acc.income += item.amountFen
    else acc.expense += item.amountFen
    return acc
  }, { income: 0, expense: 0 })
  summary.balance = summary.income - summary.expense
  summary.count = monthBills.length
  const budget = readConfig().monthlyBudgetFen
  summary.budget = budget
  summary.budgetLeft = budget ? budget - summary.expense : 0
  summary.budgetPercent = budget ? Math.min(100, Math.round(summary.expense / budget * 100)) : 0
  return summary
}

function groupByCategory(monthKey, type) {
  const map = {}
  getBillsByMonth(monthKey)
    .filter(item => item.type === type)
    .forEach(item => { map[item.category] = (map[item.category] || 0) + item.amountFen })
  return Object.keys(map)
    .map(name => ({ name, amountFen: map[name] }))
    .sort((a, b) => b.amountFen - a.amountFen)
}

module.exports = {
  readConfig,
  replaceConfig,
  getCategories,
  getAccounts,
  addCategory,
  addAccount,
  setMonthlyBudget,
  readBills,
  addBill,
  findBill,
  upsertBill,
  updateBill,
  removeBill,
  clearBills,
  replaceBills,
  mergeBills,
  getBillsByMonth,
  getSummary,
  groupByCategory
}
