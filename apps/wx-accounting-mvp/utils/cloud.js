const storage = require('./storage')
const format = require('./format')

const CLOUD_KEY = 'wx_accounting_mvp_cloud_settings'
const TOKEN_KEY = 'wx_accounting_mvp_token'
const DEFAULT_API_BASE_URL = 'http://localhost:10240/api/v1'

function readSettings() {
  const saved = wx.getStorageSync(CLOUD_KEY) || {}
  return {
    enabled: Boolean(saved.enabled),
    apiBaseUrl: saved.apiBaseUrl || DEFAULT_API_BASE_URL,
    lastSyncAt: saved.lastSyncAt || 0
  }
}

function writeSettings(settings) {
  wx.setStorageSync(CLOUD_KEY, Object.assign(readSettings(), settings || {}))
}

function isEnabled() { return readSettings().enabled }
function setEnabled(enabled) { writeSettings({ enabled: Boolean(enabled) }) }

function setApiBaseUrl(apiBaseUrl) {
  const cleanUrl = String(apiBaseUrl || '').trim().replace(/\/+$/, '')
  writeSettings({ apiBaseUrl: cleanUrl || DEFAULT_API_BASE_URL })
}

function getToken() { return wx.getStorageSync(TOKEN_KEY) || '' }
function setToken(token) { wx.setStorageSync(TOKEN_KEY, token || '') }

function request(options) {
  const settings = readSettings()
  return new Promise((resolve, reject) => {
    wx.request({
      url: settings.apiBaseUrl + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: Object.assign({
        'Content-Type': 'application/json',
        Authorization: getToken() ? `Bearer ${getToken()}` : ''
      }, options.header || {}),
      success: res => {
        if (res.statusCode >= 200 && res.statusCode < 300 && res.data && res.data.code === 'ok') {
          resolve(res.data.data)
          return
        }
        reject(new Error((res.data && res.data.message) || `HTTP ${res.statusCode}`))
      },
      fail: reject
    })
  })
}

function login() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: res => {
        if (!res.code) { reject(new Error('wx.login 未返回 code')); return }
        request({ url: '/auth/wechat-login', method: 'POST', data: { code: res.code } })
          .then(data => { setToken(data.token); resolve(data) })
          .catch(reject)
      },
      fail: reject
    })
  })
}

function ensureLogin() {
  if (getToken()) return Promise.resolve()
  return login().then(() => undefined)
}

function isServerBacked(item) {
  return /^\d+$/.test(String(item.serverId || item.id || ''))
}

function toLocalBill(item) {
  return {
    id: String(item.id),
    serverId: String(item.id),
    clientId: item.clientId || String(item.id),
    type: item.type,
    amountFen: item.amountFen,
    category: item.category,
    account: item.account,
    date: item.date,
    note: item.note || '',
    source: item.source || 'manual',
    createdAt: Date.parse(item.createdAt) || Date.now(),
    updatedAt: Date.parse(item.updatedAt) || Date.now()
  }
}

function pushLocalConfig(monthKey) {
  const config = storage.readConfig()
  const tasks = []
  ;['expense', 'income'].forEach(type => {
    ;(config.categories[type] || []).forEach(name => tasks.push(request({ url: '/categories', method: 'POST', data: { type, name } })))
  })
  ;(config.accounts || []).forEach(name => tasks.push(request({ url: '/accounts', method: 'POST', data: { name } })))
  if (config.monthlyBudgetFen > 0) {
    tasks.push(request({ url: `/budgets/${monthKey}`, method: 'PUT', data: { amountFen: config.monthlyBudgetFen } }))
  }
  return Promise.all(tasks.map(task => task.catch(error => ({ error })))).then(() => undefined)
}

function pushLocalBills(monthKey) {
  const bills = storage.getBillsByMonth(monthKey).filter(item => !isServerBacked(item))
  let chain = Promise.resolve()
  bills.forEach(item => {
    chain = chain.then(() => createBill({
      clientId: item.clientId || item.id,
      type: item.type,
      amountFen: item.amountFen,
      category: item.category,
      account: item.account,
      date: item.date,
      note: item.note,
      source: item.source || 'manual'
    }))
  })
  return chain
}

function fetchConfig(monthKey) {
  return request({ url: `/bootstrap?month=${monthKey || format.toMonthKey()}` })
    .then(data => {
      storage.replaceConfig({
        categories: data.categories,
        accounts: data.accounts,
        monthlyBudgetFen: data.monthlyBudgetFen || 0
      })
      return data
    })
}

function fetchTransactions(monthKey) {
  return request({ url: `/transactions${monthKey ? `?month=${monthKey}` : ''}` })
    .then(data => {
      const remoteBills = (data.items || []).map(toLocalBill)
      if (monthKey) {
        const otherBills = storage.readBills().filter(item => !item.date || !item.date.startsWith(monthKey))
        storage.replaceBills(otherBills.concat(remoteBills))
      } else {
        storage.replaceBills(remoteBills)
      }
      writeSettings({ lastSyncAt: Date.now() })
      return remoteBills
    })
}

function syncMonth(monthKey) {
  if (!isEnabled()) return Promise.resolve()
  const month = monthKey || format.toMonthKey()
  return ensureLogin()
    .then(() => pushLocalConfig(month))
    .then(() => fetchConfig(month))
    .then(() => pushLocalBills(month))
    .then(() => fetchTransactions(month))
}

function createBill(input) {
  const clientId = input.clientId || `${Date.now()}_${Math.random().toString(16).slice(2)}`
  return ensureLogin()
    .then(() => request({ url: '/transactions', method: 'POST', data: Object.assign({}, input, { clientId, source: input.source || 'manual' }) }))
    .then(item => storage.upsertBill(toLocalBill(item)))
}

function updateBill(id, input) {
  const local = storage.findBill(id)
  const serverId = local && local.serverId ? local.serverId : id
  return ensureLogin()
    .then(() => request({ url: `/transactions/${serverId}`, method: 'PUT', data: input }))
    .then(item => storage.upsertBill(toLocalBill(item)))
}

function removeBill(id) {
  const local = storage.findBill(id)
  const serverId = local && local.serverId ? local.serverId : id
  return ensureLogin()
    .then(() => request({ url: `/transactions/${serverId}`, method: 'DELETE' }))
    .then(() => storage.removeBill(String(id)))
}

function addCategory(type, name) {
  return ensureLogin()
    .then(() => request({ url: '/categories', method: 'POST', data: { type, name } }))
    .then(() => fetchConfig(format.toMonthKey()))
}

function addAccount(name) {
  return ensureLogin()
    .then(() => request({ url: '/accounts', method: 'POST', data: { name } }))
    .then(() => fetchConfig(format.toMonthKey()))
}

function setMonthlyBudget(monthKey, amountFen) {
  return ensureLogin()
    .then(() => request({ url: `/budgets/${monthKey}`, method: 'PUT', data: { amountFen } }))
    .then(item => { storage.replaceConfig({ monthlyBudgetFen: item.amountFen || 0 }); return item })
}

module.exports = {
  readSettings,
  setEnabled,
  setApiBaseUrl,
  isEnabled,
  login,
  syncMonth,
  createBill,
  updateBill,
  removeBill,
  addCategory,
  addAccount,
  setMonthlyBudget
}
