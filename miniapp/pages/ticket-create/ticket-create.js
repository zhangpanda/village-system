var api = require('../../utils/api')

Page({
  data: {
    title: '', content: '', catIndex: 2, priIndex: 1, images: [],
    categories: ['repair', 'complaint', 'service', 'suggestion'],
    catNames: ['报修', '投诉', '便民服务', '建议'],
    priorities: ['low', 'normal', 'urgent'],
    priNames: ['低', '普通', '紧急']
  },
  onInput: function(e) {
    var obj = {}
    obj[e.currentTarget.dataset.field] = e.detail.value
    this.setData(obj)
  },
  onCatPick: function(e) { this.setData({ catIndex: e.detail.value }) },
  onPriPick: function(e) { this.setData({ priIndex: e.detail.value }) },
  chooseImage: function() {
    var that = this
    wx.chooseMedia({
      count: 3 - that.data.images.length,
      mediaType: ['image'],
      success: function(res) {
        var paths = res.tempFiles.map(function(f) { return f.tempFilePath })
        that.setData({ images: that.data.images.concat(paths) })
      }
    })
  },
  removeImage: function(e) {
    var idx = e.currentTarget.dataset.idx
    var imgs = this.data.images
    imgs.splice(idx, 1)
    this.setData({ images: imgs })
  },
  submit: function() {
    var d = this.data
    if (!d.title || !d.content) {
      wx.showToast({ title: '请填写完整', icon: 'none' }); return
    }
    wx.showLoading({ title: '提交中...' })
    var that = this
    var uploadedUrls = []
    var uploadNext = function(i) {
      if (i >= d.images.length) {
        api.createTicket({
          title: d.title,
          content: d.content,
          category: d.categories[d.catIndex],
          priority: d.priorities[d.priIndex],
          images: JSON.stringify(uploadedUrls)
        }, function() {
          wx.hideLoading()
          wx.showToast({ title: '提交成功' })
          setTimeout(function() { wx.navigateBack() }, 1000)
        })
        return
      }
      api.upload(d.images[i], function(res) {
        if (res && res.url) uploadedUrls.push(res.url)
        uploadNext(i + 1)
      })
    }
    uploadNext(0)
  }
})
