var api = require('../../utils/api')
var roleMap = {admin:'系统管理员',secretary:'党支部书记',director:'村委会主任',deputy:'副书记/副主任',committee:'两委委员',supervisor:'监委会委员',accountant:'村会计',group_leader:'村民小组长',grid_worker:'网格员',villager:'村民'}

Page({
  data: { list: [], total: 0, page: 1, q: '' },
  onShow: function() { this.loadData() },
  onSearch: function(e) { this.setData({ q: e.detail.value, page: 1 }); this.loadData() },
  loadData: function() {
    var that = this
    api.adminUsers({ page: this.data.page, size: 20, q: this.data.q }, function(res) {
      var list = (res.data || []).map(function(u) {
        u.roleLabel = roleMap[u.role] || u.role
        return u
      })
      that.setData({ list: list, total: res.total })
    })
  },
  editUser: function(e) {
    var id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/admin-user-edit/admin-user-edit?id=' + id })
  },
  resetPwd: function(e) {
    var id = e.currentTarget.dataset.id
    var name = e.currentTarget.dataset.name
    wx.showModal({
      title: '重置密码',
      content: '确定将 ' + name + ' 的密码重置为 123456？',
      success: function(res) {
        if (res.confirm) {
          api.adminResetPassword(id, function() { wx.showToast({ title: '已重置' }) })
        }
      }
    })
  }
})
