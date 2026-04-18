var api = require('./utils/api')

App({
  globalData: {
    token: '',
    userInfo: null,
    villageName: '',
    wxPhoneEnabled: false
  },
  onLaunch: function() {
    var token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
      this.globalData.userInfo = wx.getStorageSync('userInfo')
    }
    var that = this
    api.config(function(res) {
      that.globalData.villageName = (res && res.village_name) || '村务'
      that.globalData.wxPhoneEnabled = !!(res && res.wx_phone_enabled)
    })
  }
})
