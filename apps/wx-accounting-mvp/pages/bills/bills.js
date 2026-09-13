const storage = require('../../utils/storage')
const format = require('../../utils/format')

Page({
  data: {
    monthKey: '',
    monthLabel: '',
    bills: [],
    hasBills: false
  },

  onShow() {
    if (!this.data.monthKey) this.setData({ monthKey: format.toMonthKey() })
    this.refresh()
  },

  refresh() {
    const bills = storage.getBillsByMonth(this.data.monthKey).map(item => Object.assign({}, item, {
      amountText: format.fenToYuan(item.amountFen)
    }))
    this.setData({
      monthLabel: format.getMonthLabel(this.data.monthKey),
      bills,
      hasBills: bills.length > 0
    })
  },

  onMonthChange(event) {
    this.setData({ monthKey: event.detail.value })
    this.refresh()
  },

  goAdd() {
    wx.navigateTo({ url: '/pages/add/add?type=expense' })
  },

  edit(event) {
    wx.navigateTo({ url: `/pages/add/add?id=${event.currentTarget.dataset.id}` })
  },

  remove(event) {
    const id = event.currentTarget.dataset.id
    wx.showModal({
      title: '删除这笔账单？',
      content: '删除后无法恢复。',
      confirmColor: '#ef4444',
      success: res => {
        if (res.confirm) {
          storage.removeBill(id)
          this.refresh()
          wx.showToast({ title: '已删除', icon: 'success' })
        }
      }
    })
  }
})
