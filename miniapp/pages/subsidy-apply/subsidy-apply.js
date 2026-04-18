var api = require('../../utils/api')

Page({
  data: {
    title: '', typeIndex: 0, amount: '', reason: '',
    types: ['farming', 'medical', 'education', 'housing', 'other'],
    typeNames: ['种植补贴', '医疗补贴', '教育补贴', '住房补贴', '其他']
  },
  onInput: function(e) {
    var obj = {}
    obj[e.currentTarget.dataset.field] = e.detail.value
    this.setData(obj)
  },
  onTypePick: function(e) { this.setData({ typeIndex: e.detail.value }) },
  submit: function() {
    var d = this.data
    if (!d.title || !d.amount) {
      wx.showToast({ title: '请填写完整', icon: 'none' }); return
    }
    api.applySubsidy({
      title: d.title,
      type: d.types[d.typeIndex],
      amount: Math.round(parseFloat(d.amount) * 100),
      reason: d.reason
    }, function() {
      wx.showToast({ title: '申请已提交' })
      setTimeout(function() { wx.navigateBack() }, 1000)
    })
  }
})
