var api = require('../../utils/api')

Page({
  data: { list: [], page: 1, total: 0, status: '' },
  onShow: function() {
    if (!getApp().globalData.token) {
      wx.showToast({ title: '请先登录', icon: 'none' }); return
    }
    this.loadData(true)
  },
  loadData: function(reset) {
    var that = this
    if (reset) this.setData({ page: 1, list: [] })
    api.tickets({ page: this.data.page, size: 20, state: this.data.status, mine: 1 }, function(res) {
      that.setData({ list: reset ? (res.data || []) : that.data.list.concat(res.data || []), total: res.total })
    })
  },
  filterStatus: function(e) {
    this.setData({ status: e.currentTarget.dataset.status })
    this.loadData(true)
  },
  onPullDownRefresh: function() {
    this.loadData(true)
    setTimeout(function() { wx.stopPullDownRefresh() }, 500)
  },
  onReachBottom: function() {
    if (this.data.list.length < this.data.total) {
      this.setData({ page: this.data.page + 1 })
      this.loadData(false)
    }
  },
  goCreate: function() { wx.navigateTo({ url: '/pages/ticket-create/ticket-create' }) },
  goDetail: function(e) { wx.navigateTo({ url: '/pages/ticket-detail/ticket-detail?id=' + e.currentTarget.dataset.id }) }
})
