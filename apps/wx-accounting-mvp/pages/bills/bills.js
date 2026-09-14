const storage = require('../../utils/storage')
const format = require('../../utils/format')
const cloud = require('../../utils/cloud')

Page({
  data: {
    monthKey: '',
    monthLabel: '',
    bills: [],
    hasBills: false,
    syncing: false
  },

  onShow() {
    if (!this.data.monthKey) this.setData({ monthKey: format.toMonthKey() })
    this.refresh()
    if (cloud.isEnabled()) this.syncCloud()
  },

  refresh() {
    const bills = storage.getBillsByMonth(this.data.monthKey).map(item => Object.assign({}, item, {
      amountText: format.fenToYuan(item.amountFen)
    }))
    this.setData({ monthLabel: format.getMonthLabel(this.data.monthKey), bills, hasBills: bills.length > 0 })
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

  goAdd() { wx.navigateTo({ url: '/pages/add/add?type=expense' }) },
  edit(event) { wx.navigateTo({ url: `/pages/add/add?id=${event.currentTarget.dataset.id}` }) },

  remove(event) {
    const id = event.currentTarget.dataset.id
    wx.showModal({
      title: '删除这笔账单？',
      content: '删除后无法恢复。',
      confirmColor: '#ef4444',
      success: res => {
        if (!res.confirm) return
        const remove = cloud.isEnabled() ? cloud.removeBill(id) : Promise.resolve(storage.removeBill(id))
        remove.then(() => {
          this.refresh()
          wx.showToast({ title: '已删除', icon: 'success' })
        }).catch(error => wx.showToast({ title: error.message || '删除失败', icon: 'none' }))
      }
    })
  }
})
