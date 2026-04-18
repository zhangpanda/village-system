var api = require('../../utils/api')

Page({
  data: { list: [], loading: true },
  onShow: function() { this.loadData() },
  loadData: function() {
    var that = this
    api.notifications({ size: 50 }, function(res) {
      that.setData({ list: res.data || [], loading: false })
    })
  },
  markAll: function() {
    var that = this
    api.markAllRead(function() {
      wx.showToast({ title: '全部已读' })
      that.loadData()
    })
  },
  markOne: function(e) {
    var that = this
    api.markRead(e.currentTarget.dataset.id, function() { that.loadData() })
  }
})
