var api = require('../../utils/api')

Page({
  data: { notice: null },
  onLoad: function(opts) { this.loadData(opts.id) },
  loadData: function(id) {
    var that = this
    api.notice(id, function(res) {
      that.setData({ notice: res.notice || res })
    })
  },
  onShareAppMessage: function() {
    var n = this.data.notice
    return {
      title: n ? n.title : (getApp().globalData.villageName || '村务') + ' · 公告',
      path: '/pages/notice-detail/notice-detail?id=' + (n ? n.id : '')
    }
  }
})
