/* Nova Nebula 前端共享底座 NB —— 无构建，classic script
   依赖：tokens.css。后端契约见 ai-studio.js 头部注释
   规约：本文件严禁字面三反引号（用 FENCE 常量构造），防止破坏外部代码围栏 */
(function () {
  'use strict';
  var NB = window.NB = {};

  /* ---------- 工具 ---------- */
  NB.escape = function (s) {
    return String(s == null ? '' : s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  };
  NB.ApiError = function (status, message) { this.status = status; this.message = message; };
  NB.ApiError.prototype = Object.create(Error.prototype);
  NB.debounce = function (fn, ms) { var t; return function () { var a = arguments, self = this; clearTimeout(t); t = setTimeout(function () { fn.apply(self, a); }, ms); }; };
  NB.fmtNum = function (n) { return (n || 0).toLocaleString('zh-CN'); };
  NB.fmtDate = function (ts) { return new Date(ts).toLocaleString('zh-CN'); };

  /* ---------- API 客户端 ---------- */
  function authHeaders() {
    var t = localStorage.getItem('nb_token');
    return t ? { 'Authorization': 'Bearer ' + t } : {};
  }
  function safeJson(res) { return res.json().catch(function () { return {}; }); }

  NB.api = async function (path, opts) {
    opts = opts || {};
    var res = await fetch(path, {
      method: opts.method || 'GET',
      headers: Object.assign({ 'Content-Type': 'application/json' }, authHeaders(), opts.headers || {}),
      body: opts.body != null ? JSON.stringify(opts.body) : undefined
    });
    if (res.status === 401) { let m401 = '请先登录'; try { const b401 = await res.json(); if (b401 && b401.message) m401 = b401.message; } catch (_) {} if (!opts.silent) NB.emitAuthExpired(); throw new NB.ApiError(401, m401); }
    if (res.status === 402) { var info = await safeJson(res); NB.showPaywall(info); throw new NB.ApiError(402, info.message || '该 AI 功能为付费功能'); }
    if (res.status === 410) { NB.toast('该功能已被管理员下线', 'warn'); throw new NB.ApiError(410, '功能已下线'); }
    if (res.status === 429) { NB.toast('请求过于频繁，请稍后再试', 'warn'); throw new NB.ApiError(429, '限流'); }
    if (!res.ok) { var e = await safeJson(res); throw new NB.ApiError(res.status, e.message || res.statusText); }
    return opts.raw ? res : safeJson(res);
  };

  /* SSE 流式消费（POST + text/event-stream，逐行解析 data: 帧） */
  NB.stream = async function (path, body, handlers) {
    var res = await fetch(path, {
      method: 'POST',
      headers: Object.assign({ 'Content-Type': 'application/json', 'Accept': 'text/event-stream' }, authHeaders()),
      body: JSON.stringify(body || {})
    });
    if (!res.ok || !res.body) {
      if (res.status === 401) { NB.emitAuthExpired(); throw new NB.ApiError(401, '请先登录'); }
      if (res.status === 402) { NB.showPaywall(await safeJson(res)); throw new NB.ApiError(402, '付费功能'); }
      var e0 = await safeJson(res);
      handlers.onError && handlers.onError(new NB.ApiError(res.status, e0.message || '流式请求失败'));
      return;
    }
    var reader = res.body.getReader(), dec = new TextDecoder(), buf = '';
    while (true) {
      var chunk = await reader.read();
      if (chunk.done) break;
      buf += dec.decode(chunk.value, { stream: true });
      var idx;
      while ((idx = buf.indexOf('\n\n')) !== -1) {
        var frame = buf.slice(0, idx); buf = buf.slice(idx + 2);
        frame.split('\n').forEach(function (line) {
          if (line.indexOf('data:') !== 0) return;
          var data = line.slice(5).trim();
          if (data === '[DONE]') { handlers.onDone && handlers.onDone(); return; }
          try { handlers.onEvent && handlers.onEvent(JSON.parse(data)); }
          catch (err) { /* 忽略半包帧 */ }
        });
      }
    }
    handlers.onDone && handlers.onDone();
  };

  /* ---------- 认证 ---------- */
  var meCache = null;
  NB.auth = {
    token: function () { return localStorage.getItem('nb_token'); },
    async me(force) {
      if (meCache && !force) return meCache;
      if (!NB.auth.token()) return null;
      try { meCache = (await NB.api('/api/v2/auth/me', { silent: true })).user; } catch (e) { meCache = null; }
      return meCache;
    },
    async login(email, password) {
      var r = await NB.api('/api/v2/auth/login', { method: 'POST', silent: true, body: { email: email, password: password } });
      localStorage.setItem('nb_token', r.token); meCache = r.user; return r.user;
    },
    async register(email, password) {
      var r = await NB.api('/api/v2/auth/register', { method: 'POST', silent: true, body: { email: email, password: password } });
      localStorage.setItem('nb_token', r.token); meCache = r.user; return r.user;
    },
    logout() { localStorage.removeItem('nb_token'); meCache = null; location.href = '/'; },
    isAdmin(user) { return user && (user.role === 'admin' || user.role === 'super'); }
  };
  NB.emitAuthExpired = function () {
    if (NB._authToast) return;
    NB._authToast = true;
    NB.toast('登录已过期，请重新登录', 'warn');
    setTimeout(function () { NB._authToast = false; }, 5000);
  };

  /* 顶栏用户位 */
  NB.bindUserChip = async function (el) {
    var user = await NB.auth.me();
    el.innerHTML = user
      ? '<span class="nb-user-chip">' + NB.escape(user.email) + '</span>' +
        (NB.auth.isAdmin(user) ? '<a class="nb-btn nb-btn--ghost nb-btn--sm" href="/admin/">后台</a>' : '') +
        '<a class="nb-btn nb-btn--ghost nb-btn--sm" href="/profile.html">个人中心</a>' +
        '<button class="nb-btn nb-btn--ghost nb-btn--sm" id="nb-logout">退出</button>'
      : '<a class="nb-btn nb-btn--ghost nb-btn--sm" href="/login.html">登录</a>' +
        '<a class="nb-btn nb-btn--warm nb-btn--sm" href="/register.html">注册</a>';
    var lo = document.getElementById('nb-logout');
    if (lo) lo.onclick = function () { NB.auth.logout(); };
  };

  /* ---------- Toast / Modal / 付费门 ---------- */
  NB.toast = function (msg, type) {
    var box = document.getElementById('nb-toasts');
    if (!box) { box = document.createElement('div'); box.id = 'nb-toasts'; document.body.appendChild(box); }
    var t = document.createElement('div');
    t.className = 'nb-toast' + (type ? ' nb-toast--' + type : '');
    t.setAttribute('role', 'status'); t.textContent = msg;
    box.appendChild(t);
    setTimeout(function () { t.remove(); }, 3200);
  };
  NB.modal = function (opts) {
    var mask = document.createElement('div');
    mask.className = 'nb-modal-mask';
    mask.innerHTML =
      '<div class="nb-modal" role="dialog" aria-modal="true" aria-label="' + NB.escape(opts.title) + '">' +
      '<h3>' + NB.escape(opts.title) + '</h3>' +
      '<div class="nb-modal-body">' + (opts.html || NB.escape(opts.body || '')) + '</div>' +
      '<div class="nb-modal-actions"></div></div>';
    var actions = mask.querySelector('.nb-modal-actions');
    (opts.actions || [{ label: '知道了', cls: 'nb-btn--warm' }]).forEach(function (a) {
      var b = document.createElement('button');
      b.className = 'nb-btn ' + (a.cls || 'nb-btn--ghost');
      b.textContent = a.label;
      b.onclick = function () { mask.remove(); a.onClick && a.onClick(); };
      actions.appendChild(b);
    });
    mask.addEventListener('click', function (e) { if (e.target === mask) mask.remove(); });
    document.body.appendChild(mask);
    var first = mask.querySelector('button'); if (first) first.focus();
    return function close() { mask.remove(); };
  };
  NB.showPaywall = function (info) {
    NB.modal({
      title: '该 AI 功能为付费功能',
      html: '<p>' + NB.escape(info.message || '当前模块已开启付费调用。') + '</p>' +
        '<p>开通后即可继续使用「' + NB.escape(info.module || 'AI') + '」相关能力。</p>',
      actions: [
        { label: '暂不', cls: 'nb-btn--ghost' },
        { label: '去开通', cls: 'nb-btn--ai', onClick: function () { location.href = '/profile.html#billing'; } }
      ]
    });
  };

  /* ---------- Markdown 渲染（零依赖迷你实现） ---------- */
  /* 三个反引号字符用码点构造，避免字面量破坏外部代码围栏 */
  var FENCE = String.fromCharCode(96, 96, 96);
  var fenceRe = new RegExp(FENCE + '([\\w+-]*)\\n?([\\s\\S]*?)' + FENCE, 'g');
  NB.md = function (src) {
    src = String(src == null ? '' : src);
    var blocks = [];
    src = src.replace(fenceRe, function (m, lang, code) {
      blocks.push('<pre class="nb-code"><code>' + NB.escape(code.replace(/\n$/, '')) + '</code>' +
        (lang ? '<span class="nb-code-lang">' + NB.escape(lang) + '</span>' : '') + '</pre>');
      return '\u0000B' + (blocks.length - 1) + '\u0000';
    });
    var inline = function (s) {
      s = NB.escape(s);
      s = s.replace(/`([^`]+)`/g, '<code>$1</code>');
      s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
      s = s.replace(/\*([^*\s]+)\*/g, '<em>$1</em>');
      s = s.replace(/\[([^\]]+)\]\((https?:\/\/[^)\s]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');
      return s;
    };
    var lines = src.split('\n'), out = [], lists = [];
    function closeLists() { while (lists.length) out.push('</' + lists.pop() + '>'); }
    var i = 0;
    while (i < lines.length) {
      var line = lines[i];
      var bm = line.match(/^\u0000B(\d+)\u0000\s*$/);
      if (bm) { closeLists(); out.push(blocks[+bm[1]]); i++; continue; }
      if (/^\s*$/.test(line)) { closeLists(); i++; continue; }
      var h = line.match(/^(#{1,6})\s+(.*)$/);
      if (h) { closeLists(); out.push('<h' + h[1].length + '>' + inline(h[2]) + '</h' + h[1].length + '>'); i++; continue; }
      if (/^\s*(-{3,}|\*{3,})\s*$/.test(line)) { closeLists(); out.push('<hr>'); i++; continue; }
      if (/^\s*>/.test(line)) {
        closeLists(); var q = [];
        while (i < lines.length && /^\s*>/.test(lines[i])) { q.push(inline(lines[i].replace(/^\s*>\s?/, ''))); i++; }
        out.push('<blockquote>' + q.join('<br>') + '</blockquote>'); continue;
      }
      if (line.indexOf('|') > -1 && i + 1 < lines.length && /^\s*\|?\s*:?-{2,}[-\s:|]*$/.test(lines[i + 1])) {
        closeLists();
        var cells = function (l) { var c = l.split('|').map(function (x) { return x.trim(); }); if (c[0] === '') c.shift(); if (c.length && c[c.length - 1] === '') c.pop(); return c; };
        var head = cells(line), rows = []; i += 2;
        while (i < lines.length && lines[i].indexOf('|') > -1 && !/^\s*$/.test(lines[i])) { rows.push(cells(lines[i])); i++; }
        out.push('<table class="nb-table"><thead><tr>' + head.map(function (c) { return '<th>' + inline(c) + '</th>'; }).join('') + '</tr></thead><tbody>' +
          rows.map(function (r) { return '<tr>' + r.map(function (c) { return '<td>' + inline(c) + '</td>'; }).join('') + '</tr>'; }).join('') + '</tbody></table>');
        continue;
      }
      var ul = line.match(/^\s*[-*]\s+(.*)$/), ol = line.match(/^\s*\d+[.)]\s+(.*)$/);
      if (ul || ol) {
        var tag = ul ? 'ul' : 'ol';
        if (lists[lists.length - 1] !== tag) { closeLists(); out.push('<' + tag + '>'); lists.push(tag); }
        out.push('<li>' + inline((ul || ol)[1]) + '</li>'); i++; continue;
      }
      closeLists();
      var para = [inline(line)]; i++;
      while (i < lines.length && !/^\s*$/.test(lines[i]) && !/^(#{1,6}\s|\s*[-*]\s|\s*\d+[.)]\s|\s*>|\u0000B)/.test(lines[i])) { para.push(inline(lines[i])); i++; }
      out.push('<p>' + para.join('<br>') + '</p>');
    }
    closeLists();
    return out.join('\n');
  };
})();
