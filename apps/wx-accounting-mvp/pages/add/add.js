const storage = require('../../utils/storage')
const format = require('../../utils/format')
const cloud = require('../../utils/cloud')

Page({
  data: {
    mode: 'create',
    billId: '',
    title: '记一笔',
    type: 'expense',
    typeIndex: 0,
    typeOptions: ['支出', '收入'],
    amount: '',
    categoryIndex: 0,
    categories: [],
    accountIndex: 0,
    accounts: [],
    date: '',
    note: '',
    saving: false
  },

  onLoad(options) {
    if (options.id) {
      this.loadBill(options.id)
      return
    }
    const type = options.type === 'income' ? 'income' : 'expense'
    this.setType(type)
    this.setData({ accounts: storage.getAccounts(), date: format.toDateKey() })
  },

  loadBill(id) {
    const bill = storage.findBill(id)
    if (!bill) {
      wx.showToast({ title: '账单不存在', icon: 'none' })
      setTimeout(() => wx.navigateBack(), 500)
      return
    }
    const categories = storage.getCategories(bill.type)
    const accounts = storage.getAccounts()
    this.setData({
      mode: 'edit',
      billId: bill.id,
      title: '编辑账单',
      type: bill.type,
      typeIndex: bill.type === 'income' ? 1 : 0,
      amount: format.fenToYuan(bill.amountFen),
      categories,
      categoryIndex: Math.max(0, categories.indexOf(bill.category)),
      accounts,
      accountIndex: Math.max(0, accounts.indexOf(bill.account)),
      date: bill.date,
      note: bill.note
    })
    wx.setNavigationBarTitle({ title: '编辑账单' })
  },

  setType(type) {
    this.setData({
      type,
      typeIndex: type === 'income' ? 1 : 0,
      categories: storage.getCategories(type),
      categoryIndex: 0
    })
  },

  onTypeChange(event) { this.setType(Number(event.detail.value) === 1 ? 'income' : 'expense') },
  onCategoryChange(event) { this.setData({ categoryIndex: Number(event.detail.value) }) },
  onAccountChange(event) { this.setData({ accountIndex: Number(event.detail.value) }) },
  onDateChange(event) { this.setData({ date: event.detail.value }) },

  onInput(event) {
    const field = event.currentTarget.dataset.field
    this.setData({ [field]: event.detail.value })
  },

  submit() {
    if (this.data.saving) return
    const amountFen = format.yuanToFen(this.data.amount)
    if (!amountFen || amountFen <= 0) {
      wx.showToast({ title: '请输入正确金额', icon: 'none' })
      return
    }
    const payload = {
      type: this.data.type,
      amountFen,
      category: this.data.categories[this.data.categoryIndex],
      account: this.data.accounts[this.data.accountIndex],
      date: this.data.date,
      note: this.data.note.trim()
    }
    this.setData({ saving: true })
    let save
    if (cloud.isEnabled()) {
      save = this.data.mode === 'edit' ? cloud.updateBill(this.data.billId, payload) : cloud.createBill(payload)
    } else {
      if (this.data.mode === 'edit') storage.updateBill(this.data.billId, payload)
      else storage.addBill(payload)
      save = Promise.resolve()
    }
    save.then(() => {
      wx.showToast({ title: this.data.mode === 'edit' ? '已更新' : '已记录', icon: 'success' })
      setTimeout(() => wx.navigateBack(), 450)
    }).catch(error => wx.showToast({ title: error.message || '保存失败', icon: 'none' }))
      .finally(() => this.setData({ saving: false }))
  }
})
