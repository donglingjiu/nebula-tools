/* AI 工作台 (6-2重写) — 真实API契约(旧stage契约已退役):
   POST /api/ai/flow/start {title, steps:[{name,prompt}]} -> Flow{id,title,state,steps[]}
   POST /api/ai/flow/run {id} -> Flow(state done/error)
   POST /api/ai/flow/advance {id} -> {done, step?, flow}
   POST /api/ai/flow/chat {id,message} -> {reply, flow}
   GET  /api/ai/flow/export?id= -> markdown(需Authorization, blob下载)
   门语义: 402带price(付费墙) / 410下线 / 503上游未配置 / 404不存在或非本人
   本文件自包含(JWT扫描+own fetch), 不依赖NB.* */
(function () {
  function tok() {
    try {
      for (var i = 0; i < localStorage.length; i++) {
        var v = localStorage.getItem(localStorage.key(i)) || '';
        var m = v.match(/eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}/);
        if (m) return m[0];
      }
    } catch (e) {}
    return '';
  }
  function hdrs() {
    var h = { 'Content-Type': 'application/json' };
    var t = tok(); if (t) h['Authorization'] = 'Bearer ' + t;
    return h;
  }
  function note(msg) {
    var o = document.getElementById('nbsw-out'); if (!o) return;
    var d = document.createElement('div');
    d.style.cssText = 'padding:8px 10px;margin:6px 0;border-radius:8px;background:#eff6ff;color:#1e40af;font-size:13px';
    d.textContent = msg; o.appendChild(d);
  }
  function paywall(price, msg) {
    if (typeof window.showPaywall === 'function') { window.showPaywall(price, msg); return; }
    note(msg);
  }
  async function api(method, url, body) {
    var opt = { method: method, headers: hdrs() };
    if (body) opt.body = JSON.stringify(body);
    var res = await fetch(url, opt);
    if (res.ok) return { ok: true, data: await res.json() };
    var e = null; try { e = await res.json(); } catch (_) {}
    return { ok: false, status: res.status,
      msg: (e && e.message) || ('请求失败 ' + res.status),
      price: e && e.price };
  }
  function handleErr(r) {
    if (r.status === 402) paywall(r.price, r.msg);
    else note(r.msg);
  }
  function renderFlow(f) {
    var o = document.getElementById('nbsw-out');
    var line = document.getElementById('nbsw-flow') ||
      document.createElement('div');
    line.id = 'nbsw-flow';
    line.style.cssText = 'margin:8px 0;padding:10px;border-left:3px solid #7c3aed;white-space:pre-wrap;font-size:14px;background:#faf5ff;border-radius:8px';
    var txt = '流程[' + f.title + '] ' + f.state + '\n';
    (f.steps || []).forEach(function (s) {
      txt += '[' + s.name + '] ' + s.status +
        (s.output ? '\n' + s.output : '') + '\n\n';
    });
    line.textContent = txt;
    if (!line.parentNode) o.appendChild(line);
    window.nbswFlowId = f.id;
    var ex = document.getElementById('nbsw-export');
    if (ex) ex.style.display = 'inline-block';
  }
  async function createRun() {
    var idea = (document.getElementById('nbsw-idea').value || '').trim();
    if (!idea) { note('请输入想法'); return; }
    var mode = document.getElementById('nbsw-mode').value;
    var steps = mode === 'single'
      ? [{ name: '成稿', prompt: idea }]
      : [{ name: '提纲', prompt: '为主题「' + idea + '」列出提纲' },
         { name: '正文', prompt: '按提纲写正文: ' + idea }];
    var r = await api('POST', '/api/ai/flow/start',
      { title: idea.slice(0, 30), steps: steps });
    if (!r.ok) { handleErr(r); return; }
    renderFlow(r.data);
    var r2 = await api('POST', '/api/ai/flow/run', { id: r.data.id });
    if (!r2.ok) { handleErr(r2); note('流程已创建, 可用「推进一步」继续'); return; }
    renderFlow(r2.data);
  }
  async function stepAdvance() {
    if (!window.nbswFlowId) { note('先创建流程'); return; }
    var r = await api('POST', '/api/ai/flow/advance',
      { id: window.nbswFlowId });
    if (!r.ok) { handleErr(r); return; }
    renderFlow(r.data.flow);
  }
  async function chat() {
    var inp = document.getElementById('nbsw-cin');
    var q = (inp.value || '').trim(); if (!q) return;
    if (!window.nbswFlowId) { note('先创建流程, 对话将带上流程上下文'); return; }
    var r = await api('POST', '/api/ai/flow/chat',
      { id: window.nbswFlowId, message: q });
    inp.value = '';
    if (!r.ok) { handleErr(r); return; }
    renderFlow(r.data.flow);
    note('AI: ' + r.data.reply);
  }
  async function exportMd() {
    if (!window.nbswFlowId) { note('先创建流程'); return; }
    try {
      var res = await fetch('/api/ai/flow/export?id=' +
        window.nbswFlowId, { headers: hdrs() });
      if (!res.ok) {
        var e = null; try { e = await res.json(); } catch (_) {}
        note((e && e.message) || ('导出失败 ' + res.status)); return;
      }
      var b = await res.blob();
      var a = document.createElement('a');
      a.href = URL.createObjectURL(b);
      a.download = 'flow-' + window.nbswFlowId + '.md';
      a.click();
      URL.revokeObjectURL(a.href);
    } catch (err) { note('网络异常: ' + err); }
  }
  function ui() {
    if (document.getElementById('nbsw-out')) return;
    var w = document.createElement('div');
    w.style.cssText = 'max-width:860px;margin:16px auto;padding:16px;background:#fff;border:1px solid #eee;border-radius:12px';
    w.innerHTML = '<h3 style="margin:0 0 8px">AI 工作台 · 流程模式</h3>'
      + '<textarea id="nbsw-idea" rows="2" placeholder="输入想法, 例如: 写一份新人入职指南" style="width:100%;padding:8px;border:1px solid #ddd;border-radius:8px"></textarea>'
      + '<select id="nbsw-mode" style="margin:8px 0;padding:6px;border:1px solid #ddd;border-radius:8px">'
      + '<option value="full">两步(提纲→正文)</option>'
      + '<option value="single">单步直出</option></select> '
      + '<button id="nbsw-go" style="padding:8px 18px;background:#7c3aed;color:#fff;border:0;border-radius:8px;cursor:pointer">创建并运行</button> '
      + '<button id="nbsw-adv" style="padding:8px 14px;background:#64748b;color:#fff;border:0;border-radius:8px;cursor:pointer">推进一步</button> '
      + '<button id="nbsw-export" style="padding:8px 14px;background:#059669;color:#fff;border:0;border-radius:8px;cursor:pointer;display:none">导出Markdown</button>'
      + '<div style="margin:12px 0 6px;font-weight:600">追问(带流程上下文)</div>'
      + '<input id="nbsw-cin" placeholder="例如: 把第2步扩写成3个场景" style="width:100%;padding:6px 8px;border:1px solid #ddd;border-radius:8px">'
      + '<button id="nbsw-cgo" style="margin:8px 0;padding:8px 18px;background:#2563eb;color:#fff;border:0;border-radius:8px;cursor:pointer">发送</button>'
      + '<div id="nbsw-out" style="margin-top:8px;font-size:14px"></div>';
    document.body.appendChild(w);
    document.getElementById('nbsw-go').onclick = createRun;
    document.getElementById('nbsw-adv').onclick = stepAdvance;
    document.getElementById('nbsw-cgo').onclick = chat;
    document.getElementById('nbsw-export').onclick = exportMd;
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', ui);
  } else { ui(); }
})();
