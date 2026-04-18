var api = require('../../utils/api')

Page({
  data: { detail: null, logs: [] },
  onLoad: function(opts) {
    var that = this
    api.subsidy(opts.id, function(res) {
      that.setData({ detail: res.subsidy || res, logs: res.logs || [] })
    })
  }
})
