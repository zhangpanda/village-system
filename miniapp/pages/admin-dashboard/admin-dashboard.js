var api = require('../../utils/api')

Page({
  data: { stats: null },
  onShow: function() { this.loadData() },
  loadData: function() {
    var that = this
    api.dashboard(function(res) { that.setData({ stats: res }) })
  },
  goTo: function(e) {
    wx.navigateTo({ url: e.currentTarget.dataset.url })
  }
})
