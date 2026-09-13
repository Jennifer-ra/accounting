const storage = require('../../utils/storage')
const format = require('../../utils/format')

Page({
  data: {
    monthKey: '',
    monthLabel: '',
    summary: {},
    recentBills: [],
    hasBills: false,
    showQuickAdd: false,
    quickType: 'expense',
    quickTitle: '记支出',
    quickAmount: '',
    quickDate: '',
    quickNote: '',
    quickCategoryIndex: 0,
    quickAccountIndex: 0,
    quickCategories: [],
    accounts: []
  },

  onShow() {
    this.refresh()
    if (this.data.showQuickAdd) this.prepareQuickAdd(this.data.quickType)
  },

  refresh() {
    const monthKey = format.toMonthKey()
    const bills = storage.readBills()
    this.setData({
      monthKey,
      monthLabel: format.getMonthLabel(monthKey),
      summary: this.formatSummary(storage.getSummary(monthKey)),
      recentBills: bills.slice(0, 5).map(this.formatBill),
      hasBills: bills.length > 0
    })
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

  formatBill(item) {
    return Object.assign({}, item, {
      amountText: format.fenToYuan(item.amountFen),
      typeText: item.type === 'income' ? '收入' : '支出'
    })
  },

  prepareQuickAdd(type) {
    this.setData({
      quickType: type,
      quickTitle: type === 'income' ? '记收入' : '记支出',
      quickCategories: storage.getCategories(type),
      accounts: storage.getAccounts(),
      quickCategoryIndex: 0,
      quickAccountIndex: 0,
      quickDate: format.toDateKey(),
      quickAmount: '',
      quickNote: ''
    })
  },

  goAddExpense() {
    this.prepareQuickAdd('expense')
    this.setData({ showQuickAdd: true })
  },

  goAddIncome() {
    this.prepareQuickAdd('income')
    this.setData({ showQuickAdd: true })
  },

  closeQuickAdd() {
    this.setData({ showQuickAdd: false })
  },

  noop() {},

  onQuickInput(event) {
    const field = event.currentTarget.dataset.field
    this.setData({ [field]: event.detail.value })
  },

  onQuickCategoryChange(event) {
    this.setData({ quickCategoryIndex: Number(event.detail.value) })
  },

  onQuickAccountChange(event) {
    this.setData({ quickAccountIndex: Number(event.detail.value) })
  },

  onQuickDateChange(event) {
    this.setData({ quickDate: event.detail.value })
  },

  submitQuickAdd() {
    const amountFen = format.yuanToFen(this.data.quickAmount)
    if (!amountFen || amountFen <= 0) {
      wx.showToast({ title: '请输入正确金额', icon: 'none' })
      return
    }
    storage.addBill({
      type: this.data.quickType,
      amountFen,
      category: this.data.quickCategories[this.data.quickCategoryIndex],
      account: this.data.accounts[this.data.quickAccountIndex],
      date: this.data.quickDate,
      note: this.data.quickNote.trim()
    })
    this.setData({ showQuickAdd: false })
    this.refresh()
    wx.showToast({ title: '已记录', icon: 'success' })
  },

  goFullAdd() {
    wx.navigateTo({ url: `/pages/add/add?type=${this.data.quickType}` })
  },

  goBills() {
    wx.switchTab({ url: '/pages/bills/bills' })
  }
})
