const API = location.origin + '/api';
let token = localStorage.getItem('vtoken') || '';
let user = JSON.parse(localStorage.getItem('vuser') || 'null');
let tab = 'notice';
let villageName = '村务';

const catMap = {policy:'政策通知',activity:'村务活动',urgent:'紧急通知',meeting:'会议'};
const subTypeMap = {farming:'农业补贴',medical:'医疗救助',education:'教育补助',housing:'住房补贴',other:'其他'};
const ticketCatMap = {repair:'报修',complaint:'投诉',service:'便民服务',suggestion:'建议'};
const statusMap = {draft:'草稿',pending_review:'待审核',published:'已发布',submitted:'已提交',committee_review:'村委初审中',secretary_review:'村支书终审中',approved:'已通过',rejected:'已驳回',paid:'已发放',open:'待处理',assigned:'已分配',processing:'处理中',resolved:'已解决',closed:'已关闭'};

const fmt = v => (v/100).toLocaleString('zh-CN',{minimumFractionDigits:2});
const fmtDate = s => {
  if(!s) return '';
  const d = new Date(s);
  const now = new Date();
  const diff = (now - d) / 1000;
  if(diff < 60) return '刚刚';
  if(diff < 3600) return Math.floor(diff/60) + '分钟前';
  if(diff < 86400) return Math.floor(diff/3600) + '小时前';
  if(diff < 604800) return Math.floor(diff/86400) + '天前';
  return d.toLocaleDateString('zh-CN');
};
const stripHtml = s => s ? s.replace(/<[^>]*>/g,'').replace(/&nbsp;/g,' ') : '';
const esc = s => { const d=document.createElement('div'); d.textContent=s||''; return d.innerHTML; };
const $ = id => document.getElementById(id);function headers() { return {'Content-Type':'application/json','Authorization':'Bearer '+token}; }
function toast(msg) { const t=$('toast'); t.textContent=msg; t.style.display='block'; setTimeout(()=>t.style.display='none',2000); }

// 加载站点配置
fetch(API+'/config').then(r=>r.json()).then(c=>{
  villageName = c.village_name || '村务';
  document.title = villageName + ' · 村务公开';
  $('siteTitle').textContent = '🏘️ ' + villageName + ' · 村务公开平台';
}).catch(()=>{});
const isPhone = s => /^1\d{10}$/.test(s);
function closeDetail() { $('detail').classList.remove('show'); document.body.style.overflow=''; }
function showDetail(html) { $('detailContent').innerHTML=html; $('detail').classList.add('show'); document.body.style.overflow='hidden'; }

function saveAuth(t, u) { token=t; user=u; localStorage.setItem('vtoken',t); localStorage.setItem('vuser',JSON.stringify(u)); updateUserBar(); }
function clearAuth() { token=''; user=null; localStorage.removeItem('vtoken'); localStorage.removeItem('vuser'); updateUserBar(); }

function updateUserBar() {
  const bar = $('userBar');
  if (user) {
    bar.innerHTML = `欢迎，${esc(user.name)} <a onclick="clearAuth();load()">退出</a>`;
  } else {
    bar.innerHTML = `<a onclick="switchTab('me')">登录</a> / <a onclick="switchTab('me')">注册</a>`;
  }
}

function switchTab(t) {
  tab = t;
  document.querySelectorAll('.tab').forEach(el => el.classList.toggle('active', el.dataset.tab === t));
  load();
}

document.querySelectorAll('.tab').forEach(el => { el.onclick = () => switchTab(el.dataset.tab); });
updateUserBar();

async function load() {
  const app = $('app');
  app.innerHTML = '<div class="empty">加载中...</div>';
  try {
    switch(tab) {
      case 'notice': await loadNotices(); break;
      case 'finance': await loadFinance(); break;
      case 'subsidy': await loadSubsidies(); break;
      case 'ticket': await loadTickets(); break;
      case 'me': loadMe(); break;
    }
  } catch(e) { app.innerHTML = '<div class="empty">加载失败</div>'; }
}

// ==================== 公告 ====================
let noticePage = 1;
let noticeKeyword = '';
async function loadNotices() {
  const res = await (await fetch(API+'/notices?size=10&page='+noticePage+'&q='+encodeURIComponent(noticeKeyword))).json();
  const data = res.data, total = res.total;
  const app = $('app');
  const totalPages = Math.ceil(total / 10);
  let html = `<div class="search-bar"><input id="noticeSearch" placeholder="搜索公告..." value="${noticeKeyword}" onkeydown="if(event.key==='Enter'){noticeKeyword=this.value;noticePage=1;loadNotices()}"><button onclick="noticeKeyword=$('noticeSearch').value;noticePage=1;loadNotices()">搜索</button></div>`;
  if(!data?.length) { html += '<div class="empty">暂无公告</div>'; }
  else {
    html += data.map(n => `
      <div class="card" onclick="showNotice(${n.id})" style="cursor:pointer">
        <div class="title">
          ${n.pinned?'<span class="tag pinned">置顶</span>':''}
          <span class="tag ${n.category}">${catMap[n.category]||n.category}</span>
          ${esc(n.title)}
        </div>
        <div class="body" style="max-height:60px;overflow:hidden;color:#888;font-size:12px">${stripHtml(n.content).substring(0,80)}...</div>
        <div class="meta"><span>${esc(n.author)}</span><span>👁 ${n.views} · ${fmtDate(n.created_at)}</span></div>
      </div>
    `).join('');
    if(totalPages > 1) {
      html += `<div class="pagination">`;
      if(noticePage > 1) html += `<a onclick="noticePage--;loadNotices()">‹ 上一页</a>`;
      html += `<span>${noticePage} / ${totalPages}</span>`;
      if(noticePage < totalPages) html += `<a onclick="noticePage++;loadNotices()">下一页 ›</a>`;
      html += `</div>`;
    }
  }
  app.innerHTML = html;
}

async function showNotice(id) {
  const res = await (await fetch(API+'/notices/'+id)).json();
  const n = res.notice || res;
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <div style="margin-bottom:8px">
      ${n.pinned?'<span class="tag pinned">置顶</span>':''}
      <span class="tag ${n.category}">${catMap[n.category]||n.category}</span>
    </div>
    <h2>${esc(n.title)}</h2>
    <div style="font-size:12px;color:#999;margin-bottom:16px">${esc(n.author)} · ${fmtDate(n.created_at)} · 阅读 ${n.views}</div>
    <div style="font-size:14px;line-height:1.8">${n.content.includes('<')?n.content:n.content.replace(/\n/g,'<br>')}</div>
  `);
}

// ==================== 财务 ====================
async function loadFinance() {
  const year = new Date().getFullYear();
  const [sumRes, listRes] = await Promise.all([
    fetch(API+'/finance/summary?year='+year),
    fetch(API+'/finance?year='+year+'&size=50')
  ]);
  const sum = await sumRes.json();
  const {data} = await listRes.json();
  const app = $('app');
  let catHtml = '';
  if(sum.by_category?.length) {
    const maxAmt = Math.max(...sum.by_category.map(c=>c.amount));
    catHtml = '<div class="card"><div class="title" style="font-size:14px;margin-bottom:10px">分类明细</div>' +
      sum.by_category.map(c => `
        <div class="cat-bar">
          <span style="width:70px;color:${c.type==='income'?'var(--green)':'var(--red)'}">${c.category}</span>
          <div class="bar"><div class="bar-fill" style="width:${c.amount/maxAmt*100}%;background:${c.type==='income'?'#66bb6a':'#ef5350'}"></div></div>
          <span style="width:80px;text-align:right">¥${fmt(c.amount)}</span>
        </div>
      `).join('') + '</div>';
  }
  app.innerHTML = `
    <div class="summary-grid">
      <div class="summary-item income"><div class="num">¥${fmt(sum.total_income)}</div><div class="label">${year}年收入</div></div>
      <div class="summary-item expense"><div class="num">¥${fmt(sum.total_expense)}</div><div class="label">${year}年支出</div></div>
      <div class="summary-item"><div class="num" style="color:${sum.balance>=0?'var(--green)':'var(--red)'}">¥${fmt(sum.balance)}</div><div class="label">结余</div></div>
    </div>
    ${catHtml}
    ${(!data?.length)?'<div class="empty">暂无记录</div>':data.map(r=>`
      <div class="card" onclick='showFinanceDetail(${JSON.stringify(r).replace(/'/g,"&#39;")})' style="cursor:pointer">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <div class="title" style="color:${r.type==='income'?'var(--green)':'var(--red)'}">
            ${r.type==='income'?'+':'-'}¥${fmt(r.amount)}
          </div>
          <span style="font-size:12px;color:#999">${r.date}</span>
        </div>
        <div style="font-size:12px;color:var(--gray);margin-top:4px">${esc(r.category)} · ${esc(r.remark)}</div>
        <div class="meta"><span>${esc(r.author)}</span></div>
      </div>
    `).join('')}
  `;
}

function showFinanceDetail(r) {
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <h2 style="color:${r.type==='income'?'var(--green)':'var(--red)'}">${r.type==='income'?'收入':'支出'}：¥${fmt(r.amount)}</h2>
    <div style="font-size:14px;color:#666;margin:16px 0;line-height:2.2">
      <div>📂 分类：${r.category||'未分类'}</div>
      <div>📅 日期：${r.date}</div>
      <div>📝 备注：${esc(r.remark)||'无'}</div>
      <div>👤 录入人：${esc(r.author)}</div>
      ${r.voucher?'<div>🧾 凭证：<img src="'+r.voucher+'" style="max-width:100%;border-radius:8px;margin-top:8px;cursor:pointer" onclick="window.open(this.src)"></div>':''}
    </div>
  `);
}

// ==================== 补贴 ====================
async function loadSubsidies() {
  if(!token) { $('app').innerHTML = needLoginHtml('查看和申请补贴'); return; }
  const {data} = await (await fetch(API+'/subsidies?size=50',{headers:headers()})).json();
  const app = $('app');
  let html = '<button class="btn btn-green" style="margin-bottom:12px" onclick="showSubsidyForm()">+ 申请补贴</button>';
  if(!data?.length) { html += '<div class="empty">暂无补贴申请记录</div>'; }
  else { html += data.map(s => `
    <div class="card" onclick="showSubsidyDetail(${s.id})" style="cursor:pointer">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${esc(s.title)}</div>
        <span class="status-tag ${s.workflow_state}">${statusMap[s.workflow_state]}</span>
      </div>
      <div style="font-size:13px;margin-top:6px">
        <span style="color:var(--gray)">金额：</span><b>¥${fmt(s.amount)}</b>
        <span style="color:var(--gray);margin-left:12px">类型：</span>${subTypeMap[s.type]||s.type}
      </div>
      <div style="font-size:12px;color:#888;margin-top:4px;overflow:hidden;white-space:nowrap;text-overflow:ellipsis">${esc(s.reason)||''}</div>
      <div class="meta"><span>${fmtDate(s.created_at)}</span></div>
    </div>
  `).join(''); }
  app.innerHTML = html;
}

async function showSubsidyDetail(id) {
  const {subsidy:s, logs} = await (await fetch(API+'/subsidies/'+id,{headers:headers()})).json();
  const imgs = (()=>{try{return JSON.parse(s.attachments||'[]')}catch(e){return []}})();
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <h2>${esc(s.title)} <span class="status-tag ${s.workflow_state}">${statusMap[s.workflow_state]}</span></h2>
    <div style="font-size:14px;color:#666;margin:16px 0;line-height:2.2">
      <div>📋 类型：${subTypeMap[s.type]||s.type}</div>
      <div>💰 金额：<b style="color:var(--green)">¥${fmt(s.amount)}</b></div>
      <div>📅 申请时间：${fmtDate(s.created_at)}</div>
    </div>
    <div style="font-size:14px;line-height:1.8;margin-bottom:16px"><b>申请理由：</b>${esc(s.reason)||'无'}</div>
    ${imgs.length?'<div style="margin-bottom:12px">'+imgs.map(u=>'<img src="'+u+'" style="width:80px;height:80px;object-fit:cover;border-radius:8px;margin:4px;cursor:pointer" onclick="window.open(this.src)">').join('')+'</div>':''}
    ${s.committee_name?'<div style="font-size:13px;padding:10px;background:#f5f5f5;border-radius:8px;margin-bottom:8px"><b>村委初审：</b>'+esc(s.committee_name)+(s.committee_note?' — '+s.committee_note:'')+'</div>':''}
    ${s.secretary_name?'<div style="font-size:13px;padding:10px;background:#f0fff0;border-radius:8px;margin-bottom:8px"><b>村支书终审：</b>'+esc(s.secretary_name)+(s.secretary_note?' — '+s.secretary_note:'')+'</div>':''}
    ${logs?.length?'<div style="font-size:12px;color:#999;margin-top:12px"><b>审批日志：</b>'+logs.map(l=>'<div>'+esc(l.operator_name)+' '+esc(l.action)+(l.note?' ('+esc(l.note)+')':'')+'</div>').join('')+'</div>':''}
  `);
}

function showSubsidyForm() {
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <h2>申请补贴</h2>
    <div class="form-group"><label>补贴名称</label><input id="subTitle" placeholder="如：2026年春耕补贴"></div>
    <div class="form-group"><label>补贴类型</label>
      <select id="subType"><option value="farming">农业补贴</option><option value="medical">医疗救助</option><option value="education">教育补助</option><option value="housing">住房补贴</option><option value="other">其他</option></select>
    </div>
    <div class="form-group"><label>申请金额（元）</label><input id="subAmount" type="number" step="0.01"></div>
    <div class="form-group"><label>申请理由</label><textarea id="subReason" placeholder="请详细说明申请理由..."></textarea></div>
    <button class="btn btn-green btn-block" onclick="submitSubsidy()">提交申请</button>
  `);
}

async function submitSubsidy() {
  const body = {
    title: $('subTitle').value,
    type: $('subType').value,
    amount: Math.round(parseFloat($('subAmount').value)*100),
    reason: $('subReason').value,
  };
  if(!body.title||!body.amount||!body.reason) { toast('请填写完整'); return; }
  const res = await fetch(API+'/subsidies',{method:'POST',headers:headers(),body:JSON.stringify(body)});
  if(res.ok) { closeDetail(); toast('申请已提交'); loadSubsidies(); }
  else { toast('提交失败'); }
}

// ==================== 工单 ====================
async function loadTickets() {
  if(!token) { $('app').innerHTML = needLoginHtml('提交和查看工单'); return; }
  const {data} = await (await fetch(API+'/tickets?mine=1&size=50',{headers:headers()})).json();
  const app = $('app');
  let html = '<button class="btn btn-green" style="margin-bottom:12px" onclick="showTicketForm()">+ 提交工单</button>';
  if(!data?.length) { html += '<div class="empty">暂无工单记录</div>'; }
  else { html += data.map(t => `
    <div class="card" onclick="showTicketDetail(${t.id})" style="cursor:pointer">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title"><span class="priority-dot ${t.priority}"></span>${esc(t.title)}</div>
        <span class="status-tag ${t.workflow_state}">${statusMap[t.workflow_state]}</span>
      </div>
      <div style="font-size:12px;color:#888;margin-top:4px">${ticketCatMap[t.category]||t.category} · ${esc(t.content.substring(0,50))}...</div>
      <div class="meta"><span>${fmtDate(t.created_at)}</span>${t.assignee?'<span>处理人：'+esc(t.assignee)+'</span>':''}</div>
    </div>
  `).join(''); }
  app.innerHTML = html;
}

function showTicketForm() {
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <h2>提交工单</h2>
    <div class="form-group"><label>标题</label><input id="tktTitle" placeholder="简要描述问题"></div>
    <div class="form-group"><label>类型</label>
      <select id="tktCat"><option value="repair">报修</option><option value="complaint">投诉</option><option value="service">便民服务</option><option value="suggestion">建议</option></select>
    </div>
    <div class="form-group"><label>紧急程度</label>
      <select id="tktPri"><option value="normal">普通</option><option value="urgent">紧急</option><option value="low">不急</option></select>
    </div>
    <div class="form-group"><label>详细描述</label><textarea id="tktContent" placeholder="请详细描述问题..." style="min-height:120px"></textarea></div>
    <div class="form-group"><label>图片（可选）</label><input type="file" id="tktImages" accept="image/*" multiple><div id="tktPreview" class="img-preview"></div></div>
    <button class="btn btn-green btn-block" onclick="submitTicket()">提交</button>
  `);
  $('tktImages').onchange = function() { previewFiles('tktImages','tktPreview'); };
}

function previewFiles(inputId, previewId) {
  const files = $(inputId).files;
  const preview = $(previewId);
  preview.innerHTML = '';
  for(let i=0;i<files.length;i++) {
    const img = document.createElement('img');
    img.src = URL.createObjectURL(files[i]);
    img.style.cssText = 'width:60px;height:60px;object-fit:cover;border-radius:6px;margin:4px';
    preview.appendChild(img);
  }
}

async function uploadFiles(inputId) {
  const files = $(inputId)?.files;
  if(!files || !files.length) return '[]';
  const urls = [];
  for(let i=0;i<files.length;i++) {
    const fd = new FormData();
    fd.append('file', files[i]);
    const res = await fetch(API+'/upload',{method:'POST',headers:{'Authorization':'Bearer '+token},body:fd});
    const data = await res.json();
    if(data.url) urls.push(data.url);
  }
  return JSON.stringify(urls);
}

async function submitTicket() {
  const images = await uploadFiles('tktImages');
  const body = {
    title: $('tktTitle').value,
    category: $('tktCat').value,
    priority: $('tktPri').value,
    content: $('tktContent').value,
    images: images,
  };
  if(!body.title||!body.content) { toast('请填写标题和描述'); return; }
  const res = await fetch(API+'/tickets',{method:'POST',headers:headers(),body:JSON.stringify(body)});
  if(res.ok) { closeDetail(); toast('工单已提交'); loadTickets(); }
  else { toast('提交失败'); }
}

async function showTicketDetail(id) {
  const {ticket:t, comments} = await (await fetch(API+'/tickets/'+id,{headers:headers()})).json();
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <div style="margin-bottom:8px">
      <span class="priority-dot ${t.priority}"></span>
      <span class="status-tag ${t.workflow_state}">${statusMap[t.workflow_state]}</span>
      <span style="font-size:12px;color:#999;margin-left:8px">${ticketCatMap[t.category]||t.category}</span>
    </div>
    <h2>${esc(t.title)}</h2>
    <div style="font-size:12px;color:#999;margin-bottom:12px">${fmtDate(t.created_at)}</div>
    <div style="font-size:14px;line-height:1.7;white-space:pre-line;margin-bottom:16px">${esc(t.content)}</div>
    ${(()=>{try{const imgs=JSON.parse(t.images||'[]');return imgs.length?'<div class="img-preview">'+imgs.map(u=>'<img src="'+u+'" style="width:80px;height:80px;object-fit:cover;border-radius:6px;margin:4px;cursor:pointer" onclick="window.open(this.src)">').join('')+'</div>':'';}catch(e){return '';}})()}
    ${t.assignee?'<div style="font-size:12px;color:var(--blue);margin-bottom:12px">处理人：'+esc(t.assignee)+'</div>':''}
    ${comments?.length?'<div style="font-size:14px;font-weight:600;margin-bottom:8px">处理记录</div>'+
      comments.map(c=>'<div class="comment-item"><div class="name">'+esc(c.user_name)+'</div><div class="text">'+esc(c.content)+'</div><div class="time">'+fmtDate(c.created_at)+'</div></div>').join(''):''}
    <div style="margin-top:16px">
      <div class="form-group"><textarea id="tktReply" placeholder="补充说明..."></textarea></div>
      <button class="btn btn-green btn-block" onclick="addComment(${id})">发送</button>
    </div>
  `);
}

async function addComment(id) {
  const content = $('tktReply').value;
  if(!content) { toast('请输入内容'); return; }
  await fetch(API+'/tickets/'+id+'/comments',{method:'POST',headers:headers(),body:JSON.stringify({content})});
  toast('已发送'); showTicketDetail(id);
}

// ==================== 我的 ====================
function loadMe() {
  const app = $('app');
  if(!user) { showLoginForm(); return; }
  // 每次进入"我的"页面都刷新用户信息
  fetch(API+'/me',{headers:headers()}).then(r=>{
    if(r.status===401) { clearAuth(); showLoginForm(); throw 0; }
    if(r.ok) return r.json(); throw 0;
  }).then(u=>{
    user = u; localStorage.setItem('vuser',JSON.stringify(u)); updateUserBar(); renderMe();
  }).catch(()=>{ if(user) renderMe(); });
}
function renderMe() {
  const app = $('app');
  if(!user) { showLoginForm(); return; }
  const hasPhone = isPhone(user.phone);
  const tags = [];
  if(user.is_party_member) tags.push('🎖️ 党员');
  if(user.is_low_income) tags.push('🏠 低保户');
  if(user.is_five_guarantee) tags.push('🏠 五保户');
  if(user.is_disabled) tags.push('♿ 残疾人');
  if(user.is_military) tags.push('🎗️ 军属/退役');
  app.innerHTML = `
    <div class="me-info">
      <div class="avatar">${user.name?user.name[0]:'?'}</div>
      <div class="name">${user.name}</div>
      <div class="role">${user.position_label||user.role_label||user.role}${hasPhone?' · '+user.phone:''}</div>
    </div>
    <div class="card" style="margin-top:12px">
      <div style="font-size:13px;color:#666;line-height:2.2">
        <div>📱 手机号：${hasPhone?user.phone:'<span style="color:#e33">未绑定</span>'}</div>
        <div>👤 性别：${user.gender==='male'?'男':user.gender==='female'?'女':'<span style="color:#999">未填写</span>'}</div>
        <div>🎂 出生日期：${user.birth_date||'<span style="color:#999">未填写</span>'}</div>
        <div>🏷️ 民族：${user.ethnicity||'<span style="color:#999">未填写</span>'}</div>
        <div>📖 文化程度：${user.education||'<span style="color:#999">未填写</span>'}</div>
        <div>💍 婚姻状况：${{unmarried:'未婚',married:'已婚',divorced:'离异',widowed:'丧偶'}[user.marital_status]||'<span style="color:#999">未填写</span>'}</div>
        <div>📍 地址：${user.address||'<span style="color:#999">未填写</span>'}</div>
        <div>🪪 身份证：${user.id_card||'<span style="color:#999">未填写</span>'}</div>
        <div>🏘️ 所属小组：${user.group_name||'<span style="color:#999">未分配</span>'}</div>
        ${tags.length?'<div>🏷️ '+tags.join(' ')+'</div>':''}
        <div>📞 紧急联系人：${user.emergency_contact?user.emergency_contact+' '+user.emergency_phone:'<span style="color:#999">未填写</span>'}</div>
      </div>
      <button class="btn btn-green" style="margin-top:10px" onclick="showEditProfile()">编辑资料</button>
    </div>
    <div class="me-menu">
      <div class="item" onclick="switchTab('subsidy')"><span>📋 我的补贴申请</span><span class="arrow">›</span></div>
      <div class="item" onclick="switchTab('ticket')"><span>🔧 我的工单</span><span class="arrow">›</span></div>
      <div class="item" onclick="showChangePassword()"><span>🔑 修改密码</span><span class="arrow">›</span></div>
      ${['admin','secretary','director','deputy','committee','supervisor','accountant','group_leader','grid_worker','resident_official'].some(r=>(user.role||'').includes(r))?'<div class="item" onclick="location.href=\'/admin/\'"><span>⚙️ 进入管理后台</span><span class="arrow">›</span></div>':''}
      <div class="item" onclick="clearAuth();loadMe()" style="color:var(--red)"><span>退出登录</span><span class="arrow">›</span></div>
    </div>
  `;
}

function showEditProfile() {
  const phoneVal = isPhone(user.phone) ? user.phone : '';
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <h2>编辑资料</h2>
    <div class="form-group"><label>姓名</label><input id="editName" value="${user.name||''}"></div>
    <div class="form-group"><label>性别</label><select id="editGender"><option value="">未填写</option><option value="male" ${user.gender==='male'?'selected':''}>男</option><option value="female" ${user.gender==='female'?'selected':''}>女</option></select></div>
    <div class="form-group"><label>出生日期</label><input id="editBirth" type="date" value="${user.birth_date||''}"></div>
    <div class="form-group"><label>手机号</label><input id="editPhone" value="${phoneVal}" placeholder="11位手机号" maxlength="11"></div>
    <div class="form-group"><label>身份证号</label><input id="editIDCard" value="${user.id_card||''}"></div>
    <div class="form-group"><label>地址</label><input id="editAddr" value="${user.address||''}"></div>
    <div class="form-group"><label>紧急联系人</label><input id="editEmName" value="${user.emergency_contact||''}" placeholder="姓名"></div>
    <div class="form-group"><label>紧急联系电话</label><input id="editEmPhone" value="${user.emergency_phone||''}" placeholder="电话"></div>
    <div class="form-group"><label>微信号</label><input id="editWechat" value="${user.wechat_id||''}"></div>
    <button class="btn btn-green btn-block" onclick="saveProfile()">保存</button>
  `);
}

async function saveProfile() {
  const name = $('editName').value.trim();
  const phone = $('editPhone').value.trim();
  if(!name) { toast('姓名不能为空'); return; }
  if(phone && phone.length !== 11) { toast('手机号需11位'); return; }

  const body = {
    name,
    gender: $('editGender').value,
    birth_date: $('editBirth').value,
    id_card: $('editIDCard').value.trim(),
    address: $('editAddr').value.trim(),
    emergency_contact: $('editEmName').value.trim(),
    emergency_phone: $('editEmPhone').value.trim(),
    wechat_id: $('editWechat').value.trim(),
  };
  const res = await fetch(API+'/me',{method:'PUT',headers:headers(),body:JSON.stringify(body)});
  const data = await res.json();
  if(data.error) { toast(data.error); return; }

  // 手机号变化则绑定
  const oldPhone = isPhone(user.phone) ? user.phone : '';
  if(phone && phone !== oldPhone) {
    const res2 = await fetch(API+'/me/bindphone',{method:'POST',headers:headers(),body:JSON.stringify({phone})});
    const data2 = await res2.json();
    if(data2.error) { toast(data2.error); return; }
  }

  // 刷新
  const meRes = await fetch(API+'/me',{headers:headers()});
  user = await meRes.json();
  localStorage.setItem('vuser', JSON.stringify(user));
  updateUserBar();
  closeDetail();
  toast('保存成功');
  loadMe();
}

function showChangePassword() {
  const hasPassword = user.password_set;
  showDetail(`
    <span class="close" onclick="closeDetail()">&times;</span>
    <h2>${hasPassword?'修改密码':'设置密码'}</h2>
    ${hasPassword?'<div class="form-group"><label>旧密码</label><input id="oldPwd" type="password" placeholder="输入旧密码"></div>':''}
    <div class="form-group"><label>新密码</label><input id="newPwd" type="password" placeholder="至少6位"></div>
    <div class="form-group"><label>确认新密码</label><input id="newPwd2" type="password" placeholder="再次输入"></div>
    <p id="pwdErr" style="color:var(--red);font-size:12px;margin-bottom:8px"></p>
    <button class="btn btn-green btn-block" onclick="doChangePassword()">确认</button>
  `);
}

async function doChangePassword() {
  const oldPwd = $('oldPwd') ? $('oldPwd').value : '';
  const newPwd = $('newPwd').value;
  const newPwd2 = $('newPwd2').value;
  if(newPwd.length < 6) { $('pwdErr').textContent='新密码至少6位'; return; }
  if(newPwd !== newPwd2) { $('pwdErr').textContent='两次密码不一致'; return; }
  const res = await fetch(API+'/me/password',{method:'POST',headers:headers(),body:JSON.stringify({old_password:oldPwd,new_password:newPwd})});
  const data = await res.json();
  if(data.error) { $('pwdErr').textContent=data.error; return; }
  closeDetail();
  toast('密码修改成功');
}

function showLoginForm() {
  $('app').innerHTML = `
    <div class="login-box">
      <h2>🏘️ 登录</h2>
      <div class="form-group"><label>手机号</label><input id="loginPhone" placeholder="手机号或admin"></div>
      <div class="form-group"><label>密码</label><input id="loginPwd" type="password" placeholder="密码"></div>
      <p id="loginErr" style="color:var(--red);font-size:12px;margin-bottom:8px"></p>
      <button class="btn btn-green btn-block" onclick="doLogin()">登 录</button>
      <div style="display:flex;justify-content:center;margin-top:12px">
        <div class="switch" onclick="showRegisterForm()">没有账号？注册</div>
      </div>
    </div>
  `;
}

function showRegisterForm() {
  $('app').innerHTML = `
    <div class="login-box">
      <h2>🏘️ 注册</h2>
      <div class="form-group"><label>手机号</label><input id="regPhone" placeholder="手机号"></div>
      <div class="form-group"><label>姓名</label><input id="regName" placeholder="真实姓名"></div>
      <div class="form-group"><label>密码</label><input id="regPwd" type="password" placeholder="设置密码"></div>
      <div class="form-group"><label>确认密码</label><input id="regPwd2" type="password" placeholder="再次输入密码"></div>
      <p id="regErr" style="color:var(--red);font-size:12px;margin-bottom:8px"></p>
      <button class="btn btn-green btn-block" onclick="doRegister()">注 册</button>
      <div class="switch" onclick="showLoginForm()">已有账号？点击登录</div>
    </div>
  `;
}



async function doLogin() {
  const phone = $('loginPhone').value, pwd = $('loginPwd').value;
  if(!phone||!pwd) { $('loginErr').textContent='请填写账号和密码'; return; }
  const res = await fetch(API+'/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({phone,password:pwd})});
  const data = await res.json();
  if(data.token) { saveAuth(data.token, data.user); toast('登录成功'); loadMe(); }
  else { $('loginErr').textContent = data.error||'登录失败'; }
}

async function doRegister() {
  const phone=$('regPhone').value, name=$('regName').value, pwd=$('regPwd').value, pwd2=$('regPwd2').value;
  if(!phone||!name||!pwd) { $('regErr').textContent='请填写完整'; return; }
  if(pwd!==pwd2) { $('regErr').textContent='两次密码不一致'; return; }
  const res = await fetch(API+'/register',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({phone,name,password:pwd})});
  const data = await res.json();
  if(data.token) { saveAuth(data.token, data.user); toast('注册成功'); loadMe(); }
  else { $('regErr').textContent = data.error||'注册失败'; }
}

function needLoginHtml(action) {
  return `<div class="empty">
    <div style="font-size:40px;margin-bottom:12px">🔒</div>
    <div>登录后可${action}</div>
    <button class="btn btn-green" style="margin-top:16px" onclick="switchTab('me')">去登录</button>
  </div>`;
}

load();
