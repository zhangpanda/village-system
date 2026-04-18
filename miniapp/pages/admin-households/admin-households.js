var api = require('../../utils/api')

Page({
  data: { list: [], total: 0, page: 1 },
  onShow: function() { this.loadData() },
  loadData: function() {
    var that = this
    api.adminHouseholds({ page: this.data.page, size: 20 }, function(res) {
      that.setData({ list: res.data || [], total: res.total })
    })
  },
  viewDetail: function(e) {
    wx.navigateTo({ url: '/pages/admin-household-detail/admin-household-detail?id=' + e.currentTarget.dataset.id })
  }
})
