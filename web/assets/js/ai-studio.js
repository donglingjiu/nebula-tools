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
    if (window.NB && typeof window.NB.showPaywall === 'function') { window.NB.showPaywall({ price: price, message: msg }); return; }
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
    if (r.status === 402) { buyFlow(r); return; }
    note(r.msg);
  }
  function buyFlow(r) {
    paywall(r.price, r.msg);
    var o = document.getElementById('nbsw-out'); if (!o) return;
    var b = document.createElement('button');
    b.textContent = '模拟购买并重试(¥' + (r.price || '?') + ')';
    b.style.cssText = 'margin:6px 0;padding:8px 16px;background:#d97706;color:#fff;border:0;border-radius:8px;cursor:pointer';
    b.onclick = async function () {
      b.disabled = true;
      var c = await api('POST', '/api/ai/order', { feature_key: 'ai_chat' });
      if (!c.ok) { note(c.msg); b.disabled = false; return; }
      var pr = await api('POST', '/api/ai/order/pay', { id: c.data.id });
      if (!pr.ok) { note(pr.msg); b.disabled = false; return; }
      note('已开通(模拟), 自动重试...');
      b.remove();
      if (typeof window.nbswRetry === 'function') window.nbswRetry();
    };
    o.appendChild(b);
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
    window.nbswFlow = f;
    if (typeof window.nbswFillEdit === 'function') window.nbswFillEdit();
  }
  async function createRun() {
    window.nbswRetry = createRun;
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
    if (r2.data && r2.data.state === 'error') {
      note('生成失败: 上游 AI 未配置或不可达(见步骤 error), 配置 AI Profile 后可「推进一步」重试');
    } else { note('七步全部完成'); }
  }
  async function stepAdvance() {
    window.nbswRetry = stepAdvance;
    if (!window.nbswFlowId) { note('先创建流程'); return; }
    var r = await api('POST', '/api/ai/flow/advance',
      { id: window.nbswFlowId });
    if (!r.ok) { handleErr(r); return; }
    renderFlow(r.data.flow);
  }
  async function chat() {
    window.nbswRetry = chat;
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
  // ===== OPT-009 (nt14): 标准七步 + 历史流程 =====
  async function runStd7(idea) {
    window.nbswRetry = runStd7;
    var r = await api('POST', '/api/ai/flow/start7', { title: '', idea: idea, material: (window.nbswMaterial || ''), material_ids: (function(){var v=new URLSearchParams(location.search).get('material_id');if(v){try{history.replaceState({},'',location.pathname)}catch(e){}return [v];}return [];})(), profile_id: ((document.getElementById('nbsw-prof') || {}).value || '') });
    if (!r.ok) { handleErr(r); return; }
    renderFlow(r.data);
    note('七步流程已创建, 服务端逐步生成中（7 步×真实 AI, 稍等）...');
    var r2 = await api('POST', '/api/ai/flow/run', { id: r.data.id });
    if (!r2.ok) { handleErr(r2); note('流程已创建, 可用「推进一步」逐段生成'); return; }
    renderFlow(r2.data);
  }
  async function loadHist() {
    window.nbswRetry = loadHist;
    var r = await api('GET', '/api/ai/flow/list');
    if (!r.ok) { handleErr(r); return; }
    var hp = document.getElementById('nbsw-hist');
    if (!hp) return;
    var fs = r.data || [];
    if (!fs.length) {
      hp.innerHTML = '<div style="color:#64748b;font-size:13px;padding:6px 0">暂无历史流程</div>';
      return;
    }
    hp.innerHTML = '';
    fs.slice(0, 10).forEach(function (f) {
      var it = document.createElement('div');
      it.style.cssText = 'display:flex;justify-content:space-between;align-items:center;gap:8px;padding:8px 10px;margin:4px 0;background:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;font-size:13px';
      var left = document.createElement('span');
      left.style.cssText = 'overflow:hidden;text-overflow:ellipsis;white-space:nowrap';
      left.textContent = (f.title || '(无标题)') + ' · ' + f.state + ' · ' + String(f.created_at || '').slice(0, 16).replace('T', ' ');
      var open = document.createElement('button');
      open.textContent = '打开';
      open.style.cssText = 'flex:none;padding:4px 10px;background:#7c3aed;color:#fff;border:0;border-radius:6px;cursor:pointer';
      open.onclick = function () {
        window.nbswFlowId = f.id;
        renderFlow(f);
        note('已载入: ' + f.title + '（可推进一步 / 追问 / 导出）');
      };
      it.appendChild(left);
      it.appendChild(open);
      hp.appendChild(it);
    });
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
    // OPT-009: 标准七步入口 + 历史流程
    var b7 = document.createElement('button');
    b7.textContent = '▶ 标准七步（一句话 → 成品方案）';
    b7.style.cssText = 'margin:6px 6px 0 0;padding:10px 16px;background:#7c3aed;color:#fff;border:0;border-radius:10px;cursor:pointer;font-size:14px';
    b7.onclick = async function () {
      var idea = (document.getElementById('nbsw-idea').value || '').trim();
      if (!idea) { note('请先在上方输入你的想法'); return; }
      b7.disabled = true;
      await runStd7(idea);
      b7.disabled = false;
    };
    var hb = document.createElement('button');
    hb.textContent = '↻ 历史流程';
    hb.style.cssText = 'margin:6px 6px 0 0;padding:10px 16px;background:#475569;color:#fff;border:0;border-radius:10px;cursor:pointer;font-size:14px';
    hb.onclick = loadHist;
    var hp = document.createElement('div');
    hp.id = 'nbsw-hist';
    hp.style.cssText = 'margin:10px 0 0';
    // OPT-012: model select + material upload + step edit
    var ps = document.createElement('select');
    ps.id = 'nbsw-prof';
    ps.style.cssText = 'margin:6px 6px 0 0;padding:8px;border:1px solid #cbd5e1;border-radius:8px;font-size:13px;vertical-align:top';
    ps.innerHTML = '<option value="">模型: 默认</option>';
    api('GET', '/api/ai/flow/profiles').then(function (r) {
      if (!r.ok) return;
      (r.data || []).forEach(function (p) {
        var o = document.createElement('option');
        o.value = p.id; o.textContent = '模型: ' + p.name + ' (' + p.model + ')';
        ps.appendChild(o);
      });
    });
    var mf = document.createElement('input');
    mf.type = 'file'; mf.accept = '.txt,.md,.csv,.markdown';
    mf.style.cssText = 'margin:6px 6px 0 0;font-size:13px;vertical-align:top';
    mf.onchange = function () {
      var file = mf.files && mf.files[0];
      if (!file) return;
      var rd = new FileReader();
      rd.onload = function () {
        window.nbswMaterial = String(rd.result || '');
        note('已载入材料: ' + file.name + ' (' + window.nbswMaterial.length + ' 字, 随第一步下发)');
      };
      rd.readAsText(file);
    };
    var es = document.createElement('select');
    es.id = 'nbsw-edit-step';
    es.style.cssText = 'margin:6px 6px 0 0;padding:8px;border:1px solid #cbd5e1;border-radius:8px;font-size:13px';
    var et = document.createElement('textarea');
    et.id = 'nbsw-edit-out';
    et.placeholder = '选择已完成步骤后修改其产出, 后续步骤将基于修改后内容继续';
    et.style.cssText = 'width:100%;min-height:90px;margin:6px 0;padding:8px;border:1px solid #cbd5e1;border-radius:8px;font-size:13px;font-family:inherit';
    var eb = document.createElement('button');
    eb.textContent = '保存步骤产出';
    eb.style.cssText = 'margin:6px 6px 0 0;padding:8px 14px;background:#0ea5e9;color:#fff;border:0;border-radius:8px;cursor:pointer;font-size:13px';
    eb.onclick = async function () {
      var f = window.nbswFlow;
      if (!f) { note('先创建或载入流程'); return; }
      var i = parseInt(es.value, 10);
      if (isNaN(i)) { note('先选择要编辑的步骤'); return; }
      var r = await api('POST', '/api/ai/flow/edit-step', { id: f.id, index: i, output: et.value });
      if (!r.ok) { handleErr(r); return; }
      renderFlow(r.data.flow);
      note('步骤产出已更新, 「推进一步」将基于新内容继续');
    };
    window.nbswFillEdit = function () {
      var f = window.nbswFlow;
      var cur = es.value;
      es.innerHTML = '';
      var any = false;
      ((f && f.steps) || []).forEach(function (s, i) {
        if (s.status === 'done') {
          any = true;
          var o = document.createElement('option');
          o.value = i; o.textContent = (i + 1) + '. ' + s.name;
          es.appendChild(o);
        }
      });
      if (!any) { es.innerHTML = '<option value="">（暂无已完成步骤）</option>'; et.value = ''; return; }
      var has = Array.prototype.some.call(es.options, function (op) { return op.value === cur; });
      if (has) es.value = cur;
      et.value = (f.steps[parseInt(es.value, 10)] || {}).output || '';
    };
    w.appendChild(ps); w.appendChild(mf);
    w.appendChild(b7); w.appendChild(hb);
    w.appendChild(document.createElement('hr'));
    w.appendChild(es); w.appendChild(eb); w.appendChild(et); w.appendChild(hp);
    loadHist();
    loadHist();
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', ui);
  } else { ui(); }
})();


// OPT-016c+: material_id 反填 —— idea 预填材料标题（材料内容仍经 material_ids 注入第一步）
(function () {
  function prefillMat() {
    var v = new URLSearchParams(location.search).get('material_id');
    if (!v) return;
    var tok = localStorage.getItem('nb_token') || localStorage.getItem('token') || '';
    fetch('/api/ai/materials/get?id=' + encodeURIComponent(v), { headers: { 'Authorization': 'Bearer ' + tok } })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (m) {
        if (!m || !m.title) return;
        var el = document.getElementById('nbsw-idea');
        if (el && !el.value) {
          el.value = m.title;
          try { el.dispatchEvent(new Event('input', { bubbles: true })); } catch (e) {}
        }
        try { console.log('[materials] 已引入材料《' + m.title + '》：内容将随 material_ids 注入第一步，想法框已反填标题'); } catch (e) {}
      })
      .catch(function () {});
  }
  if (document.readyState === 'loading') { document.addEventListener('DOMContentLoaded', prefillMat); } else { prefillMat(); }
})();
