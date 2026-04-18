var api = require('../../utils/api')

Page({
  data: { list: [], page: 1, total: 0, loading: false, category: '', keyword: '' },
  onShow: function() { this.loadData(true) },
  loadData: function(reset) {
    var that = this
    if (reset) this.setData({ page: 1, list: [] })
    this.setData({ loading: true })
    api.notices({ page: this.data.page, size: 20, category: this.data.category, q: this.data.keyword }, function(res) {
      that.setData({
        list: reset ? (res.data || []) : that.data.list.concat(res.data || []),
        total: res.total,
        loading: false
      })
    })
  },
  onReachBottom: function() {
    if (this.data.list.length < this.data.total) {
      this.setData({ page: this.data.page + 1 })
      this.loadData(false)
    }
  },
  onPullDownRefresh: function() {
    var that = this
    this.loadData(true)
    setTimeout(function() { wx.stopPullDownRefresh() }, 500)
  },
  filterCat: function(e) {
    this.setData({ category: e.currentTarget.dataset.cat })
    this.loadData(true)
  },
  onSearch: function(e) {
    this.setData({ keyword: e.detail.value })
  },
  doSearch: function() {
    this.loadData(true)
  },
  goDetail: function(e) {
    wx.navigateTo({ url: '/pages/notice-detail/notice-detail?id=' + e.currentTarget.dataset.id })
  },
  onShareAppMessage: function() {
    return { title: (getApp().globalData.villageName || '村务') + ' · 村务公告', path: '/pages/notices/notices' }
  }
})
