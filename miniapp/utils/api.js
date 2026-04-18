var BASE = 'https://your-domain.com'  // 修改为你的服务器地址

function request(url, options, success, fail) {
  options = options || {}
  var app = getApp() || {}
  var gd = app.globalData || {}
  var header = { 'Content-Type': 'application/json' }
  if (gd.token) {
    header['Authorization'] = 'Bearer ' + gd.token
  }
  wx.request({
    url: BASE + url,
    method: options.method || 'GET',
    data: options.data,
    header: header,
    success: function(res) {
      if (res.statusCode === 401) {
        var a = getApp()
        if (a) { a.globalData.token = '' }
        wx.removeStorageSync('token')
        wx.showToast({ title: '请重新登录', icon: 'none' })
        if (fail) fail(res)
        return
      }
      if (success) success(res.data)
    },
    fail: function(err) {
      console.log('请求失败:', url, err)
      if (fail) fail(err)
    }
  })
}

function qs(obj) {
  if (!obj) return ''
  var arr = []
  for (var k in obj) {
    if (obj[k] !== undefined && obj[k] !== '') {
      arr.push(k + '=' + encodeURIComponent(obj[k]))
    }
  }
  return arr.join('&')
}

module.exports = {
  BASE: BASE,
  // 配置
  config: function(cb) { request('/api/config', {}, cb) },
  // 微信登录
  wxLogin: function(data, cb) { request('/api/wx/login', { method: 'POST', data: data }, cb) },
  // 微信获取手机号
  wxPhone: function(data, cb) { request('/api/wx/phone', { method: 'POST', data: data }, cb) },
  // 上传
  upload: function(filePath, cb) {
    var app = getApp() || {}
    var gd = app.globalData || {}
    wx.uploadFile({
      url: BASE + '/api/upload',
      filePath: filePath,
      name: 'file',
      header: gd.token ? { 'Authorization': 'Bearer ' + gd.token } : {},
      success: function(res) {
        try { cb(JSON.parse(res.data)) } catch(e) { cb({}) }
      },
      fail: function() { cb({}) }
    })
  },
  // 公告
  notices: function(params, cb) { request('/api/notices?' + qs(params), {}, cb) },
  notice: function(id, cb) { request('/api/notices/' + id, {}, cb) },
  // 财务
  finance: function(params, cb) { request('/api/finance?' + qs(params), {}, cb) },
  financeSummary: function(params, cb) { request('/api/finance/summary?' + qs(params), {}, cb) },
  // 补贴
  subsidies: function(params, cb) { request('/api/subsidies?' + qs(params), {}, cb) },
  subsidy: function(id, cb) { request('/api/subsidies/' + id, {}, cb) },
  applySubsidy: function(data, cb) { request('/api/subsidies', { method: 'POST', data: data }, cb) },
  // 工单
  tickets: function(params, cb) { request('/api/tickets?' + qs(params), {}, cb) },
  ticket: function(id, cb) { request('/api/tickets/' + id, {}, cb) },
  createTicket: function(data, cb) { request('/api/tickets', { method: 'POST', data: data }, cb) },
  addComment: function(id, data, cb) { request('/api/tickets/' + id + '/comments', { method: 'POST', data: data }, cb) },
  updateTicketStatus: function(id, data, cb) { request('/api/tickets/' + id + '/status', { method: 'PUT', data: data }, cb) },
  // 用户
  me: function(cb) { request('/api/me', {}, cb) },
  updateProfile: function(data, cb) { request('/api/me', { method: 'PUT', data: data }, cb) },
  bindPhone: function(data, cb) { request('/api/me/bindphone', { method: 'POST', data: data }, cb) },
  changePassword: function(data, cb) { request('/api/me/password', { method: 'POST', data: data }, cb) },
  // 通知
  notifications: function(params, cb) { request('/api/notifications?' + qs(params), {}, cb) },
  unreadCount: function(cb) { request('/api/notifications/unread-count', {}, cb) },
  markRead: function(id, cb) { request('/api/notifications/' + id + '/read', { method: 'PUT' }, cb) },
  markAllRead: function(cb) { request('/api/notifications/read-all', { method: 'POST' }, cb) },
  // 管理接口
  dashboard: function(cb) { request('/api/admin/dashboard', {}, cb) },
  adminNotices: function(params, cb) { request('/api/admin/notices?' + qs(params), {}, cb) },
  reviewNotice: function(id, data, cb) { request('/api/admin/notices/' + id + '/review', { method: 'PUT', data: data }, cb) },
  adminFinance: function(params, cb) { request('/api/admin/finance?' + qs(params), {}, cb) },
  reviewFinance: function(id, data, cb) { request('/api/admin/finance/' + id + '/review', { method: 'PUT', data: data }, cb) },
  committeeReview: function(id, data, cb) { request('/api/admin/subsidies/' + id + '/committee-review', { method: 'PUT', data: data }, cb) },
  secretaryReview: function(id, data, cb) { request('/api/admin/subsidies/' + id + '/secretary-review', { method: 'PUT', data: data }, cb) },
  assignTicket: function(id, cb) { request('/api/admin/tickets/' + id + '/assign', { method: 'PUT', data: {} }, cb) },
  adminUpdateTicketStatus: function(id, data, cb) { request('/api/admin/tickets/' + id + '/status', { method: 'PUT', data: data }, cb) },
  adminUsers: function(params, cb) { request('/api/admin/users?' + qs(params), {}, cb) },
  adminUpdateUser: function(id, data, cb) { request('/api/admin/users/' + id, { method: 'PUT', data: data }, cb) },
  adminResetPassword: function(id, cb) { request('/api/admin/users/' + id + '/reset-password', { method: 'POST' }, cb) },
  adminHouseholds: function(params, cb) { request('/api/admin/households?' + qs(params), {}, cb) },
  adminHousehold: function(id, cb) { request('/api/admin/households/' + id, {}, cb) },
  adminCreateHousehold: function(data, cb) { request('/api/admin/households', { method: 'POST', data: data }, cb) },
  adminUpdateHousehold: function(id, data, cb) { request('/api/admin/households/' + id, { method: 'PUT', data: data }, cb) },
  adminAddMember: function(hid, data, cb) { request('/api/admin/households/' + hid + '/members', { method: 'POST', data: data }, cb) },
  adminRemoveMember: function(hid, mid, cb) { request('/api/admin/households/' + hid + '/members/' + mid, { method: 'DELETE' }, cb) },
  groups: function(cb) { request('/api/groups', {}, cb) }
}
