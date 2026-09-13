const storage = require('../../utils/storage')
const format = require('../../utils/format')
const importer = require('../../utils/importer')

Page({
  data: {
    billCount: 0,
    categories: {},
    accounts: [],
    categoryTypeIndex: 0,
    categoryTypes: ['支出分类', '收入分类'],
    newCategory: '',
    newAccount: '',
    monthlyBudget: ''
  },

  onShow() {
    this.refresh()
  },

  refresh() {
    const config = storage.readConfig()
    this.setData({
      billCount: storage.readBills().length,
      categories: config.categories,
      accounts: config.accounts,
      monthlyBudget: config.monthlyBudgetFen ? format.fenToYuan(config.monthlyBudgetFen) : ''
    })
  },

  onInput(event) {
    const field = event.currentTarget.dataset.field
    this.setData({ [field]: event.detail.value })
  },

  onCategoryTypeChange(event) {
    this.setData({ categoryTypeIndex: Number(event.detail.value) })
  },

  addCategory() {
    const type = this.data.categoryTypeIndex === 1 ? 'income' : 'expense'
    if (!storage.addCategory(type, this.data.newCategory)) {
      wx.showToast({ title: '请输入分类名称', icon: 'none' })
      return
    }
    this.setData({ newCategory: '' })
    this.refresh()
    wx.showToast({ title: '已添加', icon: 'success' })
  },

  addAccount() {
    if (!storage.addAccount(this.data.newAccount)) {
      wx.showToast({ title: '请输入账户名称', icon: 'none' })
      return
    }
    this.setData({ newAccount: '' })
    this.refresh()
    wx.showToast({ title: '已添加', icon: 'success' })
  },

  saveBudget() {
    const amountFen = this.data.monthlyBudget ? format.yuanToFen(this.data.monthlyBudget) : 0
    if (amountFen === null) {
      wx.showToast({ title: '预算金额不正确', icon: 'none' })
      return
    }
    storage.setMonthlyBudget(amountFen)
    this.refresh()
    wx.showToast({ title: '预算已保存', icon: 'success' })
  },

  exportData() {
    const payload = {
      app: 'wx-accounting-mvp',
      version: 2,
      exportedAt: new Date().toISOString(),
      config: storage.readConfig(),
      bills: storage.readBills()
    }
    wx.setClipboardData({
      data: JSON.stringify(payload, null, 2),
      success: () => wx.showToast({ title: '已复制', icon: 'success' })
    })
  },

  importData() {
    wx.getClipboardData({
      success: res => {
        let payload
        try {
          payload = JSON.parse(res.data)
        } catch (error) {
          this.importCsv(res.data)
          return
        }
        const bills = Array.isArray(payload) ? payload : payload.bills
        if (!Array.isArray(bills)) {
          wx.showToast({ title: '未找到账单数据', icon: 'none' })
          return
        }
        wx.showModal({
          title: '导入账单？',
          content: `将用 ${bills.length} 条账单覆盖本地数据。`,
          success: modal => {
            if (modal.confirm) {
              storage.replaceBills(bills)
              this.refresh()
              wx.showToast({ title: '导入完成', icon: 'success' })
            }
          }
        })
      }
    })
  },

  importCsv(text) {
    const bills = importer.csvToBills(text)
    if (!bills.length) {
      wx.showToast({ title: '剪贴板不是可识别数据', icon: 'none' })
      return
    }
    wx.showModal({
      title: '导入 CSV 账单？',
      content: `识别到 ${bills.length} 条账单，将合并到本地数据。`,
      success: modal => {
        if (modal.confirm) {
          storage.mergeBills(bills)
          this.refresh()
          wx.showToast({ title: '导入完成', icon: 'success' })
        }
      }
    })
  },

  clearData() {
    wx.showModal({
      title: '清空本地账单？',
      content: '这个操作会删除当前设备上的全部账单。',
      confirmColor: '#ef4444',
      success: res => {
        if (res.confirm) {
          storage.clearBills()
          this.setData({ billCount: 0 })
          wx.showToast({ title: '已清空', icon: 'success' })
        }
      }
    })
  }
})
