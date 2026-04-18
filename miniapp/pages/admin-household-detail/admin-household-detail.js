var api = require('../../utils/api')
var relations = ['户主','配偶','之子','之女','之父','之母','儿媳','女婿','之孙','之孙女','祖父','祖母','外祖父','外祖母','兄弟','姐妹','其他']

Page({
  data: { id: 0, household: null, members: [], relations: relations },
  onLoad: function(opts) { this.setData({ id: parseInt(opts.id) || 0 }) },
  onShow: function() { this.loadData() },
  loadData: function() {
    var that = this
    api.adminHousehold(this.data.id, function(res) {
      that.setData({ household: res.household || res, members: res.members || [] })
    })
  },
  removeMember: function(e) {
    var that = this; var mid = e.currentTarget.dataset.mid; var name = e.currentTarget.dataset.name
    wx.showModal({
      title: '移除成员', content: '确定移除 ' + name + '？',
      success: function(res) {
        if (res.confirm) {
          api.adminRemoveMember(that.data.id, mid, function() {
            wx.showToast({ title: '已移除' }); that.loadData()
          })
        }
      }
    })
  },
  editUser: function(e) {
    wx.navigateTo({ url: '/pages/admin-user-edit/admin-user-edit?id=' + e.currentTarget.dataset.uid })
  }
})
