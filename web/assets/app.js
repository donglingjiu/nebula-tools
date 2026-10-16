/* Nova Tools v2.2.1 — design-package UI × real API
   已合并补丁：arrowUp/shieldCheck 图标 / data-cat 分类筛选 / applyToolFilter /
   data-filter 委托 / 角色改用 change 事件 / 工作流名称保持 / 登出清态 /
   loadHomeRecent 同步 / 下拉定位兜底 / data-page 对齐设计包 token */
(() => {
'use strict';
const $=(s,r=document)=>r.querySelector(s), $$=(s,r=document)=>[...r.querySelectorAll(s)];
const esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const fmtT=t=>new Date(t).toLocaleString('zh-CN',{hour12:false});
const api=async(p,opt={})=>{const r=await fetch(p,{headers:{'Content-Type':'application/json'},...opt});const j=await r.json().catch(()=>({}));if(!r.ok||j.error)throw new Error(j.error||('HTTP '+r.status));return j};
const post=(p,d)=>api(p,{method:'POST',body:JSON.stringify(d||{})});

/* ---------- 图标（取自设计包） ---------- */
const paths={
planet:'<circle cx="12" cy="12" r="6.8"/><ellipse cx="12" cy="12" rx="10.5" ry="3.4" transform="rotate(-28 12 12)"/><path d="M8 7.5c1.8-1.8 4.6-2.5 7-1.8"/>',
sparkles:'<path d="m12 3 1.5 5.5L19 10l-5.5 1.5L12 17l-1.5-5.5L5 10l5.5-1.5L12 3Z"/><path d="m19 16 .8 2.2L22 19l-2.2.8L19 22l-.8-2.2L16 19l2.2-.8L19 16Z"/>',
search:'<circle cx="10.8" cy="10.8" r="6.8"/><path d="m16 16 4.5 4.5"/>',
bell:'<path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"/><path d="M10 21h4"/>',
user:'<circle cx="12" cy="8" r="3.5"/><path d="M4.5 20c.4-4 3.2-6 7.5-6s7.1 2 7.5 6"/>',
home:'<path d="m3 10 9-7 9 7"/><path d="M5.5 9v11h13V9M9.5 20v-6h5v6"/>',
grid:'<rect x="3.5" y="3.5" width="7" height="7" rx="1.5"/><rect x="13.5" y="3.5" width="7" height="7" rx="1.5"/><rect x="3.5" y="13.5" width="7" height="7" rx="1.5"/><rect x="13.5" y="13.5" width="7" height="7" rx="1.5"/>',
files:'<rect x="7" y="4" width="13" height="16" rx="2"/><path d="M4 7v13a2 2 0 0 0 2 2h11M10 8h6M10 12h6M10 16h4"/>',
fileText:'<path d="M13 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V10Z"/><path d="M13 3v7h7M8 14h8M8 17h6"/>',
chart:'<path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>',
code:'<path d="m8 7-5 5 5 5M16 7l5 5-5 5M14 4l-4 16"/>',
brain:'<path d="M12 4a3 3 0 0 0-5.7 1.3A3.5 3.5 0 0 0 4 11.5a3.5 3.5 0 0 0 2 6.2A3.5 3.5 0 0 0 12 19Z"/><path d="M12 4a3 3 0 0 1 5.7 1.3A3.5 3.5 0 0 1 20 11.5a3.5 3.5 0 0 1-2 6.2A3.5 3.5 0 0 1 12 19Z"/>',
upload:'<path d="M12 16V4M7 9l5-5 5 5"/><path d="M4 16v3a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-3"/>',
download:'<path d="M12 3v12m-5-5 5 5 5-5"/><path d="M4 17v3h16v-3"/>',
arrow:'<path d="M4 12h15M13 5l7 7-7 7"/>',
arrowUp:'<path d="M12 19V5M5 12l7-7 7 7"/>',
chevron:'<path d="m9 18 6-6-6-6"/>',
plus:'<path d="M12 5v14M5 12h14"/>',
close:'<path d="m6 6 12 12M18 6 6 18"/>',
check:'<path d="m5 12 4 4L19 6"/>',
clock:'<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
shield:'<path d="M12 22s8-4 8-11V5l-8-3-8 3v6c0 7 8 11 8 11Z"/><path d="m9 12 2 2 4-4"/>',
shieldCheck:'<path d="M12 22s8-4 8-11V5l-8-3-8 3v6c0 7 8 11 8 11Z"/><path d="m8.5 11.5 2.5 2.5 4.5-5"/>',
settings:'<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1.08-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1.08 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h.08a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h.08a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v.08a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1Z"/>',
workflow:'<rect x="3" y="3" width="6" height="6" rx="1.5"/><rect x="15" y="15" width="6" height="6" rx="1.5"/><rect x="15" y="3" width="6" height="6" rx="1.5"/><path d="M9 6h6M18 9v6M9 6v12h6"/>',
message:'<path d="M20 11.5a7.5 7.5 0 0 1-8 7.5 8 8 0 0 1-3.2-.7L4 20l1.4-4A7 7 0 0 1 4 11.5 7.5 7.5 0 0 1 12 4a7.5 7.5 0 0 1 8 7.5Z"/><path d="M8 11h8M8 14h5"/>',
book:'<path d="M4 5.5A2.5 2.5 0 0 1 6.5 3H20v17H6.5A2.5 2.5 0 0 1 4 17.5Z"/><path d="M4 17.5A2.5 2.5 0 0 1 6.5 15H20M8 7h7M8 10h5"/>',
zap:'<path d="m13 2-3 9h7L9 22l2-9H4l9-11Z"/>',
globe:'<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18"/>',
help:'<circle cx="12" cy="12" r="9"/><path d="M9.5 9a2.6 2.6 0 1 1 4.3 2c-1.1.8-1.8 1.2-1.8 2.8M12 17h.01"/>',
logout:'<path d="M10 17l5-5-5-5M15 12H3M12 3h7a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-7"/>',
menu:'<path d="M4 6h16M4 12h16M4 18h16"/>',
copy:'<rect x="8" y="8" width="12" height="12" rx="2"/><path d="M16 8V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h3"/>',
star:'<path d="m12 3 2.7 5.5 6 .9-4.4 4.3 1 6-5.3-2.8-5.3 2.8 1-6-4.4-4.3 6-.9L12 3Z"/>',
more:'<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>',
refresh:'<path d="M20 7v5h-5M4 17v-5h5"/><path d="M5.5 9A7 7 0 0 1 17 5l3 2M4 17l3 2a7 7 0 0 0 11.5-4"/>',
trash:'<path d="M3 6h18M8 6V4h8v2M6 6l1 15h10l1-15M10 10v7M14 10v7"/>',
users:'<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM22 21v-2a4 4 0 0 0-3-3.87"/>',
image:'<rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/>',
terminal:'<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m7 9 3 3-3 3M13 15h4"/>',
database:'<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>',
key:'<circle cx="8" cy="15" r="5"/><path d="m12 12 8-8 2 2-2 2 2 2-2 2-2-2-3 3"/>',
lock:'<rect x="4" y="10" width="16" height="11" rx="2"/><path d="M8 10V7a4 4 0 1 1 8 0v3"/>',
eye:'<path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6-10-6-10-6Z"/><circle cx="12" cy="12" r="2.5"/>'};
const icon=(n,s=18)=>`<svg width="${s}" height="${s}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${paths[n]||paths.sparkles}</svg>`;

/* ---------- 状态 ---------- */
const S={me:null,tools:[],notices:[],health:{},q:'',catFilter:'all',curTool:null,aiMsgs:[],wfSteps:[],wfName:'',favs:[],hist:[]};
const CATS={dev:'开发工具',file:'文件处理',gen:'生成工具',data:'数据转换'};
const TOOL_ICON={'dev.json':'code','dev.md':'fileText','dev.encode':'lock','dev.hash':'shield','dev.sql':'database','dev.util':'clock','gen.qr':'grid','gen.base':'terminal','gen.pwd':'key','dev.diff':'copy','data.csv2json':'files','data.json2csv':'files','file.pdf':'fileText','file.excel':'chart','file.word':'files','file.img':'image'};
const COLORS=['','mint','violet','peach','gold'];
const tIcon=t=>TOOL_ICON[t.id]||'sparkles';
const tName=id=>{const t=S.tools.find(x=>x.id===id);return t?t.name:id};

const OPTS={
'dev.json':[{k:'mode',label:'处理模式',choices:[['format','格式化','排版美化，便于阅读','code'],['minify','压缩','移除空白，减小体积','zap'],['validate','校验','检查 JSON 是否合法','check']]}],
'dev.encode':[{k:'op',label:'转换方向',choices:[['b64enc','Base64 编码','文本 → Base64','lock'],['b64dec','Base64 解码','Base64 → 文本','key'],['uenc','URL 编码','Query 转义','globe'],['udec','URL 解码','还原转义','globe']]}],
'dev.util':[{k:'op',label:'功能',choices:[['now','时间现在','当前时间与 UUID','clock'],['uuid','批量 UUID','生成 5 个','copy'],['ts2date','时间戳 → 日期','支持秒/毫秒','clock'],['date2ts','日期 → 时间戳','毫秒输出','clock']]}],
'gen.base':[{k:'base',label:'输入进制',choices:[['10','十进制','默认','terminal'],['2','二进制','0/1','terminal'],['8','八进制','0-7','terminal'],['16','十六进制','0-F','terminal']]}],
'gen.pwd':[{k:'len',label:'长度',choices:[['12','12 位','推荐','key'],['8','8 位','简洁','key'],['16','16 位','强','key'],['24','24 位','极强','key']]},
{k:'count',label:'数量',choices:[['5','5 个','默认','copy'],['1','1 个','单个','copy'],['10','10 个','批量','copy']]}],
'file.img':[{k:'op',label:'操作',choices:[['compress','压缩','转 JPG 减小体积','image'],['to_jpg','转 JPG','兼容性最好','image'],['to_png','转 PNG','无损透明','image']]},
{k:'quality',label:'JPG 质量',choices:[['80','80%','推荐','zap'],['60','60%','高压缩','zap'],['92','92%','高清','zap']]},
{k:'maxw',label:'最大宽度',choices:[['','保持原尺寸','不缩放','eye'],['1280','1280px','网页常用','eye'],['800','800px','更小','eye']]}]};
OPTS['gen.qr']=[{k:'size',label:'尺寸',choices:[['512','512px','推荐','grid'],['256','256px','小','grid'],['1024','1024px','高清打印','grid']]}];

/* ---------- Toast / Modal / 命令面板 ---------- */
function toast(msg,kind='success'){let st=$('.toast-stack');if(!st){st=document.createElement('div');st.className='toast-stack';document.body.appendChild(st)}const el=document.createElement('div');el.className='toast';el.innerHTML=`<span class="toast-icon" ${kind==='error'?'style="background:#fff0ea;color:#c05f5f"':''}>${icon(kind==='error'?'close':'check',14)}</span><span>${esc(msg)}</span>`;st.appendChild(el);setTimeout(()=>{el.style.opacity='0';el.style.transform='translateY(8px)';setTimeout(()=>el.remove(),220)},3000)}
function openModal(title,desc,body,wide=false){let ov=$('#globalModal');if(!ov){ov=document.createElement('div');ov.id='globalModal';ov.className='modal-backdrop';document.body.appendChild(ov);ov.addEventListener('click',e=>{if(e.target===ov)closeModal()})}ov.innerHTML=`<section class="modal ${wide?'wide':''}" role="dialog" aria-modal="true"><div class="modal-head"><div><h2>${esc(title)}</h2><p>${desc||''}</p></div><button class="modal-close" data-close aria-label="关闭">${icon('close',16)}</button></div><div class="modal-content">${body}</div></section>`;requestAnimationFrame(()=>ov.classList.add('open'));document.body.style.overflow='hidden'}
function closeModal(){const ov=$('#globalModal');if(!ov)return;ov.classList.remove('open');document.body.style.overflow='';setTimeout(()=>ov.innerHTML='',220)}
function commandPalette(){let ov=$('#commandPalette');if(!ov){ov=document.createElement('div');ov.id='commandPalette';ov.className='modal-backdrop';ov.innerHTML=`<section class="modal command-modal" role="dialog" aria-modal="true"><div class="command-search">${icon('search',18)}<input id="commandInput" placeholder="搜索页面、工具或操作…"><span class="kbd">ESC</span></div><div class="command-items" id="commandItems"></div><div class="command-footer"><span>↑↓ 选择</span><span>Enter 打开</span><span>Esc 关闭</span><span style="margin-left:auto">Nova Quick Switch</span></div></section>`;document.body.appendChild(ov);ov.addEventListener('click',e=>{if(e.target===ov)closePalette()});$('#commandInput',ov).addEventListener('input',renderPalette);$('#commandInput',ov).addEventListener('keydown',e=>{const items=$$('.command-item',ov);let i=items.findIndex(x=>x.classList.contains('active'));if(e.key==='ArrowDown'){e.preventDefault();items.forEach(x=>x.classList.remove('active'));items[(i+1)%items.length]?.classList.add('active')}if(e.key==='ArrowUp'){e.preventDefault();items.forEach(x=>x.classList.remove('active'));items[(i-1+items.length)%items.length]?.classList.add('active')}if(e.key==='Enter'){e.preventDefault();items.find(x=>x.classList.contains('active'))?.click()}})}renderPalette();requestAnimationFrame(()=>ov.classList.add('open'));document.body.style.overflow='hidden';setTimeout(()=>$('#commandInput',ov)?.focus(),50)}
function closePalette(){const ov=$('#commandPalette');if(ov){ov.classList.remove('open');document.body.style.overflow='';setTimeout(()=>ov.remove(),220)}}
function renderPalette(){const root=$('#commandItems');if(!root)return;const q=($('#commandInput')?.value||'').toLowerCase().trim();const routes=[['首页','#/','home','AI 工作空间总览'],['工具大全','#/tools','grid','探索全部工具'],['AI 工作台','#/ai','sparkles','智能对话与任务'],['文件工作台','#/workspace','files','历史与收藏'],['数据洞察','#/analytics','chart','用量统计'],['智能工作流','#/workflows','workflow','自动化流水线'],['个人中心','#/profile','user','账户与安全'],['帮助中心','#/help','help','指南与常见问题']];const items=[...routes.map(r=>({label:r[0],href:r[1],ic:r[2],desc:r[3]})),...(S.me&&S.me.role==='admin'?[{label:'管理后台',href:'#/admin/overview',ic:'shield',desc:'用户、工具与系统配置'}]:[]),...S.tools.filter(t=>t.on).map(t=>({label:t.name,href:'#/tool/'+t.id,ic:tIcon(t),desc:t.desc}))].filter(x=>(x.label+' '+x.desc).toLowerCase().includes(q));root.innerHTML=items.slice(0,12).map((x,i)=>`<a class="command-item ${i===0?'active':''}" href="${x.href}"><span class="tool-icon">${icon(x.ic,16)}</span><span><strong>${esc(x.label)}</strong><small>${esc(x.desc)}</small></span><kbd class="kbd">↵</kbd></a>`).join('')||'<div class="empty-search">没有找到匹配的页面或工具。</div>'}

/* ---------- 壳层 ---------- */
const logo=()=>`<a class="brand" href="#/" aria-label="Nova Tools 首页"><span class="brand-symbol">${icon('planet',22)}</span><span>Nova <span style="font-weight:500">Tools</span></span></a>`;
function unreadCount(){const last=localStorage.getItem('nova_bell')||'';return S.notices.filter(n=>fmtT(n.at)>last).length}
function profileDropdown(){return `<div style="position:relative"><button class="icon-btn" data-dropdown="profileMenu" aria-label="个人菜单"><span class="avatar">${S.me?esc(S.me.username[0].toUpperCase()):'N'}</span></button><div class="dropdown" id="profileMenu"><div class="dropdown-head"><strong>你好，${S.me?esc(S.me.username):'访客'}</strong><span>${S.me?esc(S.me.role==='admin'?'管理员':'个人账户'):'未登录'}</span></div><a href="#/profile">${icon('user')}个人中心</a><a href="#/workspace">${icon('files')}历史与收藏</a>${S.me&&S.me.role==='admin'?`<a href="#/admin/overview">${icon('shield')}管理后台</a>`:''}<div class="sidebar-divider"></div>${S.me?`<button data-logout class="dropdown-danger">${icon('logout')}退出登录</button>`:`<a href="#/login">${icon('lock')}登录 / 注册</a>`}</div></div>`}
function frontHeader(active){const un=unreadCount();const nav=[['#/tools','工具大全','tools'],['#/ai','AI 工作台','ai'],['#/workflows','智能工作流','workflows'],['#/analytics','数据洞察','analytics'],['#/workspace','文件工作台','workspace'],['#/help','帮助中心','help']];return `<header class="topbar">${logo()}<button class="icon-btn mobile-menu" data-mobile-menu aria-label="导航">${icon('menu')}</button><nav class="nav">${nav.map(([h,l,k])=>`<a href="${h}" class="${active===k?'active':''}">${l}</a>`).join('')}</nav><div class="top-actions"><button class="icon-btn hide-mobile" data-command-open aria-label="搜索">${icon('search')}</button><button class="icon-btn" data-open-notices aria-label="通知">${icon('bell')}${un?`<span class="notification-dot"></span>`:''}</button>${profileDropdown()}</div></header>`}
const footer=()=>`<footer class="page-footer"><span>${logo()}<span style="margin-left:12px">AI 让工作更简单 · 零依赖单二进制 · ${esc(S.health.db_driver||'json')} 存储 · ${esc(S.health.version||'v2.2.1')}</span></span><div class="footer-links"><a href="#/help">帮助中心</a><a href="#/tools">工具大全</a>${S.me&&S.me.role==='admin'?'<a href="#/admin/overview">管理后台</a>':''}<span>运行 ${S.health.uptime_sec||0}s</span></div></footer>`;
function render(html,active,page){document.body.dataset.page=page||active;document.body.classList.remove('is-admin');$('#app').innerHTML=frontHeader(active)+`<div class="page page-transition">${html}</div>`+footer()}

/* ---------- 通用组件 ---------- */
function metric(label,value,sub='',ic='chart',color=''){return `<article class="stat-card"><div class="stat-top"><span>${label}</span><span class="stat-icon ${color}">${icon(ic,16)}</span></div><div class="stat-value">${esc(value)}</div><div class="stat-foot">${sub?esc(sub):'&nbsp;'}</div></article>`}
function toolCard(t,i){const badge={'file.pdf':'热门','gen.qr':'新','dev.json':'高频'}[t.id]||'';return `<a class="tool-card" href="#/tool/${t.id}" data-cat="${t.cat}"><div class="tool-card-top"><span class="tool-icon ${COLORS[i%5]}">${icon(tIcon(t),21)}</span><span class="card-arrow">${icon('arrow',16)}</span></div><h3>${esc(t.name)}${badge?` <span class="tag info">${badge}</span>`:''}</h3><p>${esc(t.desc)}</p></a>`}
function bars(map,labelFn){const e=Object.entries(map).sort((a,b)=>b[1]-a[1]).slice(0,12);const max=Math.max(1,...e.map(x=>x[1]));return e.length?e.map(([k,v])=>`<div class="bar-row"><span class="lb">${esc(labelFn(k))}</span><div class="bar"><i style="width:${(v/max)*100}%"></i></div><span class="v">${v}</span></div>`).join(''):'<div class="empty-state" style="padding:24px"><strong>暂无数据</strong><p>运行几个工具后这里会出现统计。</p></div>'}
function showOut(el,txt,id){el.dataset.raw=txt;if(txt.startsWith('data:image')){el.innerHTML=`<img src="${txt}" alt="结果图片"><div style="margin-top:10px"><a class="secondary btn-sm" download="nova-result" href="${txt}">${icon('download',13)} 下载图片</a></div>`;return}if(id==='dev.diff'){el.innerHTML=txt.split('\n').map(l=>l.startsWith('+')?`<div style="color:#189e83">${esc(l)}</div>`:l.startsWith('-')?`<div style="color:#d66d5c">${esc(l)}</div>`:`<div>${esc(l)}</div>`).join('');return}el.textContent=txt}
function applyToolFilter(){const q=(($('#toolSearch')?.value)||'').toLowerCase().trim();const c=S.catFilter||'all';let n=0;$$('.tool-library .tool-card').forEach(card=>{const hit=(c==='all'||card.dataset.cat===c)&&(!q||card.textContent.toLowerCase().includes(q));card.hidden=!hit;if(hit)n++});const tc=$('#toolsCount');if(tc)tc.textContent=n;const te=$('#toolsEmpty');if(te)te.hidden=n>0}

/* ---------- 页面 ---------- */
function pageTitle(eyebrow,title,sub,actions=''){return `<div class="page-title"><div><div class="eyebrow">${eyebrow}</div><h1>${title}</h1><p>${sub}</p></div><div class="title-actions">${actions}</div></div>`}
function homePage(){const on=S.tools.filter(t=>t.on);const pop=['dev.json','file.pdf','gen.qr','dev.encode'].map(id=>on.find(t=>t.id===id)).filter(Boolean);return `
<section class="hero"><div class="hero-copy"><div class="eyebrow">AI POWERED PRODUCTIVITY · NOVA OS</div><h1>让复杂工作<br><span class="h1-gradient">变得简单而聪明</span></h1><p class="lead">16 个即开即用的工具、智能工作流与 AI 助手——数据不出你的机器，专注留给真正的创造。</p><form class="searchbar" data-hero-search><span>${icon('sparkles',18)}</span><input id="globalSearch" autocomplete="off" placeholder="搜索工具，例如：二维码、JSON、PDF…"><button class="primary search-submit" type="submit" aria-label="搜索">${icon('arrow',18)}</button><span class="kbd hide-mobile">↵</span></form><div class="hero-tags"><span class="small" style="color:#9eafd3">试试</span>${['二维码生成','JSON 格式化','PDF 提取','图片压缩'].map(x=>`<button class="chip" data-quick="${x}">${x}</button>`).join('')}</div><div class="hero-meta"><span class="meta-item">${icon('zap')} 快速启动</span><span class="meta-item">${icon('shieldCheck')} 本地处理 · 隐私优先</span><span class="meta-item">${icon('workflow')} 一站式工作流</span></div></div><div class="hero-visual"><div class="cosmos"><div class="orbit"></div><div class="orbit orbit-2"></div><div class="orbit orbit-3"></div><div class="cosmos-core"></div><span class="float-tool ft-a">${icon('fileText')}</span><span class="float-tool ft-b">${icon('chart')}</span><span class="float-tool ft-c">${icon('code')}</span><span class="float-tool ft-d">${icon('image')}</span><span class="float-tool ft-e">${icon('brain')}</span><span class="orbit-label">NOVA INTELLIGENCE CORE</span><div class="hero-mini-card hero-mini-1"><i class="mini-pulse"></i><span><strong>${on.length} 个工具已就绪</strong><small>点击即用 · 无需安装</small></span></div><div class="hero-mini-card hero-mini-2">${icon('sparkles',16)}<span><strong>AI 助手随时待命</strong><small>配置端点即可对话</small></span></div></div></div></section>
<section class="section"><div class="section-head"><div><div class="eyebrow">QUICK ACCESS</div><h2>从一个小任务开始</h2><p>精选高频工具，工作流自然衔接。</p></div><a class="text-link" href="#/tools">浏览全部 ${on.length} 个工具 ${icon('arrow',14)}</a></div><div class="quick-row">${[['PDF / Word / Excel','fileText','#/workspace','文档处理',''],['趋势、图表与结论','chart','#/analytics','数据洞察','mint'],['智能对话','brain','#/ai','AI 助手','violet'],['二维码 / 密码','grid','#/tool/gen.qr','生成工具','peach'],['JSON / 编码 / 哈希','code','#/tool/dev.json','开发工具',''],['多步骤自动化','workflow','#/workflows','工作流','violet']].map(x=>`<a href="${x[2]}" class="quick-link"><span class="quick-icon ${x[4]}">${icon(x[1],19)}</span><span><strong>${x[3]}</strong><small>${x[0]}</small></span></a>`).join('')}</div></section>
<section class="section"><div class="section-head"><div><div class="eyebrow">POPULAR TOOLS</div><h2>为专注而设计的工具</h2><p>功能足够强大，界面始终保持清晰。</p></div><a class="text-link" href="#/tools">查看全部 ${icon('arrow',14)}</a></div><div class="tool-grid">${pop.map(toolCard).join('')}</div></section>
<section class="section split"><div class="panel"><div class="panel-head"><div><h3>最近使用</h3><p>你的任务记录，回到这里继续。</p></div><a href="#/workspace" class="text-link">全部记录 ${icon('arrow',13)}</a></div><div id="homeRecent"><div class="empty-state" style="padding:24px"><strong>暂无记录</strong><p>登录后使用工具，历史会出现在这里。</p></div></div></div><div class="panel" style="position:relative;overflow:hidden;background:linear-gradient(140deg,#fff,#f0f3ff)"><div class="eyebrow">YOUR AI COMPANION</div><h3 style="font-size:23px;line-height:1.25;letter-spacing:-.04em;margin:12px 0">别让琐事占据<br>你的好奇心。</h3><p class="muted" style="font-size:12px;line-height:1.85;max-width:280px">告诉 AI 你想达成什么，从目标、素材和下一步开始。复杂工作，也可以轻盈一点。</p><a class="primary" href="#/ai" style="margin-top:12px">开启 AI 工作台 ${icon('arrow',15)}</a></div></section>
<section class="ai-banner"><div><div class="eyebrow" style="color:#6478d6">NOVA WORKFLOW</div><h3>把零散操作，串成一条聪明的工作流</h3><p>文本工具按顺序执行，上一步输出自动成为下一步输入；每一步都有迹可循。</p></div><div class="banner-right"><div class="banner-orb"></div><a class="secondary" href="#/workflows">探索工作流 ${icon('arrow',14)}</a></div></section>`}
function toolsPage(){S.catFilter='all';const on=S.tools.filter(t=>t.on);return `
${pageTitle('TOOL UNIVERSE','工具大全','把常用能力整理成一套清晰、可组合的工具箱。')}
<section class="panel" style="margin-bottom:19px;background:linear-gradient(110deg,#f4f7ff,#fff 62%,#f4efff)"><div class="toolbar" style="margin:0"><div class="searchbar" style="max-width:none;flex:1"><span>${icon('search')}</span><input id="toolSearch" placeholder="搜索工具、格式或任务…"><span class="kbd">⌘ K</span></div></div><div class="chips" style="margin-top:15px"><button class="chip active" data-filter="all">全部工具 <span>${on.length}</span></button>${Object.entries(CATS).map(([c,l])=>`<button class="chip" data-filter="${c}">${l} <span>${on.filter(t=>t.cat===c).length}</span></button>`).join('')}</div></section>
<div class="section-head"><div><h2>全部工具</h2><p>按场景组织，不需要记住工具的名字。</p></div><span class="small muted"><span id="toolsCount">${on.length}</span> 个工具</span></div>
<div class="tool-library" id="toolGrid">${on.map(toolCard).join('')}</div>
<div class="empty-search" id="toolsEmpty" hidden>没有找到匹配的工具，试试更短的关键词。</div>
<section class="ai-banner"><div><h3>没找到正合适的工具？</h3><p>用 AI 工作台描述你的任务，或把多个工具串成工作流。</p></div><a class="primary" href="#/ai">问 AI 助手 ${icon('arrow',14)}</a></section>`}
async function toolDetail(id){const t=S.tools.find(x=>x.id===id);if(!t||!t.on)return render('<div class="empty-state"><strong>工具不存在或未开启</strong></div>','tool','tool-detail');S.curTool=id;const opts=OPTS[id]||[];if(S.me){try{S.favs=await api('/api/favs')}catch{}}
const favBtn=S.me?`<button class="secondary" data-fav>${(S.favs.includes(id)?'★ 已收藏':'☆ 收藏')}</button>`:'';
const optionsHTML=opts.map(o=>`<div style="margin:18px 0"><div class="eyebrow" style="margin-bottom:9px">${o.label}</div><div class="converter-options" style="margin:0">${o.choices.map((c,i)=>`<button class="option-card ${i===0?'selected':''}" data-optk="${o.k}" data-optv="${c[0]}"><strong>${icon(c[2]||'zap',15)} ${c[1]}</strong><span>${c[3]||''}</span></button>`).join('')}</div></div>`).join('');
if(t.kind==='file'){render(`
${pageTitle('TOOL DETAIL',esc(t.name),esc(t.desc)+' · 文件仅在服务器内存处理，不落盘',`<a class="secondary" href="#/tools">${icon('arrow',14)} 返回工具大全</a>`)}
<div class="workgrid"><section class="panel"><div class="dropzone" id="dropzone"><span class="tool-icon" style="width:54px;height:54px;border-radius:17px;margin:0 auto 12px">${icon('upload',24
