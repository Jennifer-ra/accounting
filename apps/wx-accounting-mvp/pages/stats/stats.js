const storage = require('../../utils/storage')
const format = require('../../utils/format')
const cloud = require('../../utils/cloud')

Page({
  data: {
    monthKey: '',
    monthLabel: '',
    summary: {},
    expenseRanks: [],
    incomeRanks: [],
    syncing: false
  },

  onShow() {
    if (!this.data.monthKey) this.setData({ monthKey: format.toMonthKey() })
    this.refresh()
    if (cloud.isEnabled()) this.syncCloud()
  },

  refresh() {
    const expenseRanks = storage.groupByCategory(this.data.monthKey, 'expense')
    const incomeRanks = storage.groupByCategory(this.data.monthKey, 'income')
    const maxExpense = expenseRanks[0] ? expenseRanks[0].amountFen : 0
    const maxIncome = incomeRanks[0] ? incomeRanks[0].amountFen : 0
    this.setData({
      monthLabel: format.getMonthLabel(this.data.monthKey),
      summary: this.formatSummary(storage.getSummary(this.data.monthKey)),
      expenseRanks: expenseRanks.map(item => this.formatRank(item, maxExpense)),
      incomeRanks: incomeRanks.map(item => this.formatRank(item, maxIncome))
    })
  },

  syncCloud() {
    this.setData({ syncing: true })
    cloud.syncMonth(this.data.monthKey)
      .then(() => this.refresh())
      .catch(error => wx.showToast({ title: error.message || '同步失败', icon: 'none' }))
      .finally(() => this.setData({ syncing: false }))
  },

  onMonthChange(event) {
    this.setData({ monthKey: event.detail.value })
    this.refresh()
    if (cloud.isEnabled()) this.syncCloud()
  },

  formatSummary(summary) {
    return {
      income: format.fenToYuan(summary.income),
      expense: format.fenToYuan(summary.expense),
      balance: format.fenToYuan(summary.balance),
      budget: format.fenToYuan(summary.budget),
      budgetLeft: format.fenToYuan(summary.budgetLeft),
      budgetPercent: summary.budgetPercent,
      hasBudget: summary.budget > 0,
      budgetOver: summary.budgetLeft < 0,
      count: summary.count
    }
  },

  formatRank(item, max) {
    return Object.assign({}, item, {
      amountText: format.fenToYuan(item.amountFen),
      percent: max ? Math.max(8, Math.round(item.amountFen / max * 100)) : 0
    })
  }
})
