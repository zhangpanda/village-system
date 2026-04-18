var api = require('../../utils/api')

var stateMap = { draft:'草稿',pending_review:'待审核',published:'已发布',submitted:'已提交',committee_review:'村委初审中',secretary_review:'村支书终审中',approved:'已通过',rejected:'已驳回',open:'待处理',assigned:'已分配',processing:'处理中',resolved:'已解决',closed:'已关闭' }
var subTypeMap = { farming:'农业补贴',medical:'医疗救助',education:'教育补助',housing:'住房补贴',other:'其他' }
var ticketCatMap = { repair:'报修',complaint:'投诉',service:'便民服务',suggestion:'建议' }

Page({
  data: { type: '', list: [], total: 0, page: 1, detail: null, note: '', commentText: '' },
  onLoad: function(opts) {
    this.setData({ type: opts.type || 'notice' })
    var titles = { notice:'公告审核', subsidy:'补贴审批', ticket:'工单处理', finance:'财务审核' }
    wx.setNavigationBarTitle({ title: titles[opts.type] || '审核' })
  },
  onShow: function() { this.loadList() },
  loadList: function() {
    var that = this
    var type = this.data.type
    var page = this.data.page
    if (type === 'notice') {
      api.adminNotices({ page: page, size: 20, state: 'pending_review' }, function(res) {
        that.setData({ list: (res.data || []).map(function(n) { return { id: n.id, title: n.title, sub: n.author + ' · ' + (stateMap[n.workflow_state] || ''), state: n.workflow_state, raw: n } }), total: res.total })
      })
    } else if (type === 'subsidy') {
      api.subsidies({ page: page, size: 20 }, function(res) {
        that.setData({ list: (res.data || []).filter(function(s) { return s.workflow_state === 'submitted' || s.workflow_state === 'secretary_review' }).map(function(s) { return { id: s.id, title: s.title, sub: s.applicant + ' · ¥' + (s.amount/100).toFixed(2) + ' · ' + (subTypeMap[s.type]||s.type), state: s.workflow_state, raw: s } }), total: res.total })
      })
    } else if (type === 'ticket') {
      api.tickets({ page: page, size: 20 }, function(res) {
        that.setData({ list: (res.data || []).filter(function(t) { return t.workflow_state !== 'resolved' && t.workflow_state !== 'closed' }).map(function(t) { return { id: t.id, title: t.title, sub: t.submitter + ' · ' + (ticketCatMap[t.category]||t.category), state: t.workflow_state, raw: t } }), total: res.total })
      })
    } else if (type === 'finance') {
      api.adminFinance({ page: page, size: 20 }, function(res) {
        that.setData({ list: (res.data || []).filter(function(f) { return f.workflow_state === 'pending_review' }).map(function(f) { return { id: f.id, title: (f.type==='income'?'收入':'支出') + ' ¥' + (f.amount/100).toFixed(2), sub: f.category + ' · ' + f.author + ' · ' + f.date, state: f.workflow_state, raw: f } }), total: res.total })
      })
    }
  },
  showDetail: function(e) {
    var item = e.currentTarget.dataset.item
    this.setData({ detail: item, note: '', commentText: '' })
  },
  closeDetail: function() { this.setData({ detail: null }) },
  onNoteInput: function(e) { this.setData({ note: e.detail.value }) },
  onCommentInput: function(e) { this.setData({ commentText: e.detail.value }) },
  approve: function() {
    var that = this; var d = this.data.detail; var type = this.data.type; var note = this.data.note
    if (type === 'notice') {
      api.reviewNotice(d.id, { action: 'approve', note: note }, function() { wx.showToast({ title: '已通过' }); that.setData({ detail: null }); that.loadList() })
    } else if (type === 'subsidy') {
      if (d.state === 'submitted') {
        api.committeeReview(d.id, { action: 'approve', note: note }, function() { wx.showToast({ title: '已转终审' }); that.setData({ detail: null }); that.loadList() })
      } else {
        api.secretaryReview(d.id, { action: 'approve', note: note }, function() { wx.showToast({ title: '已通过' }); that.setData({ detail: null }); that.loadList() })
      }
    } else if (type === 'finance') {
      api.reviewFinance(d.id, { action: 'approve', note: note }, function() { wx.showToast({ title: '已通过' }); that.setData({ detail: null }); that.loadList() })
    }
  },
  reject: function() {
    var that = this; var d = this.data.detail; var type = this.data.type; var note = this.data.note
    if (type === 'notice') {
      api.reviewNotice(d.id, { action: 'reject', note: note }, function() { wx.showToast({ title: '已驳回' }); that.setData({ detail: null }); that.loadList() })
    } else if (type === 'subsidy') {
      if (d.state === 'submitted') {
        api.committeeReview(d.id, { action: 'reject', note: note }, function() { wx.showToast({ title: '已驳回' }); that.setData({ detail: null }); that.loadList() })
      } else {
        api.secretaryReview(d.id, { action: 'reject', note: note }, function() { wx.showToast({ title: '已驳回' }); that.setData({ detail: null }); that.loadList() })
      }
    } else if (type === 'finance') {
      api.reviewFinance(d.id, { action: 'reject', note: note }, function() { wx.showToast({ title: '已驳回' }); that.setData({ detail: null }); that.loadList() })
    }
  },
  assignTicket: function() {
    var that = this; var d = this.data.detail
    api.assignTicket(d.id, function() { wx.showToast({ title: '已认领' }); that.setData({ detail: null }); that.loadList() })
  },
  resolveTicket: function() {
    var that = this; var d = this.data.detail
    api.adminUpdateTicketStatus(d.id, { status: 'resolved' }, function() { wx.showToast({ title: '已解决' }); that.setData({ detail: null }); that.loadList() })
  },
  sendComment: function() {
    var that = this; var d = this.data.detail; var text = this.data.commentText
    if (!text) return
    api.addComment(d.id, { content: text }, function() { wx.showToast({ title: '已回复' }); that.setData({ commentText: '' }) })
  }
})
