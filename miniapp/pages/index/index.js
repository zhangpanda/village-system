var api = require('../../utils/api')

Page({
  data: {
    notices: [],
    loading: true,
    villageName: ''
  },
  onLoad: function() {
    var app = getApp()
    this.setData({ villageName: app.globalData.villageName || '' })
    this.loadData()
    if (!app.globalData.villageName) {
      var that = this
      setTimeout(function() { that.setData({ villageName: app.globalData.villageName || '村务' }) }, 800)
    }
  },
  onPullDownRefresh: function() {
    var that = this
    this.loadData()
    setTimeout(function() { wx.stopPullDownRefresh() }, 500)
  },
  loadData: function() {
    var that = this
    that.setData({ loading: true })
    api.notices({ page: 1, size: 5 }, function(res) {
      that.setData({ notices: res.data || [], loading: false })
    })
  },
  goNotices: function() { wx.switchTab({ url: '/pages/notices/notices' }) },
  goFinance: function() { wx.navigateTo({ url: '/pages/finance/finance' }) },
  goSubsidy: function() { wx.navigateTo({ url: '/pages/subsidy/subsidy' }) },
  goTickets: function() { wx.navigateTo({ url: '/pages/tickets/tickets' }) },
  goNotice: function(e) {
    wx.navigateTo({ url: '/pages/notice-detail/notice-detail?id=' + e.currentTarget.dataset.id })
  },
  onShareAppMessage: function() {
    return {
      title: (getApp().globalData.villageName || '村务') + ' · 村务公开',
      path: '/pages/index/index'
    }
  }
})
