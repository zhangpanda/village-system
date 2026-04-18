var api = require('../../utils/api')

Page({
  data: { ticket: null, comments: [], commentText: '' },
  onLoad: function(opts) { this.loadData(opts.id) },
  loadData: function(id) {
    var that = this
    api.ticket(id, function(res) {
      that.setData({ ticket: res.ticket, comments: res.comments || [] })
    })
  },
  onInput: function(e) { this.setData({ commentText: e.detail.value }) },
  closeTicket: function() {
    var that = this
    wx.showModal({
      title: '确认关闭', content: '关闭后工单将不再处理',
      success: function(res) {
        if (!res.confirm) return
        api.updateTicketStatus(that.data.ticket.id, { status: 'closed' }, function(r) {
          if (r && r.error) { wx.showToast({ title: r.error, icon: 'none' }); return }
          wx.showToast({ title: '已关闭' })
          that.loadData(that.data.ticket.id)
        })
      }
    })
  },
  sendComment: function() {
    var that = this
    if (!this.data.commentText.trim()) return
    api.addComment(this.data.ticket.id, { content: this.data.commentText }, function() {
      that.setData({ commentText: '' })
      that.loadData(that.data.ticket.id)
    })
  },
  previewImage: function(e) {
    var url = e.currentTarget.dataset.url
    try {
      var imgs = JSON.parse(this.data.ticket.images || '[]')
      wx.previewImage({ current: url, urls: imgs })
    } catch(err) {}
  }
})
