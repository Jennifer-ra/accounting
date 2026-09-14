const storage = require('../../utils/storage')
const format = require('../../utils/format')
const importer = require('../../utils/importer')
const cloud = require('../../utils/cloud')

Page({
  data: {
    billCount: 0,
    categories: {},
    accounts: [],
    categoryTypeIndex: 0,
    categoryTypes: ['支出分类', '收入分类'],
    newCategory: '',
    newAccount: '',
    monthlyBudget: '',
    cloudEnabled: false,
    apiBaseUrl: '',
    syncing: false
  },

  onShow() { this.refresh() },

  refresh() {
    const config = storage.readConfig()
    const cloudSettings = cloud.readSettings()
    this.setData({
      billCount: storage.readBills().length,
      categories: config.categories,
      accounts: config.accounts,
      monthlyBudget: config.monthlyBudgetFen ? format.fenToYuan(config.monthlyBudgetFen) : '',
      cloudEnabled: cloudSettings.enabled,
      apiBaseUrl: cloudSettings.apiBaseUrl,
      syncing: false
    })
  },

  onInput(event) {
    const field = event.currentTarget.dataset.field
    this.setData({ [field]: event.detail.value })
  },

  onCloudModeChange(event) {
    const enabled = event.detail.value
    cloud.setEnabled(enabled)
    this.setData({ cloudEnabled: enabled })
    if (enabled) this.loginAndSync()
  },

  saveApiBaseUrl() {
    cloud.setApiBaseUrl(this.data.apiBaseUrl)
    this.refresh()
    wx.showToast({ title: '地址已保存', icon: 'success' })
  },

  loginAndSync() {
    this.setData({ syncing: true })
    cloud.login()
      .then(() => cloud.syncMonth(format.toMonthKey()))
      .then(() => {
        this.refresh()
        wx.showToast({ title: '云端已连接', icon: 'success' })
      })
      .catch(error => {
        cloud.setEnabled(false)
        this.refresh()
        wx.showToast({ title: error.message || '连接失败', icon: 'none' })
      })
  },

  onCategoryTypeChange(event) { this.setData({ categoryTypeIndex: Number(event.detail.value) }) },

  addCategory() {
    const type = this.data.categoryTypeIndex === 1 ? 'income' : 'expense'
    const save = cloud.isEnabled() ? cloud.addCategory(type, this.data.newCategory) : Promise.resolve(storage.addCategory(type, this.data.newCategory))
    save.then(result => {
      if (result === false) {
        wx.showToast({ title: '请输入分类名称', icon: 'none' })
        return
      }
      this.setData({ newCategory: '' })
      this.refresh()
      wx.showToast({ title: '已添加', icon: 'success' })
    }).catch(error => wx.showToast({ title: error.message || '添加失败', icon: 'none' }))
  },

  addAccount() {
    const save = cloud.isEnabled() ? cloud.addAccount(this.data.newAccount) : Promise.resolve(storage.addAccount(this.data.newAccount))
    save.then(result => {
      if (result === false) {
        wx.showToast({ title: '请输入账户名称', icon: 'none' })
        return
      }
      this.setData({ newAccount: '' })
      this.refresh()
      wx.showToast({ title: '已添加', icon: 'success' })
    }).catch(error => wx.showToast({ title: error.message || '添加失败', icon: 'none' }))
  },

  saveBudget() {
    const amountFen = this.data.monthlyBudget ? format.yuanToFen(this.data.monthlyBudget) : 0
    if (amountFen === null) {
      wx.showToast({ title: '预算金额不正确', icon: 'none' })
      return
    }
    const monthKey = format.toMonthKey()
    const save = cloud.isEnabled() ? cloud.setMonthlyBudget(monthKey, amountFen) : Promise.resolve(storage.setMonthlyBudget(amountFen))
    save.then(() => {
      this.refresh()
      wx.showToast({ title: '预算已保存', icon: 'success' })
    }).catch(error => wx.showToast({ title: error.message || '保存失败', icon: 'none' }))
  },

  exportData() {
    const payload = {
      app: 'wx-accounting-mvp',
      version: 3,
      exportedAt: new Date().toISOString(),
      cloud: cloud.readSettings(),
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
