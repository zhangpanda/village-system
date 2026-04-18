var api = require('../../utils/api')

Page({
  data: {
    userInfo: null, logged: false, editMode: false,
    editName: '', editPhone: '', editAddress: '', editGender: '', editGenderIdx: -1,
    showPwdEdit: false, oldPwd: '', newPwd: '', newPwd2: '',
    unreadCount: 0, wxPhoneEnabled: false
  },
  onShow: function() {
    var app = getApp()
    this.setData({ wxPhoneEnabled: app.globalData.wxPhoneEnabled })
    if (app.globalData.token) {
      this.setData({ logged: true, userInfo: app.globalData.userInfo, editMode: false })
      this.loadProfile()
      this.loadUnread()
    }
  },
  loadUnread: function() {
    var that = this
    api.unreadCount(function(res) {
      that.setData({ unreadCount: (res && res.count) || 0 })
    })
  },
  loadProfile: function() {
    var that = this
    api.me(function(u) {
      getApp().globalData.userInfo = u
      wx.setStorageSync('userInfo', u)
      var adminRoles = ['admin','secretary','resident_official','director','deputy','supervisor','committee','accountant','group_leader','grid_worker']
      var isAdmin = adminRoles.some(function(r) { return (u.role || '').indexOf(r) >= 0 })
      that.setData({ userInfo: u, isAdmin: isAdmin })
    })
  },
  wxLogin: function() {
    var that = this
    wx.login({
      success: function(res) {
        if (!res.code) return
        api.wxLogin({ code: res.code }, function(data) {
          if (!data || !data.token) {
            wx.showToast({ title: '登录失败', icon: 'none' }); return
          }
          var app = getApp()
          app.globalData.token = data.token
          app.globalData.userInfo = data.user
          wx.setStorageSync('token', data.token)
          wx.setStorageSync('userInfo', data.user)
          that.setData({ logged: true, userInfo: data.user })
          // 新用户引导填写昵称
          if (!data.user.name || data.user.name.indexOf('微信用户') === 0) {
            that.showNicknameSetting()
          } else {
            wx.showToast({ title: '登录成功' })
          }
        })
      }
    })
  },
  startEdit: function() {
    var u = this.data.userInfo
    var phone = u.phone || ''
    if (phone.substring(0, 3) === 'wx_') phone = ''
    var genderIdx = u.gender === 'male' ? 0 : u.gender === 'female' ? 1 : -1
    this.setData({
      editMode: true,
      editName: u.name || '',
      editPhone: phone,
      editAddress: u.address || '',
      editGender: u.gender || '',
      editGenderIdx: genderIdx
    })
  },
  cancelEdit: function() { this.setData({ editMode: false }) },
  onEditName: function(e) { this.setData({ editName: e.detail.value }) },
  onEditPhone: function(e) { this.setData({ editPhone: e.detail.value }) },
  onGetPhoneNumber: function(e) {
    if (!e.detail.code) { wx.showToast({ title: '取消授权', icon: 'none' }); return }
    var that = this
    api.wxPhone({ code: e.detail.code }, function(res) {
      if (res && res.phone) {
        that.setData({ editPhone: res.phone })
        wx.showToast({ title: '获取成功' })
        that.loadProfile()
      } else {
        wx.showToast({ title: (res && res.error) || '获取失败', icon: 'none' })
      }
    })
  },
  onEditAddress: function(e) { this.setData({ editAddress: e.detail.value }) },
  onEditGender: function(e) {
    var idx = e.detail.value
    this.setData({ editGender: idx == 0 ? 'male' : 'female', editGenderIdx: idx })
  },
  saveProfile: function() {
    var that = this
    var name = this.data.editName.trim()
    var phone = this.data.editPhone.trim()
    var address = this.data.editAddress.trim()
    if (!name) {
      wx.showToast({ title: '姓名不能为空', icon: 'none' }); return
    }
    if (phone && phone.length !== 11) {
      wx.showToast({ title: '手机号需11位', icon: 'none' }); return
    }
    // 先保存姓名和地址
    api.updateProfile({ name: name, address: address, gender: this.data.editGender }, function(res) {
      if (res && res.error) {
        wx.showToast({ title: res.error, icon: 'none' }); return
      }
      // 如果填了手机号且有变化，再绑定手机号
      var oldPhone = that.data.userInfo.phone || ''
      if (phone && phone !== oldPhone) {
        api.bindPhone({ phone: phone }, function(res2) {
          if (res2 && res2.error) {
            wx.showToast({ title: res2.error, icon: 'none' }); return
          }
          wx.showToast({ title: '保存成功' })
          that.setData({ editMode: false })
          that.loadProfile()
        })
      } else {
        wx.showToast({ title: '保存成功' })
        that.setData({ editMode: false })
        that.loadProfile()
      }
    })
  },
  logout: function() {
    var app = getApp()
    app.globalData.token = ''
    app.globalData.userInfo = null
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    this.setData({ logged: false, userInfo: null })
  },
  goSubsidy: function() { wx.navigateTo({ url: '/pages/subsidy/subsidy' }) },
  goTickets: function() { wx.navigateTo({ url: '/pages/tickets/tickets' }) },
  goAdmin: function() { wx.navigateTo({ url: '/pages/admin-dashboard/admin-dashboard' }) },
  goNotifications: function() {
    wx.navigateTo({
      url: '/pages/notifications/notifications',
      fail: function() {
        wx.showToast({ title: '通知页面开发中', icon: 'none' })
      }
    })
  },
  showPasswordEdit: function() { this.setData({ showPwdEdit: true, oldPwd: '', newPwd: '', newPwd2: '' }) },
  hidePwdEdit: function() { this.setData({ showPwdEdit: false }) },
  onOldPwd: function(e) { this.setData({ oldPwd: e.detail.value }) },
  onNewPwd: function(e) { this.setData({ newPwd: e.detail.value }) },
  onNewPwd2: function(e) { this.setData({ newPwd2: e.detail.value }) },
  doChangePwd: function() {
    var d = this.data
    if (d.newPwd.length < 6) { wx.showToast({ title: '密码至少6位', icon: 'none' }); return }
    if (d.newPwd !== d.newPwd2) { wx.showToast({ title: '两次密码不一致', icon: 'none' }); return }
    var that = this
    api.changePassword({ old_password: d.oldPwd, new_password: d.newPwd }, function(res) {
      if (res && res.error) { wx.showToast({ title: res.error, icon: 'none' }); return }
      wx.showToast({ title: '密码修改成功' })
      that.setData({ showPwdEdit: false })
    })
  },
  showNicknameSetting: function() {
    this.setData({ showNicknameInput: true, nicknameValue: '' })
  },
  onNicknameInput: function(e) { this.setData({ nicknameValue: e.detail.value }) },
  saveNickname: function() {
    var name = (this.data.nicknameValue || '').trim()
    if (!name) { wx.showToast({ title: '请输入姓名', icon: 'none' }); return }
    var that = this
    api.updateProfile({ name: name }, function(res) {
      if (res && res.error) { wx.showToast({ title: res.error, icon: 'none' }); return }
      that.setData({ showNicknameInput: false })
      wx.showToast({ title: '欢迎你，' + name })
      that.loadProfile()
    })
  },
  onShareAppMessage: function() {
    return { title: (getApp().globalData.villageName || '村务') + ' · 村务公开', path: '/pages/index/index' }
  }
})
