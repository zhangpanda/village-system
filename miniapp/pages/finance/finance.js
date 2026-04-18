var api = require('../../utils/api')

var currentYear = new Date().getFullYear()

Page({
  data: { summary: null, list: [], page: 1, total: 0, year: '' + currentYear, type: '', years: [] },
  onLoad: function() {
    var arr = []
    for (var y = currentYear; y >= currentYear - 5; y--) arr.push('' + y)
    this.setData({ years: arr })
  },
  onShow: function() { this.loadAll() },
  loadAll: function() {
    var that = this
    api.financeSummary({ year: this.data.year }, function(sum) {
      that.setData({ summary: sum })
    })
    this.setData({ page: 1, list: [] })
    this.loadPage()
  },
  loadPage: function() {
    var that = this
    api.finance({ page: this.data.page, size: 20, year: this.data.year, type: this.data.type }, function(res) {
      that.setData({ list: that.data.list.concat(res.data || []), total: res.total })
    })
  },
  onReachBottom: function() {
    if (this.data.list.length < this.data.total) {
      this.setData({ page: this.data.page + 1 })
      this.loadPage()
    }
  },
  onPullDownRefresh: function() {
    this.loadAll()
    setTimeout(function() { wx.stopPullDownRefresh() }, 500)
  },
  filterType: function(e) {
    this.setData({ type: e.currentTarget.dataset.type })
    this.loadAll()
  },
  onYearPick: function(e) {
    this.setData({ year: this.data.years[e.detail.value] })
    this.loadAll()
  },
  onShareAppMessage: function() {
    return { title: (getApp().globalData.villageName || '村务') + ' · 财务公示', path: '/pages/finance/finance' }
  }
})
