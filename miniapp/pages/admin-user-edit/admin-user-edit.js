var api = require('../../utils/api')

var roles = ['admin','secretary','resident_official','director','deputy','supervisor','committee','accountant','group_leader','grid_worker','villager']
var roleLabels = ['系统管理员','党支部书记','驻村干部','村委会主任','副书记/副主任','监委会委员','两委委员','村会计','村民小组长','网格员','村民']
var genders = ['male','female']
var genderLabels = ['男','女']
var eduList = ['文盲','小学','初中','高中','大专','本科','硕士及以上']
var maritalList = ['unmarried','married','divorced','widowed']
var maritalLabels = ['未婚','已婚','离异','丧偶']

Page({
  data: {
    id: 0, form: {}, groups: [], groupNames: [],
    roles: roles, roleLabels: roleLabels, roleIdx: 10,
    genders: genders, genderLabels: genderLabels, genderIdx: -1,
    eduList: eduList, eduIdx: -1,
    maritalList: maritalList, maritalLabels: maritalLabels, maritalIdx: -1,
    groupIdx: -1
  },
  onLoad: function(opts) {
    this.setData({ id: parseInt(opts.id) || 0 })
    this.loadGroups()
    this.loadUser()
  },
  loadGroups: function() {
    var that = this
    api.groups(function(res) {
      var list = res || []
      that.setData({ groups: list, groupNames: list.map(function(g) { return g.name }) })
    })
  },
  loadUser: function() {
    var that = this; var id = this.data.id
    api.adminUsers({ q: '', page: 1, size: 1000 }, function(res) {
      var u = (res.data || []).filter(function(x) { return x.id === id })[0]
      if (!u) return
      that.setData({
        form: u,
        roleIdx: roles.indexOf(u.role) >= 0 ? roles.indexOf(u.role) : 10,
        genderIdx: genders.indexOf(u.gender),
        eduIdx: eduList.indexOf(u.education),
        maritalIdx: maritalList.indexOf(u.marital_status),
        groupIdx: that.data.groups.map(function(g){return g.id}).indexOf(u.group_id)
      })
    })
  },
  onInput: function(e) {
    var key = 'form.' + e.currentTarget.dataset.key
    this.setData({ [key]: e.detail.value })
  },
  onRolePick: function(e) { this.setData({ roleIdx: e.detail.value, 'form.role': roles[e.detail.value] }) },
  onGenderPick: function(e) { this.setData({ genderIdx: e.detail.value, 'form.gender': genders[e.detail.value] }) },
  onEduPick: function(e) { this.setData({ eduIdx: e.detail.value, 'form.education': eduList[e.detail.value] }) },
  onMaritalPick: function(e) { this.setData({ maritalIdx: e.detail.value, 'form.marital_status': maritalList[e.detail.value] }) },
  onGroupPick: function(e) {
    var g = this.data.groups[e.detail.value]
    this.setData({ groupIdx: e.detail.value, 'form.group_id': g ? g.id : 0 })
  },
  toggleFlag: function(e) {
    var key = 'form.' + e.currentTarget.dataset.key
    this.setData({ [key]: !this.data.form[e.currentTarget.dataset.key] })
  },
  save: function() {
    var f = this.data.form
    if (!f.name) { wx.showToast({ title: '请填写姓名', icon: 'none' }); return }
    api.adminUpdateUser(this.data.id, f, function() {
      wx.showToast({ title: '保存成功' })
      setTimeout(function() { wx.navigateBack() }, 800)
    })
  }
})
