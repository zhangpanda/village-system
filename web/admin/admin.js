const API = location.origin + '/api';
let token = localStorage.getItem('token') || '';
let currentSec = 'notices';
let currentUser = null;
const stateMap = {
  draft:'草稿',pending_review:'待审核',published:'已发布',
  submitted:'已提交',committee_review:'村委初审中',secretary_review:'村支书终审中',
  approved:'已通过',rejected:'已驳回',
  open:'待处理',assigned:'已分配',processing:'处理中',resolved:'已解决',closed:'已关闭'
};
const roleMap = {admin:'系统管理员',secretary:'党支部书记',director:'村委会主任',deputy:'副书记/副主任',committee:'两委委员',supervisor:'监委会委员',accountant:'村会计',group_leader:'村民小组长',villager:'村民'};
const catMap = {policy:'政策通知',activity:'村务活动',urgent:'紧急通知',meeting:'会议'};
const subTypeMap = {farming:'农业补贴',medical:'医疗救助',education:'教育补助',housing:'住房补贴',other:'其他'};
const ticketCatMap = {repair:'报修',complaint:'投诉',service:'便民服务',suggestion:'建议'};
const priorityMap = {low:'不急',normal:'普通',urgent:'紧急'};
const genderMap = {male:'男',female:'女'};
const fmt = v => (v/100).toLocaleString('zh-CN',{minimumFractionDigits:2});
const roleLevels = {admin:99,secretary:90,resident_official:88,director:85,deputy:70,supervisor:65,committee:60,accountant:50,group_leader:40,grid_worker:35,villager:10};
function myRoleLevel() { if(!currentUser) return 0; return Math.max(...currentUser.role.split(',').map(r=>roleLevels[r.trim()]||0)); }
function canRole(minRole) { return myRoleLevel() >= (roleLevels[minRole]||0); }

const esc = s => { const d=document.createElement('div'); d.textContent=s||''; return d.innerHTML; };
function headers() { return {'Content-Type':'application/json','Authorization':'Bearer '+token}; }
function toast(msg) { const t=document.getElementById('toast'); t.textContent=msg; t.style.display='block'; setTimeout(()=>t.style.display='none',2000); }
function closeModal() { document.getElementById('modal').classList.remove('show'); document.body.style.overflow=''; }
function showModal(html) { document.getElementById('modalBody').innerHTML=html; document.getElementById('modal').classList.add('show'); document.body.style.overflow='hidden'; }

fetch(API+'/config').then(r=>r.json()).then(c=>{
  const name = c.village_name || '';
  document.title = (name?name+' · ':'') + '村务管理后台';
  document.getElementById('loginTitle').textContent = '🏘️ ' + (name?name+' · ':'') + '村务管理后台';
  document.getElementById('adminTitle').textContent = '🏘️ ' + (name?name+' · ':'') + '村务管理';
}).catch(()=>{});

// === Auth ===
async function doLogin() {
  const phone = document.getElementById('loginPhone').value;
  const pwd = document.getElementById('loginPwd').value;
  const res = await fetch(API+'/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({phone,password:pwd})});
  const data = await res.json();
  if(data.token) {
    token = data.token; currentUser = data.user;
    if(!canRole('grid_worker')) { token=''; alert('当前账号无管理后台权限'); location.href='/app/'; return; }
    localStorage.setItem('token',token);
    document.getElementById('loginPage').style.display='none';
    document.getElementById('adminPage').style.display='block';
    loadDashboard(); loadSection(); loadNotifyBadge();
  } else {
    document.getElementById('loginErr').textContent = data.error||'登录失败';
  }
}
function logout() { token=''; localStorage.removeItem('token'); location.reload(); }

async function initApp() {
  if(!token) return;
  // 有token时先隐藏登录页，避免闪烁
  document.getElementById('loginPage').style.display='none';
  try {
    const res = await fetch(API+'/me',{headers:headers()});
    if(!res.ok) throw new Error();
    currentUser = await res.json();
    if(!canRole('grid_worker')) { alert('当前账号无管理后台权限'); location.href='/app/'; return; }
    document.getElementById('adminPage').style.display='block';
    loadDashboard(); loadSection(); loadNotifyBadge();
  } catch(e) { token=''; localStorage.removeItem('token'); document.getElementById('loginPage').style.display=''; }
}
initApp();

document.querySelectorAll('.nav-tab').forEach(el => {
  el.onclick = () => {
    document.querySelectorAll('.nav-tab').forEach(t=>t.classList.remove('active'));
    el.classList.add('active');
    currentSec = el.dataset.sec;
    currentFilter = '';
    currentPage = 1;
    loadSection();
  };
});

// === Dashboard ===
async function loadDashboard() {
  try {
    const res = await fetch(API+'/admin/dashboard',{headers:headers()});
    if(!res.ok) throw new Error();
    const d = await res.json();
  document.getElementById('dashboard').innerHTML = `
    <div class="dash-card" onclick="switchTo('users')" style="cursor:pointer"><div class="num" style="color:var(--green)">${d.user_count}</div><div class="label">用户 / ${d.group_count}个小组</div></div>
    <div class="dash-card" onclick="switchTo('notices')" style="cursor:pointer"><div class="num">${d.notice_count}${d.notice_pending?'<span style="font-size:12px;color:var(--orange)"> +'+d.notice_pending+'待审</span>':''}</div><div class="label">已发布公告</div></div>
    <div class="dash-card" onclick="switchTo('subsidies','submitted')" style="cursor:pointer"><div class="num" style="color:var(--orange)">${d.subsidy_pending}<span style="font-size:12px;color:#999"> / ${d.subsidy_total}</span></div><div class="label">待审补贴 ›</div></div>
    <div class="dash-card" onclick="switchTo('tickets','open')" style="cursor:pointer"><div class="num" style="color:var(--red)">${d.ticket_open}<span style="font-size:12px;color:#999"> / ${d.ticket_total}</span></div><div class="label">待处理工单 ›</div></div>
    ${d.finance_pending?'<div class="dash-card" onclick="switchTo(\'finance\')" style="cursor:pointer"><div class="num" style="color:var(--orange)">'+d.finance_pending+'</div><div class="label">待审财务 ›</div></div>':''}
  `;
  } catch(e) { token=''; localStorage.removeItem('token'); location.reload(); }
}

function switchTo(sec, filter) {
  document.querySelectorAll('.nav-tab').forEach(t=>{t.classList.remove('active');if(t.dataset.sec===sec)t.classList.add('active');});
  currentSec = sec;
  currentFilter = filter || '';
  loadSection();
}
let currentFilter = '';
let currentPage = 1;
const PAGE_SIZE = 10;

function paginator(page, total) {
  const totalPages = Math.ceil(total / PAGE_SIZE);
  let h = '<div style="display:flex;justify-content:center;align-items:center;gap:12px;padding:12px;font-size:13px;color:#999">';
  if(totalPages <= 1) return h + `共 ${total} 条</div>`;
  if(page > 1) h += `<a style="color:var(--green);cursor:pointer" onclick="currentPage--;loadSection()">‹ 上一页</a>`;
  h += `<span>第 ${page}/${totalPages} 页 · 共 ${total} 条</span>`;
  if(page < totalPages) h += `<a style="color:var(--green);cursor:pointer" onclick="currentPage++;loadSection()">下一页 ›</a>`;
  return h + '</div>';
}

function filterBar(tabs, current, onclick) {
  return '<div style="display:flex;gap:6px;margin-bottom:12px;flex-wrap:wrap">'+tabs.map(([v,l])=>`<button class="btn ${current===v?'btn-green':'btn-outline'}" style="padding:4px 12px;font-size:12px" onclick="currentFilter='${v}';currentPage=1;${onclick}">${l}</button>`).join('')+'</div>';
}

function searchBar(placeholder, onclick) {
  return `<div style="display:flex;gap:8px;margin-bottom:12px"><input id="adminSearch" placeholder="${placeholder}" style="flex:1;padding:6px 10px;border:1px solid #ddd;border-radius:6px;font-size:13px" onkeydown="if(event.key==='Enter'){currentPage=1;${onclick}}"><button class="btn btn-green" style="padding:6px 12px;font-size:12px" onclick="currentPage=1;${onclick}">搜索</button></div>`;
}

// === Section Router ===
function loadSection() {
  const fab = document.getElementById('fabBtn');
  fab.style.display = ['notices','finance','groups','households'].includes(currentSec) ? '' : 'none';
  switch(currentSec) {
    case 'notices': loadNoticesAdmin(); break;
    case 'finance': loadFinanceAdmin(); break;
    case 'subsidies': loadSubsidiesAdmin(); break;
    case 'tickets': loadTicketsAdmin(); break;
    case 'users': loadUsersAdmin(); break;
    case 'households': loadHouseholdsAdmin(); break;
    case 'groups': loadGroupsAdmin(); break;
    case 'reports': loadReports(); break;
    case 'workflows': loadWorkflows(); break;
    case 'logs': loadWorkflowLogs(); break;
  }
}

// === Notices ===
async function loadNoticesAdmin() {
  const f = currentFilter;
  const {data,total} = await (await fetch(API+'/admin/notices?size='+PAGE_SIZE+'&page='+currentPage+(f?'&state='+f:''),{headers:headers()})).json();
  const sec = document.getElementById('section');
  let html = filterBar([['','全部'],['published','已发布'],['pending_review','待审核'],['draft','草稿'],['rejected','已驳回']], f, 'loadNoticesAdmin()');
  if(!data?.length) { sec.innerHTML=html+'<div style="text-align:center;color:#999;padding:30px">暂无公告</div>'; return; }
  html += data.map(n => `
    <div class="list-item" onclick="showNoticeDetail(${n.id})" style="cursor:pointer">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${n.pinned?'📌 ':''}${esc(n.title)}</div>
        <span class="status-tag ${n.workflow_state}">${stateMap[n.workflow_state]||n.workflow_state}</span>
      </div>
      <div class="meta"><span>${n.author} · ${catMap[n.category]||n.category} · 👁${n.views}</span></div>
    </div>
  `).join('');
  html += paginator(currentPage, total);
  sec.innerHTML = html;
}

async function showNoticeDetail(id) {
  const {notice:n, logs} = await (await fetch(API+'/notices/'+id)).json();
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3 style="margin-bottom:8px">${esc(n.title)}</h3>
    <div style="margin-bottom:12px">
      ${n.pinned?'<span class="status-tag approved">📌 置顶</span> ':''}
      <span class="status-tag ${n.workflow_state}">${stateMap[n.workflow_state]}</span>
      <span style="font-size:12px;color:#999;margin-left:8px">${catMap[n.category]||n.category}</span>
    </div>
    <div style="font-size:12px;color:#999;margin-bottom:12px">${esc(n.author)} · ${n.created_at} · 阅读 ${n.views}</div>
    <div style="font-size:14px;line-height:1.8;margin-bottom:16px;padding:12px;background:#fafafa;border-radius:8px" class="notice-body-html">${n.content}</div>
    ${n.reviewer_name?'<div style="font-size:12px;color:var(--blue);margin-bottom:8px">审核人：'+esc(n.reviewer_name)+(n.review_note?' — '+esc(n.review_note):'')+'</div>':''}
    ${logs?.length?'<div style="font-size:12px;color:#999;margin-bottom:12px"><b>操作日志：</b>'+logs.map(l=>'<div>'+esc(l.operator_name)+' '+esc(l.action)+(l.note?' ('+esc(l.note)+')':'')+'</div>').join('')+'</div>':''}
    ${n.workflow_state==='pending_review'&&canRole('deputy')?`
      <div class="form-group"><label>审核意见</label><textarea id="noticeReviewNote" placeholder="审核意见..."></textarea></div>
      <div class="action-btns">
        <button class="btn btn-green" onclick="reviewNotice(${n.id},'approve')">审核通过并发布</button>
        <button class="btn btn-red" onclick="reviewNotice(${n.id},'reject')">驳回</button>
      </div>
    `:''}
    <div class="action-btns" style="margin-top:8px">
      ${canRole('committee')?'<button class="btn btn-blue" onclick="editNotice('+n.id+')">编辑</button>':''}
      ${canRole('committee')?'<button class="btn '+(n.pinned?'btn-outline':'btn-blue')+'" onclick="togglePin('+n.id+','+!n.pinned+')">'+(n.pinned?'取消置顶':'📌 置顶')+'</button>':''}
      ${canRole('deputy')?'<button class="btn btn-red" onclick="deleteNotice('+n.id+')">删除</button>':''}
    </div>
  `);
}
async function reviewNotice(id, action) {
  const note = document.getElementById('noticeReviewNote')?.value || '';
  await fetch(API+'/admin/notices/'+id+'/review',{method:'PUT',headers:headers(),body:JSON.stringify({action,note})});
  closeModal(); toast(action==='approve'?'已发布':'已驳回'); loadNoticesAdmin(); loadDashboard();
}
async function editNotice(id) {
  const {notice:n} = await (await fetch(API+'/notices/'+id)).json();
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>编辑公告</h3>
    <div class="form-group"><label>标题</label><input id="enTitle" value="${esc(n.title)}"></div>
    <div class="form-group"><label>分类</label><select id="enCat">
      <option value="policy" ${n.category==='policy'?'selected':''}>政策通知</option>
      <option value="activity" ${n.category==='activity'?'selected':''}>村务活动</option>
      <option value="urgent" ${n.category==='urgent'?'selected':''}>紧急通知</option>
      <option value="meeting" ${n.category==='meeting'?'selected':''}>会议</option>
    </select></div>
    <div class="form-group"><label>内容</label><div id="editor-container"></div></div>
    <div class="form-group"><label style="display:flex;align-items:center;gap:8px"><input type="checkbox" id="enPinned" ${n.pinned?'checked':''} style="width:auto"> 置顶显示</label></div>
    <button class="btn btn-green btn-block" onclick="saveEditNotice(${id})">保存</button>
  `);
  initEditor(n.content);
}
async function saveEditNotice(id) {
  const content = quillEditor ? quillEditor.root.innerHTML : '';
  const body = {
    title: document.getElementById('enTitle').value,
    content: content,
    category: document.getElementById('enCat').value,
    pinned: document.getElementById('enPinned').checked,
  };
  if(!body.title||!content||content==='<p><br></p>') { toast('请填写标题和内容'); return; }
  await fetch(API+'/admin/notices/'+id,{method:'PUT',headers:headers(),body:JSON.stringify(body)});
  closeModal(); toast('保存成功'); loadNoticesAdmin();
}

async function togglePin(id, pinned) {
  await fetch(API+'/admin/notices/'+id,{method:'PUT',headers:headers(),body:JSON.stringify({pinned})});
  closeModal(); toast(pinned?'已置顶':'已取消置顶'); loadNoticesAdmin();
}
async function deleteNotice(id) {
  if(!confirm('确定删除？')) return;
  await fetch(API+'/admin/notices/'+id,{method:'DELETE',headers:headers()});
  closeModal(); toast('已删除'); loadNoticesAdmin(); loadDashboard();
}

// === Finance (with workflow) ===
async function loadFinanceAdmin() {
  const year = new Date().getFullYear();
  const [sumRes,listRes] = await Promise.all([
    fetch(API+'/finance/summary?year='+year),
    fetch(API+'/admin/finance?year='+year+'&size='+PAGE_SIZE+'&page='+currentPage,{headers:headers()})
  ]);
  const sum = await sumRes.json();
  const {data,total} = await listRes.json();
  const sec = document.getElementById('section');
  sec.innerHTML = `
    <div style="display:flex;gap:8px;margin-bottom:12px">
      <div class="dash-card" style="flex:1"><div class="num" style="color:var(--green);font-size:16px">¥${fmt(sum.total_income)}</div><div class="label">收入</div></div>
      <div class="dash-card" style="flex:1"><div class="num" style="color:var(--red);font-size:16px">¥${fmt(sum.total_expense)}</div><div class="label">支出</div></div>
      <div class="dash-card" style="flex:1"><div class="num" style="font-size:16px">¥${fmt(sum.balance)}</div><div class="label">结余</div></div>
    </div>
    <div style="margin-bottom:12px"><button class="btn btn-outline" style="padding:4px 12px;font-size:12px" onclick="downloadFile(API+'/admin/export/finance?year=${year}')">📥 导出${year}年财务</button> <button class="btn btn-outline" style="padding:4px 12px;font-size:12px" onclick="openPrintPage(API+'/admin/print/finance?year=${year}')">🖨️ 打印</button></div>
    ${(data||[]).map(r => `
      <div class="list-item" onclick="showFinanceDetail(${JSON.stringify(r).replace(/"/g,'&quot;')})" style="cursor:pointer">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span class="title" style="color:${r.type==='income'?'var(--green)':'var(--red)'}">${r.type==='income'?'+':'-'}¥${fmt(r.amount)}</span>
          <span class="status-tag ${r.workflow_state}">${stateMap[r.workflow_state]||r.workflow_state}</span>
        </div>
        <div style="font-size:12px;color:#666;margin-top:4px">${r.category} · ${r.remark} · ${r.date}</div>
      </div>
    `).join('')}
    ${paginator(currentPage, total)}
  `;
}
function showFinanceDetail(r) {
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3 style="color:${r.type==='income'?'var(--green)':'var(--red)'}">${r.type==='income'?'收入':'支出'}：¥${fmt(r.amount)}</h3>
    <div style="font-size:13px;color:#666;margin:12px 0;line-height:2">
      <div>分类：${r.category}</div>
      <div>日期：${r.date}</div>
      <div>备注：${r.remark||'无'}</div>
      <div>录入人：${r.author}</div>
      <div>状态：<span class="status-tag ${r.workflow_state}">${stateMap[r.workflow_state]}</span></div>
      ${r.reviewer_name?'<div>审核人：'+r.reviewer_name+(r.review_note?' — '+r.review_note:'')+'</div>':''}
      ${r.voucher?'<div>凭证：<img src="'+r.voucher+'" style="max-width:200px;border-radius:6px;margin-top:4px;cursor:pointer" onclick="window.open(this.src)"></div>':''}
    </div>
    ${r.workflow_state==='pending_review'&&canRole('supervisor')?`
      <div class="form-group"><label>审核意见</label><textarea id="finReviewNote" placeholder="审核意见..."></textarea></div>
      <div class="action-btns">
        <button class="btn btn-green" onclick="reviewFinance(${r.id},'approve')">审核通过</button>
        <button class="btn btn-red" onclick="reviewFinance(${r.id},'reject')">驳回</button>
      </div>
    `:''}
    ${canRole('deputy')?'<div class="action-btns" style="margin-top:8px"><button class="btn btn-red" onclick="deleteFinance('+r.id+')">删除</button></div>':''}
  `);
}

async function reviewFinance(id, action) {
  const note = document.getElementById('finReviewNote')?.value || '';
  await fetch(API+'/admin/finance/'+id+'/review',{method:'PUT',headers:headers(),body:JSON.stringify({action,note})});
  closeModal(); toast('已审核'); loadFinanceAdmin(); loadDashboard();
}
async function deleteFinance(id) {
  if(!confirm('确定删除？')) return;
  await fetch(API+'/admin/finance/'+id,{method:'DELETE',headers:headers()});
  closeModal(); toast('已删除'); loadFinanceAdmin(); loadDashboard();
}

// === Subsidies (two-level approval) ===
async function loadSubsidiesAdmin() {
  const filter = currentFilter;
  const url = API+'/subsidies?size='+PAGE_SIZE+'&page='+currentPage+(filter?'&state='+filter:'');
  const {data,total} = await (await fetch(url,{headers:headers()})).json();
  const sec = document.getElementById('section');
  const tabs = [['','全部'],['submitted','待初审'],['secretary_review','待终审'],['approved','已通过'],['rejected','已驳回']];
  let html = filterBar(tabs, filter, 'loadSubsidiesAdmin()');
  html += '<div style="margin-bottom:12px"><button class="btn btn-outline" style="padding:4px 12px;font-size:12px" onclick="downloadFile(API+\'/admin/export/subsidies\')">📥 导出补贴台账</button></div>';
  if(!data?.length) { sec.innerHTML=html+'<div style="text-align:center;color:#999;padding:30px">暂无补贴申请</div>'; return; }
  html += data.map(s => `
    <div class="list-item" onclick="showSubsidyDetail(${s.id})" style="cursor:pointer">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${esc(s.title)}</div>
        <span class="status-tag ${s.workflow_state}">${stateMap[s.workflow_state]||s.workflow_state}</span>
      </div>
      <div style="font-size:12px;color:#666;margin-top:4px">${s.applicant} · ¥${fmt(s.amount)} · ${subTypeMap[s.type]||s.type}</div>
    </div>
  `).join('');
  html += paginator(currentPage, total);
  sec.innerHTML = html;
}
async function showSubsidyDetail(id) {
  const {subsidy:s, logs} = await (await fetch(API+'/subsidies/'+id,{headers:headers()})).json();
  const imgs = (()=>{try{return JSON.parse(s.attachments||'[]')}catch(e){return []}})();
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>${esc(s.title)} <span class="status-tag ${s.workflow_state}">${stateMap[s.workflow_state]}</span></h3>
    <div style="font-size:13px;color:#666;margin-bottom:12px">
      <div>申请人：${s.applicant}</div>
      <div>补贴类型：${subTypeMap[s.type]||s.type}</div>
      <div>申请金额：<b style="color:var(--green)">¥${fmt(s.amount)}</b></div>
      <div>申请时间：${s.created_at}</div>
    </div>
    <div style="font-size:14px;line-height:1.7;margin-bottom:12px;white-space:pre-wrap"><b>申请理由：</b>${esc(s.reason)||'无'}</div>
    ${imgs.length?'<div style="margin-bottom:12px"><b>附件：</b><br>'+imgs.map(u=>'<img src="'+u+'" style="width:80px;height:80px;object-fit:cover;border-radius:6px;margin:2px;cursor:pointer" onclick="window.open(this.src)">').join('')+'</div>':''}
    ${s.committee_name?'<div style="font-size:13px;padding:8px;background:#f5f5f5;border-radius:6px;margin-bottom:8px"><b>村委初审：</b>'+esc(s.committee_name)+' — '+esc(s.committee_note||'无意见')+'</div>':''}
    ${s.secretary_name?'<div style="font-size:13px;padding:8px;background:#f0fff0;border-radius:6px;margin-bottom:8px"><b>村支书终审：</b>'+esc(s.secretary_name)+' — '+esc(s.secretary_note||'无意见')+'</div>':''}
    ${logs?.length?'<div style="font-size:12px;color:#999;margin-bottom:12px"><b>审批日志：</b>'+logs.map(l=>'<div>'+esc(l.operator_name)+' '+esc(l.action)+(l.note?' ('+esc(l.note)+')':'')+'</div>').join('')+'</div>':''}
    ${s.workflow_state==='submitted'&&canRole('committee')?`
      <div class="form-group"><label>初审意见</label><textarea id="reviewNote" placeholder="审批意见..."></textarea></div>
      <div class="action-btns">
        <button class="btn btn-green" onclick="committeeReview(${s.id},'approve')">初审通过</button>
        <button class="btn btn-red" onclick="committeeReview(${s.id},'reject')">初审驳回</button>
      </div>
    `:''}
    ${s.workflow_state==='secretary_review'&&canRole('secretary')?`
      <div class="form-group"><label>终审意见</label><textarea id="reviewNote" placeholder="审批意见..."></textarea></div>
      <div class="action-btns">
        <button class="btn btn-green" onclick="secretaryReview(${s.id},'approve')">终审通过</button>
        <button class="btn btn-red" onclick="secretaryReview(${s.id},'reject')">终审驳回</button>
      </div>
    `:''}
    <div class="action-btns" style="margin-top:8px"><button class="btn btn-outline" onclick="openPrintPage(API+'/admin/print/subsidy/${s.id}')">🖨️ 打印审批单</button></div>
  `);
}

async function committeeReview(id, action) {
  const note = document.getElementById('reviewNote')?.value || '';
  await fetch(API+'/admin/subsidies/'+id+'/committee-review',{method:'PUT',headers:headers(),body:JSON.stringify({action,note})});
  closeModal(); toast(action==='approve'?'已转村支书终审':'已驳回'); loadSubsidiesAdmin(); loadDashboard();
}
async function secretaryReview(id, action) {
  const note = document.getElementById('reviewNote')?.value || '';
  await fetch(API+'/admin/subsidies/'+id+'/secretary-review',{method:'PUT',headers:headers(),body:JSON.stringify({action,note})});
  closeModal(); toast(action==='approve'?'已通过':'已驳回'); loadSubsidiesAdmin(); loadDashboard();
}

// === Tickets (assign + process) ===
async function loadTicketsAdmin() {
  const filter = currentFilter;
  const url = API+'/tickets?size='+PAGE_SIZE+'&page='+currentPage+(filter?'&state='+filter:'');
  const {data,total} = await (await fetch(url,{headers:headers()})).json();
  const sec = document.getElementById('section');
  const tabs = [['','全部'],['open','待处理'],['assigned','已分配'],['processing','处理中'],['resolved','已解决'],['closed','已关闭']];
  let html = '<div style="display:flex;gap:6px;margin-bottom:12px;flex-wrap:wrap">'+tabs.map(([v,l])=>`<button class="btn ${filter===v?'btn-green':'btn-outline'}" style="padding:4px 12px;font-size:12px" onclick="currentFilter='${v}';currentPage=1;loadTicketsAdmin()">${l}</button>`).join('')+'</div>';
  if(!data?.length) { sec.innerHTML=html+'<div style="text-align:center;color:#999;padding:30px">暂无工单</div>'; return; }
  html += data.map(t => `
    <div class="list-item" onclick="showTicketAdmin(${t.id})" style="cursor:pointer">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${t.priority==='urgent'?'🔴 ':''}${esc(t.title)}</div>
        <span class="status-tag ${t.workflow_state}">${stateMap[t.workflow_state]||t.workflow_state}</span>
      </div>
      <div style="font-size:12px;color:#666;margin-top:4px">${t.submitter} · ${ticketCatMap[t.category]||t.category}${t.assignee?' · 处理人：'+t.assignee:''}</div>
    </div>
  `).join('');
  html += paginator(currentPage, total);
  sec.innerHTML = html;
}
async function showTicketAdmin(id) {
  const {ticket:t, comments, logs} = await (await fetch(API+'/tickets/'+id,{headers:headers()})).json();
  const imgs = (()=>{try{return JSON.parse(t.images||'[]')}catch(e){return []}})();
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>${esc(t.title)} <span class="status-tag ${t.workflow_state}">${stateMap[t.workflow_state]}</span></h3>
    <div style="font-size:13px;color:#666;margin-bottom:12px">${esc(t.submitter)} · ${ticketCatMap[t.category]||t.category} · ${priorityMap[t.priority]||t.priority}${t.assignee?' · 处理人：'+esc(t.assignee):''}</div>
    <div style="font-size:14px;line-height:1.7;white-space:pre-wrap;margin-bottom:12px">${esc(t.content)}</div>
    ${imgs.length?'<div style="margin-bottom:12px">'+imgs.map(u=>'<img src="'+u+'" style="width:80px;height:80px;object-fit:cover;border-radius:6px;margin:2px;cursor:pointer" onclick="window.open(this.src)">').join('')+'</div>':''}
    ${logs?.length?'<div style="font-weight:600;margin:8px 0">操作日志</div>'+logs.map(l=>'<div style="font-size:12px;color:#666;padding:4px 0">'+esc(l.operator_name)+' '+esc(l.action)+' → '+esc(stateMap[l.to_state] ?? l.to_state ?? '')+(l.note?' ('+esc(l.note)+')':'')+'</div>').join(''):''}
    ${comments?.length?'<div style="font-weight:600;margin:8px 0">回复记录</div>'+comments.map(c=>'<div style="padding:6px 0;border-bottom:1px solid #f0f0f0"><b style="font-size:13px">'+esc(c.user_name)+'</b><div style="font-size:13px;white-space:pre-wrap">'+esc(c.content)+'</div></div>').join(''):''}
    ${!['resolved','closed'].includes(t.workflow_state)?`<div style="margin-top:16px">
      <div class="form-group"><textarea id="ticketReply" placeholder="回复内容..."></textarea></div>
      <div class="action-btns">
        ${t.workflow_state==='open'?'<button class="btn btn-blue" onclick="assignTicket('+id+')">认领</button>':''}
        ${['assigned','open'].includes(t.workflow_state)?'<button class="btn btn-blue" onclick="replyTicket('+id+',\'processing\')">处理中</button>':''}
        ${['assigned','processing'].includes(t.workflow_state)?'<button class="btn btn-green" onclick="replyTicket('+id+',\'resolved\')">已解决</button>':''}
      </div>
    </div>`:''}
  `);
}
async function assignTicket(id) {
  await fetch(API+'/admin/tickets/'+id+'/assign',{method:'PUT',headers:headers(),body:JSON.stringify({})});
  closeModal(); toast('已认领'); loadTicketsAdmin(); loadDashboard();
}
async function replyTicket(id, status) {
  const content = document.getElementById('ticketReply').value;
  if(content) await fetch(API+'/tickets/'+id+'/comments',{method:'POST',headers:headers(),body:JSON.stringify({content})});
  await fetch(API+'/admin/tickets/'+id+'/status',{method:'PUT',headers:headers(),body:JSON.stringify({status})});
  closeModal(); toast('已处理'); loadTicketsAdmin(); loadDashboard();
}

// === Users ===
let userFilters = {};
function setUserFilter(k,v) { userFilters[k]=v; currentPage=1; loadUsersAdmin(); }

function printRoster() {
  let url = API+'/admin/print/roster?';
  const q = document.getElementById('adminSearch')?.value||'';
  if(q) url += 'q='+encodeURIComponent(q)+'&';
  for(const [k,v] of Object.entries(userFilters)) { if(v) url += k+'='+encodeURIComponent(v)+'&'; }
  openPrintPage(url);
}

async function loadUsersAdmin() {
  const q = document.getElementById('adminSearch')?.value || '';
  let url = API+'/admin/users?size='+PAGE_SIZE+'&page='+currentPage;
  if(q) url += '&q='+encodeURIComponent(q);
  for(const [k,v] of Object.entries(userFilters)) { if(v) url += '&'+k+'='+encodeURIComponent(v); }
  const {data,total} = await (await fetch(url,{headers:headers()})).json();
  const sec = document.getElementById('section');
  const f = userFilters;
  let html = searchBar('搜索姓名或手机号...','loadUsersAdmin()');
  // 角色
  html += '<div style="display:flex;gap:4px;margin-bottom:6px;flex-wrap:wrap;font-size:12px">';
  html += '<span style="color:#999;line-height:24px">角色:</span>';
  [['','全部'],['secretary','村支书'],['committee','两委委员'],['group_leader','小组长'],['villager','村民']].forEach(([v,l])=>{
    html += `<button class="btn ${(f.role||'')===v?'btn-green':'btn-outline'}" style="padding:2px 8px;font-size:11px" onclick="setUserFilter('role','${v}')">${l}</button>`;
  });
  html += '</div>';
  // 性别
  html += '<div style="display:flex;gap:4px;margin-bottom:6px;flex-wrap:wrap;font-size:12px">';
  html += '<span style="color:#999;line-height:24px">性别:</span>';
  [['','全部'],['male','男'],['female','女']].forEach(([v,l])=>{
    html += `<button class="btn ${(f.gender||'')===v?'btn-green':'btn-outline'}" style="padding:2px 8px;font-size:11px" onclick="setUserFilter('gender','${v}')">${l}</button>`;
  });
  html += '</div>';
  // 特殊身份
  html += '<div style="display:flex;gap:4px;margin-bottom:6px;flex-wrap:wrap;font-size:12px">';
  html += '<span style="color:#999;line-height:24px">身份:</span>';
  [['','全部'],['party','党员'],['low_income','低保'],['five_guarantee','五保'],['disabled','残疾'],['military','军属']].forEach(([v,l])=>{
    html += `<button class="btn ${(f.tag||'')===v?'btn-green':'btn-outline'}" style="padding:2px 8px;font-size:11px" onclick="setUserFilter('tag','${v}')">${l}</button>`;
  });
  html += '</div>';
  // 文化程度
  html += '<div style="display:flex;gap:4px;margin-bottom:6px;flex-wrap:wrap;font-size:12px">';
  html += '<span style="color:#999;line-height:24px">学历:</span>';
  [['','全部'],['文盲','文盲'],['小学','小学'],['初中','初中'],['高中','高中'],['大专','大专'],['本科','本科'],['硕士及以上','硕士+']].forEach(([v,l])=>{
    html += `<button class="btn ${(f.education||'')===v?'btn-green':'btn-outline'}" style="padding:2px 8px;font-size:11px" onclick="setUserFilter('education','${v}')">${l}</button>`;
  });
  html += '</div>';
  // 婚姻
  html += '<div style="display:flex;gap:4px;margin-bottom:10px;flex-wrap:wrap;font-size:12px">';
  html += '<span style="color:#999;line-height:24px">婚姻:</span>';
  [['','全部'],['unmarried','未婚'],['married','已婚'],['divorced','离异'],['widowed','丧偶']].forEach(([v,l])=>{
    html += `<button class="btn ${(f.marital_status||'')===v?'btn-green':'btn-outline'}" style="padding:2px 8px;font-size:11px" onclick="setUserFilter('marital_status','${v}')">${l}</button>`;
  });
  html += '</div>';
  html += '<div style="display:flex;gap:8px;margin-bottom:12px"><button class="btn btn-green" style="padding:8px 16px;font-size:13px" onclick="showCreateUserForm()">+ 创建用户</button><button class="btn btn-outline" style="padding:8px 16px;font-size:13px" onclick="downloadFile(API+\'/admin/export/users\')">📥 导出</button><button class="btn btn-outline" style="padding:8px 16px;font-size:13px" onclick="showImportUsers()">📤 导入</button><button class="btn btn-outline" style="padding:8px 16px;font-size:13px" onclick="printRoster()">🖨️ 花名册</button></div>';
  if(!data?.length) { sec.innerHTML=html+'<div style="text-align:center;color:#999;padding:30px">暂无用户</div>'; return; }
  html += (data||[]).map(u => `
    <div class="list-item" style="cursor:pointer" onclick='showEditUserForm(${JSON.stringify(u).replace(/'/g,"&#39;")})'>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${esc(u.name)} <span style="font-size:11px;color:#999">${u.phone}</span></div>
        <span class="status-tag approved">${u.role_label||u.role}</span>
      </div>
      <div style="font-size:12px;color:#666;margin-top:4px">${u.gender?genderMap[u.gender]||'':''}${u.gender?' · ':''}${u.address||'未填写地址'}${u.group_name?' · '+u.group_name:''}</div>
      <div style="font-size:11px;margin-top:4px">
        <span style="color:var(--blue)">编辑</span>
        <span style="margin-left:12px;color:#e33" onclick="event.stopPropagation();resetUserPwd(${u.id},'${esc(u.name)}')">重置密码</span>
      </div>
    </div>
  `).join('');
  html += paginator(currentPage, total);
  sec.innerHTML = html;
}

function showCreateUserForm() {
  const roleOpts = Object.entries(roleMap).map(([k,v])=>`<option value="${k}">${v}</option>`).join('');
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>创建用户</h3>
    <div class="form-group"><label>手机号</label><input id="uPhone"></div>
    <div class="form-group"><label>姓名</label><input id="uName"></div>
    <div class="form-group"><label>密码</label><input id="uPwd" type="password" value="123456"></div>
    <div class="form-group"><label>角色</label><select id="uRole">${roleOpts}</select></div>
    <div class="form-group"><label>地址</label><input id="uAddr"></div>
    <button class="btn btn-green btn-block" onclick="submitCreateUser()">创建</button>
  `);
}
async function submitCreateUser() {
  const phone=document.getElementById('uPhone').value, name=document.getElementById('uName').value, pwd=document.getElementById('uPwd').value;
  if(!phone||!name||!pwd) { toast('必填项不能为空'); return; }
  const res = await fetch(API+'/register',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({phone,name,password:pwd})});
  const data = await res.json();
  if(data.error) { toast(data.error); return; }
  const role=document.getElementById('uRole').value, addr=document.getElementById('uAddr').value;
  if(role!=='villager'||addr) {
    await fetch(API+'/admin/users/'+data.user.id,{method:'PUT',headers:headers(),body:JSON.stringify({name,role,address:addr,active:true})});
  }
  closeModal(); toast('创建成功'); loadUsersAdmin(); loadDashboard();
}

function showEditUserForm(u) {
  var userRoles = (u.role||'').split(',').map(function(r){return r.trim()});
  var roleChecks = '<div style="display:grid;grid-template-columns:1fr 1fr;gap:4px 16px">'+Object.entries(roleMap).map(([k,v])=>`<label style="display:flex;align-items:center;gap:6px;font-size:13px"><input type="checkbox" class="role-cb" value="${k}" ${userRoles.includes(k)?'checked':''} style="width:auto"> ${v}</label>`).join('')+'</div>';
  Promise.all([fetch(API+'/groups').then(r=>r.json()), fetch(API+'/admin/households?size=200',{headers:headers()}).then(r=>r.json())]).then(([gd,hd])=>{
    var groups = gd.data||[];
    var groupDl = groups.map(g=>`<option value="${g.name}" data-id="${g.id}">`).join('');
    var curGroup = groups.find(g=>g.id===u.group_id);
    var hhs = hd.data||[];
    var hhDl = hhs.map(h=>`<option value="${h.household_no} · ${h.head_name||'无户主'}" data-id="${h.id}">`).join('');
    var curHH = hhs.find(h=>h.id===u.household_id);
    showModal(`
      <span class="close" onclick="closeModal()">&times;</span>
      <h3>编辑用户</h3>
      <div class="form-group"><label>手机号</label><input value="${u.phone}" disabled style="background:#f5f5f5"></div>
      <div class="form-group"><label>姓名</label><input id="euName" value="${u.name}"></div>
      <div class="form-group"><label>性别</label><select id="euGender"><option value="">未填写</option><option value="male" ${u.gender==='male'?'selected':''}>男</option><option value="female" ${u.gender==='female'?'selected':''}>女</option></select></div>
      <div class="form-group"><label>出生日期</label><input id="euBirth" type="date" value="${u.birth_date||''}"></div>
      <div class="form-group"><label>民族</label><input id="euEthnicity" value="${u.ethnicity||'汉族'}"></div>
      <div class="form-group"><label>文化程度</label><select id="euEducation"><option value="">未填写</option><option ${u.education==='文盲'?'selected':''}>文盲</option><option ${u.education==='小学'?'selected':''}>小学</option><option ${u.education==='初中'?'selected':''}>初中</option><option ${u.education==='高中'?'selected':''}>高中</option><option ${u.education==='大专'?'selected':''}>大专</option><option ${u.education==='本科'?'selected':''}>本科</option><option ${u.education==='硕士及以上'?'selected':''}>硕士及以上</option></select></div>
      <div class="form-group"><label>婚姻状况</label><select id="euMarital"><option value="">未填写</option><option value="unmarried" ${u.marital_status==='unmarried'?'selected':''}>未婚</option><option value="married" ${u.marital_status==='married'?'selected':''}>已婚</option><option value="divorced" ${u.marital_status==='divorced'?'selected':''}>离异</option><option value="widowed" ${u.marital_status==='widowed'?'selected':''}>丧偶</option></select></div>
      <div class="form-group"><label>角色（可多选）</label><div style="padding:4px 0">${roleChecks}</div></div>
      <div class="form-group"><label>所属小组</label><input id="euGroup" list="dlGroup" value="${curGroup?curGroup.name:''}" placeholder="输入搜索..."><datalist id="dlGroup">${groupDl}</datalist></div>
      <div class="form-group"><label>所属户籍</label><input id="euHousehold" list="dlHH" value="${curHH?curHH.household_no+' · '+(curHH.head_name||'无户主'):''}" placeholder="输入搜索..."><datalist id="dlHH">${hhDl}</datalist></div>
      <div class="form-group"><label>地址</label><input id="euAddr" value="${u.address||''}"></div>
      <div class="form-group"><label>身份证号</label><input id="euIDCard" value="${u.id_card||''}"></div>
      <div class="form-group"><label>微信号</label><input id="euWechat" value="${u.wechat_id||''}"></div>
      <div class="form-group"><label>紧急联系人</label><input id="euEmName" value="${u.emergency_contact||''}"></div>
      <div class="form-group"><label>紧急联系电话</label><input id="euEmPhone" value="${u.emergency_phone||''}"></div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:4px 16px;margin-bottom:12px">
        <label style="display:flex;align-items:center;gap:6px;font-size:13px"><input type="checkbox" class="eu-tag" id="euParty" ${u.is_party_member?'checked':''} style="width:auto"> 党员</label>
        <label style="display:flex;align-items:center;gap:6px;font-size:13px"><input type="checkbox" class="eu-tag" id="euLowIncome" ${u.is_low_income?'checked':''} style="width:auto"> 低保户</label>
        <label style="display:flex;align-items:center;gap:6px;font-size:13px"><input type="checkbox" class="eu-tag" id="euFiveG" ${u.is_five_guarantee?'checked':''} style="width:auto"> 五保户</label>
        <label style="display:flex;align-items:center;gap:6px;font-size:13px"><input type="checkbox" class="eu-tag" id="euDisabled" ${u.is_disabled?'checked':''} style="width:auto"> 残疾人</label>
        <label style="display:flex;align-items:center;gap:6px;font-size:13px"><input type="checkbox" class="eu-tag" id="euMilitary" ${u.is_military?'checked':''} style="width:auto"> 军属/退役</label>
      </div>
      <div class="form-group"><label>备注</label><input id="euRemark" value="${u.remark||''}"></div>
      <div class="form-group"><label style="display:flex;align-items:center;gap:6px"><input type="checkbox" id="euActive" ${u.active?'checked':''} style="width:auto"> 启用</label></div>
      <button class="btn btn-green btn-block" onclick="submitEditUser(${u.id})">保存</button>
    `);
  });
}
async function submitEditUser(id) {
  var roles = Array.from(document.querySelectorAll('.role-cb:checked')).map(cb=>cb.value).join(',') || 'villager';
  const body = {
    name: document.getElementById('euName').value,
    gender: document.getElementById('euGender').value,
    birth_date: document.getElementById('euBirth').value,
    ethnicity: document.getElementById('euEthnicity').value,
    education: document.getElementById('euEducation').value,
    marital_status: document.getElementById('euMarital').value,
    role: roles,
    address: document.getElementById('euAddr').value,
    id_card: document.getElementById('euIDCard').value,
    wechat_id: document.getElementById('euWechat').value,
    emergency_contact: document.getElementById('euEmName').value,
    emergency_phone: document.getElementById('euEmPhone').value,
    is_party_member: document.getElementById('euParty').checked,
    is_low_income: document.getElementById('euLowIncome').checked,
    is_five_guarantee: document.getElementById('euFiveG').checked,
    is_disabled: document.getElementById('euDisabled').checked,
    is_military: document.getElementById('euMilitary').checked,
    remark: document.getElementById('euRemark').value,
    group_id: getPickedId('euGroup'),
    household_id: getPickedId('euHousehold'),
    active: document.getElementById('euActive').checked,
  };
  await fetch(API+'/admin/users/'+id,{method:'PUT',headers:headers(),body:JSON.stringify(body)});
  closeModal(); toast('保存成功'); loadUsersAdmin();
}
async function resetUserPwd(id, name) {
  if(!confirm('确定将 '+name+' 的密码重置为 123456？')) return;
  await fetch(API+'/admin/users/'+id+'/reset-password',{method:'POST',headers:headers()});
  toast('已重置为 123456');
}

// === Groups ===
async function loadGroupsAdmin() {
  const {data} = await (await fetch(API+'/groups')).json();
  const sec = document.getElementById('section');
  sec.innerHTML = (data||[]).map(g => `
    <div class="list-item">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${esc(g.name)}</div>
        <span style="font-size:12px;color:#999">${g.member_count||0} 人</span>
      </div>
      <div style="font-size:12px;color:#666;margin-top:4px">组长：${g.leader_name||'未设置'}</div>
      <div class="action-btns">
        <button class="btn btn-blue" onclick="editGroup(${g.id},'${esc(g.name)}',${g.leader_id})">编辑</button>
        <button class="btn btn-red" onclick="deleteGroup(${g.id})">删除</button>
      </div>
    </div>
  `).join('') || '<div style="text-align:center;color:#999;padding:30px">暂无小组</div>';
}
function editGroup(id, name, leaderId) {
  fetch(API+'/admin/users?size=100',{headers:headers()}).then(r=>r.json()).then(ud=>{
    const users = ud.data||[];
    const leaderOpts = '<option value="0">未设置</option>'+users.map(u=>`<option value="${u.id}" ${u.id===leaderId?'selected':''}>${u.name} (${u.phone})</option>`).join('');
    showModal(`
      <span class="close" onclick="closeModal()">&times;</span>
      <h3>${id?'编辑':'创建'}小组</h3>
      <div class="form-group"><label>小组名称</label><input id="gName" value="${name||''}"></div>
      <div class="form-group"><label>组长</label><select id="gLeader">${leaderOpts}</select></div>
      <button class="btn btn-green btn-block" onclick="saveGroup(${id})">${id?'保存':'创建'}</button>
    `);
  });
}
async function saveGroup(id) {
  const body = { name: document.getElementById('gName').value, leader_id: parseInt(document.getElementById('gLeader').value)||0 };
  if(!body.name) { toast('名称不能为空'); return; }
  if(id) await fetch(API+'/admin/groups/'+id,{method:'PUT',headers:headers(),body:JSON.stringify(body)});
  else await fetch(API+'/admin/groups',{method:'POST',headers:headers(),body:JSON.stringify(body)});
  closeModal(); toast('保存成功'); loadGroupsAdmin(); loadDashboard();
}
async function deleteGroup(id) {
  if(!confirm('确定删除？')) return;
  await fetch(API+'/admin/groups/'+id,{method:'DELETE',headers:headers()});
  toast('已删除'); loadGroupsAdmin(); loadDashboard();
}

// === Create Forms ===
function showCreateForm() {
  if(currentSec==='groups') { editGroup(0,'',0); return; }
  if(currentSec==='households') { showCreateHouseholdForm(); return; }
  if(currentSec==='notices') showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>发布公告</h3>
    <div class="form-group"><label>标题</label><input id="nTitle"></div>
    <div class="form-group"><label>分类</label><select id="nCat"><option value="policy">政策通知</option><option value="activity">村务活动</option><option value="urgent">紧急通知</option><option value="meeting">会议</option></select></div>
    <div class="form-group"><label>内容</label><div id="editor-container"></div></div>
    <div class="form-group"><label style="display:flex;align-items:center;gap:8px"><input type="checkbox" id="nPinned" style="width:auto"> 置顶显示</label></div>
    <button class="btn btn-green btn-block" onclick="submitNotice()">发布</button>
  `);
  if(currentSec==='notices') initEditor('');
  else if(currentSec==='finance') showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>添加财务记录</h3>
    <div class="form-group"><label>类型</label><select id="fType"><option value="income">收入</option><option value="expense">支出</option></select></div>
    <div class="form-group"><label>金额（元）</label><input id="fAmount" type="number" step="0.01"></div>
    <div class="form-group"><label>分类</label><input id="fCat" placeholder="如：上级拨款、基础设施"></div>
    <div class="form-group"><label>日期</label><input id="fDate" type="date"></div>
    <div class="form-group"><label>备注</label><textarea id="fRemark" style="min-height:80px"></textarea></div>
    <button class="btn btn-green btn-block" onclick="submitFinance()">保存</button>
  `);
}
let quillEditor = null;
function initEditor(content) {
  setTimeout(()=>{
    quillEditor = new Quill('#editor-container', {
      theme: 'snow',
      modules: {
        toolbar: [
          ['bold','italic','underline','strike'],
          [{'header':[1,2,3,false]}],
          [{'list':'ordered'},{'list':'bullet'}],
          [{'color':[]},{'background':[]}],
          ['image','link'],
          ['clean']
        ]
      },
      placeholder: '输入公告内容...'
    });
    if(content) {
      if(content.startsWith('<')) quillEditor.root.innerHTML = content;
      else quillEditor.setText(content);
    }
    // 图片上传
    quillEditor.getModule('toolbar').addHandler('image', function(){
      const input = document.createElement('input');
      input.type='file'; input.accept='image/*';
      input.onchange = async function(){
        const fd = new FormData(); fd.append('file', input.files[0]);
        const res = await fetch(API+'/upload',{method:'POST',headers:{'Authorization':'Bearer '+token},body:fd});
        const data = await res.json();
        if(data.url) {
          const range = quillEditor.getSelection(true);
          quillEditor.insertEmbed(range.index, 'image', data.url);
        }
      };
      input.click();
    });
  }, 100);
}

async function submitNotice() {
  const content = quillEditor ? quillEditor.root.innerHTML : '';
  const body = {
    title: document.getElementById('nTitle').value,
    content: content,
    category: document.getElementById('nCat').value,
    pinned: document.getElementById('nPinned').checked,
  };
  if(!body.title||!content||content==='<p><br></p>') { toast('请填写标题和内容'); return; }
  await fetch(API+'/admin/notices',{method:'POST',headers:headers(),body:JSON.stringify(body)});
  closeModal(); toast('已提交'); loadNoticesAdmin(); loadDashboard();
}
async function submitFinance() {
  const amount = Math.round(parseFloat(document.getElementById('fAmount').value)*100);
  const body = { type:document.getElementById('fType').value, amount, category:document.getElementById('fCat').value, date:document.getElementById('fDate').value, remark:document.getElementById('fRemark').value };
  if(!amount||!body.date) { toast('请填写金额和日期'); return; }
  await fetch(API+'/admin/finance',{method:'POST',headers:headers(),body:JSON.stringify(body)});
  closeModal(); toast('已提交'); loadFinanceAdmin(); loadDashboard();
}

// === Admin Profile ===
async function showAdminProfile() {
  const me = await (await fetch(API+'/me',{headers:headers()})).json();
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>个人设置 <span style="font-size:12px;color:#999">${me.role_label||me.role}</span></h3>
    <div class="form-group"><label>姓名</label><input id="apName" value="${me.name||''}"></div>
    <div class="form-group"><label>手机号</label><input id="apPhone" value="${me.phone||''}" maxlength="11"></div>
    <div class="form-group"><label>地址</label><input id="apAddr" value="${me.address||''}"></div>
    <button class="btn btn-green btn-block" onclick="saveAdminProfile()">保存资料</button>
    <hr style="margin:16px 0">
    <h3>修改密码</h3>
    <div class="form-group"><label>旧密码</label><input id="apOldPwd" type="password"></div>
    <div class="form-group"><label>新密码</label><input id="apNewPwd" type="password" placeholder="至少6位"></div>
    <div class="form-group"><label>确认密码</label><input id="apNewPwd2" type="password"></div>
    <p id="apPwdErr" style="color:red;font-size:12px"></p>
    <button class="btn btn-green btn-block" onclick="saveAdminPassword()">修改密码</button>
  `);
}
async function saveAdminProfile() {
  const name=document.getElementById('apName').value.trim(), phone=document.getElementById('apPhone').value.trim(), address=document.getElementById('apAddr').value.trim();
  if(!name) { toast('姓名不能为空'); return; }
  await fetch(API+'/me',{method:'PUT',headers:headers(),body:JSON.stringify({name,address})});
  if(phone&&phone.length===11) await fetch(API+'/me/bindphone',{method:'POST',headers:headers(),body:JSON.stringify({phone})});
  closeModal(); toast('保存成功');
}
async function saveAdminPassword() {
  const o=document.getElementById('apOldPwd').value, n=document.getElementById('apNewPwd').value, n2=document.getElementById('apNewPwd2').value;
  if(n.length<6){document.getElementById('apPwdErr').textContent='新密码至少6位';return;}
  if(n!==n2){document.getElementById('apPwdErr').textContent='两次密码不一致';return;}
  const res=await fetch(API+'/me/password',{method:'POST',headers:headers(),body:JSON.stringify({old_password:o,new_password:n})});
  const data=await res.json();
  if(data.error){document.getElementById('apPwdErr').textContent=data.error;return;}
  closeModal(); toast('密码修改成功');
}

// === Households ===
async function loadHouseholdsAdmin() {
  const {data,total} = await (await fetch(API+'/admin/households?size='+PAGE_SIZE+'&page='+currentPage,{headers:headers()})).json();
  const sec = document.getElementById('section');
  let html = '';
  if(!data?.length) { sec.innerHTML='<div style="text-align:center;color:#999;padding:30px">暂无户籍</div>'; return; }
  html += data.map(h=>`
    <div class="list-item" onclick="showHouseholdDetail(${h.id})" style="cursor:pointer">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">户号：${h.household_no} · 户主：${h.head_name||'未设置'}</div>
        <span style="font-size:12px;color:#999">${h.member_count||0}人</span>
      </div>
      <div style="font-size:12px;color:#666;margin-top:4px">${h.group_name||''} · ${h.address} · 耕地${h.farmland_area}亩 · 林地${h.forest_area}亩 · 宅基地${h.homesite_area}㎡</div>
    </div>
  `).join('');
  html += paginator(currentPage, total);
  sec.innerHTML = html;
}

async function showHouseholdDetail(id) {
  const [{household:h, members}, memberData] = await Promise.all([
    (await fetch(API+'/admin/households/'+id,{headers:headers()})).json(),
    getUserOpts(0, u=>!u.household_id)
  ]);
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>户号：${h.household_no}</h3>
    <div style="font-size:13px;color:#666;margin:8px 0;line-height:2">
      <div>户主：${h.head_name||'未设置'}</div>
      <div>地址：${h.address}</div>
      <div>小组：${h.group_name||'未分配'}</div>
      <div>耕地：${h.farmland_area}亩</div>
      <div>林地：${h.forest_area}亩</div>
      <div>宅基地：${h.homesite_area}㎡</div>
      <div>备注：${h.remark||'无'}</div>
    </div>
    <div style="font-weight:600;margin:12px 0 8px">家庭成员 (${(members||[]).length}人)</div>
    ${(members||[]).map(m=>`
      <div style="display:flex;justify-content:space-between;align-items:center;padding:6px 0;border-bottom:1px solid #f0f0f0">
        <span>${m.user_name} · <a style="color:var(--blue);cursor:pointer" onclick="editMemberRelation(${id},${m.id},'${m.relation}')">${m.relation}</a></span>
        <a style="color:var(--red);font-size:12px;cursor:pointer" onclick="removeHouseholdMember(${id},${m.id})">移除</a>
      </div>
    `).join('')}
    <div style="margin-top:12px;display:flex;gap:8px">
      <input id="hhMemberInput" list="dlMember" placeholder="输入姓名搜索..." style="flex:1;padding:6px;border:1px solid #ddd;border-radius:6px;font-size:13px">
      <datalist id="dlMember">${memberData.opts}</datalist>
      <select id="hhMemberRel" style="padding:6px;border:1px solid #ddd;border-radius:6px">
        <option>户主</option><option>配偶</option><option>之子</option><option>之女</option><option>之父</option><option>之母</option><option>儿媳</option><option>女婿</option><option>之孙</option><option>之孙女</option><option>祖父</option><option>祖母</option><option>外祖父</option><option>外祖母</option><option>兄弟</option><option>姐妹</option><option>其他</option>
      </select>
      <button class="btn btn-green" style="padding:6px 12px;font-size:12px" onclick="addHouseholdMember(${id})">添加</button>
    </div>
    <div class="action-btns" style="margin-top:12px">
      <button class="btn btn-blue" onclick="editHousehold(${id})">编辑</button>
      <button class="btn btn-red" onclick="deleteHousehold(${id})">删除</button>
    </div>
  `);
}

async function addHouseholdMember(hhId) {
  const uid = getPickedId('hhMemberInput');
  const rel = document.getElementById('hhMemberRel').value;
  if(!uid) { toast('请选择用户'); return; }
  await fetch(API+'/admin/households/'+hhId+'/members',{method:'POST',headers:headers(),body:JSON.stringify({user_id:uid,relation:rel})});
  toast('已添加'); showHouseholdDetail(hhId);
}

function editMemberRelation(hhId, memberId, current) {
  const rels = ['户主','配偶','之子','之女','之父','之母','儿媳','女婿','之孙','之孙女','祖父','祖母','外祖父','外祖母','兄弟','姐妹','其他'];
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>修改关系</h3>
    <div class="form-group"><label>与户主关系</label><select id="editRel">${rels.map(r=>`<option ${r===current?'selected':''}>${r}</option>`).join('')}</select></div>
    <button class="btn btn-green btn-block" onclick="saveRelation(${hhId},${memberId})">保存</button>
  `);
}
function saveRelation(hhId, memberId) {
  const rel = document.getElementById('editRel').value;
  fetch(API+'/admin/households/'+hhId+'/members/'+memberId,{method:'PUT',headers:headers(),body:JSON.stringify({relation:rel})}).then(()=>{
    closeModal(); toast('已修改'); showHouseholdDetail(hhId);
  });
}

async function removeHouseholdMember(hhId, memberId) {
  if(!confirm('确定移除？')) return;
  await fetch(API+'/admin/households/'+hhId+'/members/'+memberId,{method:'DELETE',headers:headers()});
  toast('已移除'); showHouseholdDetail(hhId);
}

async function editHousehold(id) {
  const {household:h} = await (await fetch(API+'/admin/households/'+id,{headers:headers()})).json();
  const groups = (await (await fetch(API+'/groups')).json()).data||[];
  const groupOpts = '<option value="0">未分配</option>'+groups.map(g=>`<option value="${g.id}" ${h.group_id===g.id?'selected':''}>${g.name}</option>`).join('');
  const headData = await getUserOpts(h.head_id);
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>编辑户籍</h3>
    <div class="form-group"><label>户号</label><input id="hhNo" value="${h.household_no}"></div>
    <div class="form-group"><label>户主</label><input id="hhHeadInput" list="dlHead" value="${headData.selectedName}" placeholder="输入姓名搜索..." style="width:100%"><datalist id="dlHead">${headData.opts}</datalist></div>
    <div class="form-group"><label>地址</label><input id="hhAddr" value="${h.address||''}"></div>
    <div class="form-group"><label>小组</label><select id="hhGroup">${groupOpts}</select></div>
    <div class="form-group"><label>耕地面积(亩)</label><input id="hhFarm" type="number" step="0.1" value="${h.farmland_area}"></div>
    <div class="form-group"><label>林地面积(亩)</label><input id="hhForest" type="number" step="0.1" value="${h.forest_area}"></div>
    <div class="form-group"><label>宅基地面积(㎡)</label><input id="hhHomeSite" type="number" step="0.1" value="${h.homesite_area}"></div>
    <div class="form-group"><label>备注</label><input id="hhRemark" value="${h.remark||''}"></div>
    <button class="btn btn-green btn-block" onclick="saveHousehold(${id})">保存</button>
  `);
}

function showCreateHouseholdForm() {
  Promise.all([fetch(API+'/groups').then(r=>r.json()), getUserOpts(0, u=>!u.household_id)]).then(([gd, headData])=>{
    const groupOpts = '<option value="0">未分配</option>'+(gd.data||[]).map(g=>`<option value="${g.id}">${g.name}</option>`).join('');
    showModal(`
      <span class="close" onclick="closeModal()">&times;</span>
      <h3>创建户籍</h3>
      <div class="form-group"><label>户号</label><input id="hhNo" placeholder="如：001"></div>
      <div class="form-group"><label>户主</label><input id="hhHeadInput" list="dlHead" placeholder="输入姓名搜索..." style="width:100%"><datalist id="dlHead">${headData.opts}</datalist></div>
      <div class="form-group"><label>地址</label><input id="hhAddr"></div>
      <div class="form-group"><label>小组</label><select id="hhGroup">${groupOpts}</select></div>
      <div class="form-group"><label>耕地面积(亩)</label><input id="hhFarm" type="number" step="0.1" value="0"></div>
      <div class="form-group"><label>林地面积(亩)</label><input id="hhForest" type="number" step="0.1" value="0"></div>
      <div class="form-group"><label>宅基地面积(㎡)</label><input id="hhHomeSite" type="number" step="0.1" value="0"></div>
      <div class="form-group"><label>备注</label><input id="hhRemark"></div>
      <button class="btn btn-green btn-block" onclick="saveHousehold(0)">创建</button>
    `);
  });
}

async function saveHousehold(id) {
  const body = {
    household_no: document.getElementById('hhNo').value,
    head_id: getPickedId('hhHeadInput'),
    address: document.getElementById('hhAddr').value,
    group_id: parseInt(document.getElementById('hhGroup').value)||0,
    farmland_area: parseFloat(document.getElementById('hhFarm').value)||0,
    forest_area: parseFloat(document.getElementById('hhForest').value)||0,
    homesite_area: parseFloat(document.getElementById('hhHomeSite').value)||0,
    remark: document.getElementById('hhRemark').value,
  };
  if(!body.household_no) { toast('户号不能为空'); return; }
  if(id) await fetch(API+'/admin/households/'+id,{method:'PUT',headers:headers(),body:JSON.stringify(body)});
  else await fetch(API+'/admin/households',{method:'POST',headers:headers(),body:JSON.stringify(body)});
  closeModal(); toast('保存成功'); loadHouseholdsAdmin(); loadDashboard();
}

async function deleteHousehold(id) {
  if(!confirm('确定删除该户籍？所有成员关系将被清除')) return;
  await fetch(API+'/admin/households/'+id,{method:'DELETE',headers:headers()});
  closeModal(); toast('已删除'); loadHouseholdsAdmin(); loadDashboard();
}

// === Notifications ===
async function loadNotifyBadge() {
  try {
    const {count} = await (await fetch(API+'/notifications/unread-count',{headers:headers()})).json();
    document.getElementById('notifyBadge').textContent = count > 0 ? count : '';
  } catch(e) {}
}
setInterval(loadNotifyBadge, 30000);

async function showNotifications() {
  const {data,unread_count} = await (await fetch(API+'/notifications?size=30',{headers:headers()})).json();
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>消息通知 ${unread_count>0?'<span style="font-size:12px;color:var(--orange)">'+unread_count+'条未读</span>':''}</h3>
    ${unread_count>0?'<button class="btn btn-outline" style="margin-bottom:12px;padding:4px 12px;font-size:12px" onclick="markAllRead()">全部已读</button>':''}
    ${(data||[]).length===0?'<div style="text-align:center;color:#999;padding:30px">暂无消息</div>':''}
    ${(data||[]).map(n=>`
      <div style="padding:10px 0;border-bottom:1px solid #f0f0f0;${n.is_read?'opacity:.6':''}">
        <div style="font-size:13px;font-weight:${n.is_read?'400':'600'}">${esc(n.title)}</div>
        <div style="font-size:12px;color:#666;margin-top:4px">${n.content}</div>
        <div style="font-size:11px;color:#999;margin-top:4px">${n.created_at}${n.is_read?'':' · <a style="color:var(--green);cursor:pointer" onclick="markRead('+n.id+')">标为已读</a>'}</div>
      </div>
    `).join('')}
  `);
}
async function markRead(id) {
  await fetch(API+'/notifications/'+id+'/read',{method:'PUT',headers:headers()});
  showNotifications(); loadNotifyBadge();
}
async function markAllRead() {
  await fetch(API+'/notifications/read-all',{method:'POST',headers:headers()});
  showNotifications(); loadNotifyBadge();
}

// === Export buttons ===
function printModal() {
  const content = document.getElementById('modalBody').innerHTML;
  const win = window.open('','_blank');
  win.document.write('<html><head><title>打印</title><style>body{font-family:-apple-system,"PingFang SC",sans-serif;padding:20px;font-size:14px;line-height:1.8} .close,.action-btns,.btn,.form-group{display:none!important} .status-tag{display:inline-block;padding:2px 8px;border-radius:10px;font-size:12px;background:#f0f0f0} @media print{body{padding:0}}</style></head><body>'+content+'</body></html>');
  win.document.close();
  win.print();
}

// === Download with auth ===
async function getUserOpts(selectedId, filterFn) {
  const {data} = await (await fetch(API+'/admin/users?size=500',{headers:headers()})).json();
  const list = filterFn ? (data||[]).filter(filterFn) : (data||[]);
  // 返回 {options: datalist的option, selected: 选中用户名, list: 原始数据}
  const opts = list.map(u => `<option value="${u.name}" data-id="${u.id}">`).join('');
  const sel = list.find(u => u.id === selectedId);
  return { opts, list, selectedName: sel ? sel.name : '' };
}
// 从 datalist input 获取选中的 data-id
function getPickedId(inputId) {
  const input = document.getElementById(inputId);
  if(!input) return 0;
  const val = input.value.trim();
  if(!val) return 0;
  const dl = input.list;
  if(!dl) return 0;
  for(const opt of dl.options) {
    if(opt.value === val) return parseInt(opt.dataset.id)||0;
  }
  return 0;
}
function downloadFile(url) {
  fetch(url, {headers:headers()}).then(r=>{
    if(!r.ok) { toast('导出失败'); return; }
    const cd = r.headers.get('Content-Disposition')||'';
    const match = cd.match(/filename=(.+)/);
    const filename = match ? match[1] : 'export.xlsx';
    return r.blob().then(blob=>{
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = filename;
      a.click();
      URL.revokeObjectURL(a.href);
    });
  });
}

// === Reports ===
async function loadReports() {
  const list = await (await fetch(API+'/admin/reports',{headers:headers()})).json() || [];
  const sec = document.getElementById('section');
  let html = '<h3 style="font-size:16px;margin-bottom:12px">📊 报表中心</h3>';
  if(!list?.length) { sec.innerHTML=html+'<div style="text-align:center;color:#999;padding:30px">暂无报表</div>'; return; }
  html += list.map(r => {
    let paramDefs = []; try { paramDefs = JSON.parse(r.params||'[]'); } catch(e) {}
    const paramInputs = paramDefs.map(p => `<input id="rp_${r.name}_${p.name}" placeholder="${p.label}" value="${p.default||''}" style="padding:4px 8px;border:1px solid #ddd;border-radius:4px;font-size:12px;width:80px">`).join(' ');
    return `
    <div class="list-item">
      <div class="title">${r.label}</div>
      <div style="display:flex;gap:8px;align-items:center;margin-top:8px;flex-wrap:wrap">
        ${paramInputs}
        <button class="btn btn-green" style="padding:4px 12px;font-size:12px" onclick="runReport('${r.name}')">查询</button>
        <button class="btn btn-outline" style="padding:4px 12px;font-size:12px" onclick="printReport('${r.name}')">🖨️ 打印</button>
      </div>
      <div id="report_result_${r.name}" style="margin-top:8px"></div>
    </div>`;
  }).join('');
  sec.innerHTML = html;
}

async function runReport(name) {
  const container = document.getElementById('report_result_'+name);
  container.innerHTML = '<div style="color:#999;font-size:12px">查询中...</div>';
  // 收集参数
  const params = new URLSearchParams();
  document.querySelectorAll(`[id^="rp_${name}_"]`).forEach(el => {
    const pName = el.id.replace(`rp_${name}_`,'');
    params.set(pName, el.value);
  });
  try {
    const {result} = await (await fetch(API+'/admin/reports/'+name+'?'+params,{headers:headers()})).json();
    if(!result?.columns?.length) { container.innerHTML='<div style="color:#999;font-size:12px">无数据</div>'; return; }
    let html = '<div style="overflow-x:auto"><table style="width:100%;border-collapse:collapse;font-size:12px;margin-top:4px">';
    html += '<tr>'+result.columns.map(c=>'<th style="border:1px solid #e0e0e0;padding:4px 8px;background:#f5f5f5;white-space:nowrap">'+c+'</th>').join('')+'</tr>';
    (result.rows||[]).forEach(row => {
      html += '<tr>'+row.map(v=>'<td style="border:1px solid #e0e0e0;padding:4px 8px;white-space:nowrap">'+(v??'')+'</td>').join('')+'</tr>';
    });
    html += '</table></div>';
    container.innerHTML = html;
  } catch(e) { container.innerHTML='<div style="color:var(--red);font-size:12px">查询失败</div>'; }
}

function printReport(name) {
  const params = new URLSearchParams();
  document.querySelectorAll(`[id^="rp_${name}_"]`).forEach(el => {
    params.set(el.id.replace(`rp_${name}_`,''), el.value);
  });
  openPrintPage(API+'/admin/print/report/'+name+'?'+params);
}

// 带认证的打印页打开
function openPrintPage(url) {
  fetch(url, {headers:headers()}).then(r=>{
    if(!r.ok) { toast('打印页加载失败'); return; }
    return r.text();
  }).then(html=>{
    if(!html) return;
    const win = window.open('','_blank');
    if(!win) { toast('请允许弹出窗口'); return; }
    win.document.write(html);
    win.document.close();
  }).catch(()=>toast('打印页加载失败'));
}

// === Workflows ===
async function loadWorkflows() {
  const list = await (await fetch(API+'/admin/workflow-defs',{headers:headers()})).json() || [];
  const sec = document.getElementById('section');
  const docTypeMap = {notice:'公告',finance:'财务',subsidy:'补贴',ticket:'工单'};
  let html = '<h3 style="font-size:16px;margin-bottom:12px">⚙️ 审批流程</h3>';
  if(!list?.length) { sec.innerHTML=html+'<div style="text-align:center;color:#999;padding:30px">暂无流程</div>'; return; }
  html += list.map(d => {
    const stateLabels = (d.states||[]).map(s => `<span class="status-tag ${s.name}" style="margin:1px">${s.label}</span>`).join(' → ');
    const transHtml = (d.transitions||[]).map(t => `<div style="font-size:11px;color:#666;padding:2px 0">${t.label}：${(d.states.find(s=>s.name===t.from)||{}).label||t.from} → ${(d.states.find(s=>s.name===t.to)||{}).label||t.to}（${roleMap[t.min_role]||t.min_role}以上）</div>`).join('');
    return `
    <div class="list-item">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div class="title">${d.label}</div>
        <span style="font-size:11px;color:${d.active?'var(--green)':'#999'}">${d.active?'启用':'停用'}</span>
      </div>
      <div style="font-size:12px;color:#666;margin-top:4px">适用：${docTypeMap[d.doc_type]||d.doc_type}</div>
      <div style="margin-top:6px;display:flex;flex-wrap:wrap;gap:2px;align-items:center">${stateLabels}</div>
      <div style="margin-top:6px">${transHtml}</div>
    </div>`;
  }).join('');
  sec.innerHTML = html;
}

// === Import Users ===
function showImportUsers() {
  showModal(`
    <span class="close" onclick="closeModal()">&times;</span>
    <h3>📤 批量导入村民</h3>
    <div style="font-size:13px;color:#666;margin-bottom:16px;line-height:1.8">
      <p>1. 先下载模板，按格式填写村民信息</p>
      <p>2. 上传 Excel 文件，系统自动导入</p>
      <p>3. 默认密码为 123456，角色为村民</p>
    </div>
    <div style="display:flex;gap:12px;margin-bottom:16px">
      <button class="btn btn-outline" onclick="downloadFile(API+'/admin/import/template')">📥 下载模板</button>
    </div>
    <div class="form-group">
      <label>选择 Excel 文件</label>
      <input type="file" id="importFile" accept=".xlsx,.xls" style="padding:8px">
    </div>
    <div id="importResult" style="margin-bottom:12px"></div>
    <button class="btn btn-green btn-block" onclick="submitImport()">开始导入</button>
  `);
}

async function submitImport() {
  const file = document.getElementById('importFile').files[0];
  if(!file) { toast('请选择文件'); return; }
  const resultDiv = document.getElementById('importResult');
  resultDiv.innerHTML = '<div style="color:var(--blue);font-size:13px">导入中...</div>';
  const fd = new FormData();
  fd.append('file', file);
  try {
    const res = await fetch(API+'/admin/import/users',{method:'POST',headers:{'Authorization':'Bearer '+token},body:fd});
    const data = await res.json();
    if(data.error) { resultDiv.innerHTML = `<div style="color:var(--red);font-size:13px">${data.error}</div>`; return; }
    let html = `<div style="font-size:13px;padding:12px;background:#f5f5f5;border-radius:8px">
      <div>总计：${data.total} 条</div>
      <div style="color:var(--green)">成功：${data.success} 条</div>
      ${data.failed?'<div style="color:var(--red)">失败：'+data.failed+' 条</div>':''}
    </div>`;
    if(data.errors?.length) {
      html += '<div style="font-size:12px;color:var(--red);margin-top:8px;max-height:150px;overflow-y:auto">'+data.errors.map(e=>'<div>'+e+'</div>').join('')+'</div>';
    }
    resultDiv.innerHTML = html;
    if(data.success > 0) { loadDashboard(); }
  } catch(e) { resultDiv.innerHTML = '<div style="color:var(--red);font-size:13px">导入失败</div>'; }
}

// ==================== 操作日志 ====================
let logsPage = 1;
let logsDocType = '';
function loadWorkflowLogs(page) {
  if (page) logsPage = page;
  const typeMap = {notice:'公告',finance:'财务',subsidy:'补贴',ticket:'工单'};
  let html = '<h3 style="font-size:16px;margin-bottom:12px">📋 操作日志</h3>';
  html += '<div style="margin-bottom:12px"><select id="logTypeFilter" onchange="logsDocType=this.value;loadWorkflowLogs(1)" style="padding:6px 12px;border:1px solid #ddd;border-radius:6px;font-size:13px">';
  html += '<option value="">全部类型</option>';
  for (const [k,v] of Object.entries(typeMap)) html += `<option value="${k}" ${logsDocType===k?'selected':''}>${v}</option>`;
  html += '</select></div>';
  html += '<div id="logsBody">加载中...</div>';
  document.getElementById('section').innerHTML = html;

  fetch(API+`/admin/workflow-logs?page=${logsPage}&size=20&doc_type=${logsDocType}`,{headers:headers()}).then(r=>r.json()).then(res => {
    if (!res.data || res.data.length === 0) {
      document.getElementById('logsBody').innerHTML = '<div style="color:#999;padding:20px;text-align:center">暂无日志</div>';
      return;
    }
    let t = '<table style="width:100%;border-collapse:collapse;font-size:13px"><tr style="background:#f5f5f5">';
    t += '<th style="padding:8px;text-align:left">时间</th><th style="padding:8px;text-align:left">类型</th><th style="padding:8px;text-align:left">内容</th><th style="padding:8px;text-align:left">操作</th><th style="padding:8px;text-align:left">操作人</th><th style="padding:8px;text-align:left">状态变更</th><th style="padding:8px;text-align:left">备注</th></tr>';
    res.data.forEach(l => {
      const from = stateMap[l.from_state] || l.from_state || '-';
      const to = stateMap[l.to_state] || l.to_state;
      const title = l.doc_title || '#'+l.doc_id;
      t += `<tr style="border-bottom:1px solid #eee">`;
      t += `<td style="padding:8px">${l.created_at.replace('T',' ').substring(0,19)}</td>`;
      t += `<td style="padding:8px">${typeMap[l.doc_type]||l.doc_type}</td>`;
      t += `<td style="padding:8px;color:var(--green);cursor:pointer;text-decoration:underline" onclick='showLogDetail(${JSON.stringify(l).replace(/'/g,"&#39;")})'>${title}</td>`;
      t += `<td style="padding:8px">${l.action}</td>`;
      t += `<td style="padding:8px">${l.operator_name}</td>`;
      t += `<td style="padding:8px"><span style="color:#999">${from}</span> → <span style="color:var(--green)">${to}</span></td>`;
      t += `<td style="padding:8px;color:#666">${l.note||'-'}</td></tr>`;
    });
    t += '</table>';
    const totalPages = Math.ceil(res.total / 20);
    if (totalPages > 1) {
      t += '<div style="margin-top:12px;text-align:center">';
      if (logsPage > 1) t += `<button onclick="loadWorkflowLogs(${logsPage-1})" class="btn btn-outline" style="margin:0 4px">上一页</button>`;
      t += `<span style="font-size:13px;color:#666;margin:0 8px">${logsPage}/${totalPages}</span>`;
      if (logsPage < totalPages) t += `<button onclick="loadWorkflowLogs(${logsPage+1})" class="btn btn-outline" style="margin:0 4px">下一页</button>`;
      t += '</div>';
    }
    document.getElementById('logsBody').innerHTML = t;
  });
}

function showLogDetail(l) {
  const typeMap = {notice:'公告',finance:'财务',subsidy:'补贴',ticket:'工单'};
  const from = stateMap[l.from_state] || l.from_state || '-';
  const to = stateMap[l.to_state] || l.to_state;
  const title = l.doc_title || '#'+l.doc_id;
  let html = `<div style="position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.4);z-index:999;display:flex;align-items:center;justify-content:center" onclick="if(event.target===this)this.remove()">
    <div style="background:#fff;border-radius:12px;padding:24px;max-width:480px;width:90%;box-shadow:0 8px 32px rgba(0,0,0,.15)">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
        <h3 style="font-size:16px;margin:0">操作详情</h3>
        <span style="cursor:pointer;font-size:20px;color:#999" onclick="this.closest('[style*=fixed]').remove()">✕</span>
      </div>
      <table style="width:100%;font-size:14px;line-height:2">
        <tr><td style="color:#999;width:80px">类型</td><td>${typeMap[l.doc_type]||l.doc_type}</td></tr>
        <tr><td style="color:#999">内容</td><td style="font-weight:600">${title}</td></tr>
        <tr><td style="color:#999">操作</td><td>${l.action}</td></tr>
        <tr><td style="color:#999">操作人</td><td>${l.operator_name}</td></tr>
        <tr><td style="color:#999">状态变更</td><td>${from} → <span style="color:var(--green)">${to}</span></td></tr>
        <tr><td style="color:#999">备注</td><td>${l.note||'无'}</td></tr>
        <tr><td style="color:#999">时间</td><td>${l.created_at.replace('T',' ').substring(0,19)}</td></tr>
      </table>
      <div style="margin-top:16px;text-align:right">
        <button class="btn btn-primary" style="padding:8px 24px" onclick="this.closest('[style*=fixed]').remove()">关闭</button>
      </div>
    </div>
  </div>`;
  document.body.insertAdjacentHTML('beforeend', html);
}
