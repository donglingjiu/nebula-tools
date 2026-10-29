/* AI 全流程工作台 —— 7 阶段状态机
   ============ 前后端契约（阶段5按此实现）============
   POST /api/ai/flow/start {idea, mode:"full"|"single"} -> {flow_id, model, stages:[{id,status}]}
   GET  /api/ai/flow/{flow_id} -> {flow_id, idea, model, mode, stages:[...]}
   POST /api/ai/flow/{flow_id}/stage/{stage_id}/run   -> SSE: {type:"delta",text} {type:"usage",tokens_in,tokens_out} {type:"stage_done",stage_id,content,ms} {type:"error",message}
   POST /api/ai/flow/{flow_id}/stage/{stage_id}/chat {message} -> SSE 同上
   PUT  /api/ai/flow/{flow_id}/stage/{stage_id} {content} -> {ok, updated_at}
   GET  /api/ai/flow/{flow_id}/export?format=zip|md   -> 文件流
   错误码：401 未登录 / 402 付费门 / 410 功能下线 / 429 限流
   ==================================================== */
(function () {
  'use strict';
  var NB = window.NB;
  var STAGES = [
    { id: 'brainstorm', label: '头脑风暴', hint: '发散角度、用户画像与创意方向' },
    { id: 'evaluate',   label: '需求评估', hint: '价值/可行性/风险打分与结论' },
    { id: 'analyze',    label: '需求分析', hint: '用户故事、场景与非功能需求' },
    { id: 'design',     label: '方案设计', hint: '架构、技术选型与数据流' },
    { id: 'breakdown',  label: '功能拆解', hint: '功能点清单、优先级与里程碑' },
    { id: 'implement',  label: '代码实现', hint: 'RESTful 接口与代码片段' },
    { id: 'testing',    label: '测试落地', hint: '用例设计与验收清单' }
  ];
  var $ = function (id) { return document.getElementById(id); };
  var S = { flowId: null, idea: '', model: '', mode: 'full', cur: 0, busy: false,
            stages: STAGES.map(function (s) { return { id: s.id, status: 'pending', content: '', tin: 0, tout: 0, ms: 0 }; }) };

  function renderStepper() {
    var ul = $('stepper'); ul.innerHTML = '';
    STAGES.forEach(function (st, i) {
      var li = document.createElement('li');
      var stt = S.stages[i].status;
      li.className = (stt === 'done' ? 'done ' : '') + (i === S.cur ? 'current ' : '') + (S.busy && i === S.cur ? 'busy' : '');
      var b = document.createElement('button');
      b.innerHTML = '<span class="nb-step-dot">' + (i + 1) + '</span><span>' + st.label +
        (stt === 'done' ? ' <span style="color:var(--ok);font-size:11px">✓</span>' : '') + '</span>';
      b.title = st.hint;
      b.setAttribute('aria-current', i === S.cur ? 'step' : 'false');
      b.onclick = function () { if (!S.busy) gotoStage(i); };
      li.appendChild(b); ul.appendChild(li);
    });
  }
  function renderOut() {
    var box = $('outList'); box.innerHTML = '';
    var doneN = 0;
    S.stages.forEach(function (st, i) {
      if (st.status !== 'done') return; doneN++;
      var b = document.createElement('button');
      b.className = 'out-item';
      b.innerHTML = '<span class="dot"></span>' + STAGES[i].label;
      b.onclick = function () { if (!S.busy) gotoStage(i); };
      box.appendChild(b);
    });
    if (!doneN) box.innerHTML = '<p class="hint-ai">还没有完成的阶段，先跑起来吧～</p>';
    $('btnExportZip').disabled = $('btnExportMd').disabled = doneN === 0;
  }
  function renderStage() {
    var st = STAGES[S.cur], d = S.stages[S.cur];
    $('stageTitle').textContent = (S.cur + 1) + '. ' + st.label;
    $('stageContent').innerHTML = d.content ? NB.md(d.content) :
      '<div class="nb-empty"><div class="big">NB</div><p>' + st.hint + '，点击「重新生成」开始本阶段。</p></div>';
    setStatus();
  }
  function setStatus() {
    $('stModel').textContent = 'model: ' + (S.model || '–');
    $('stStage').textContent = 'stage: ' + STAGES[S.cur].id;
    var tin = 0, tout = 0;
    S.stages.forEach(function (x) { tin += x.tin; tout += x.tout; });
    $('stTokens').textContent = (tin + tout).toLocaleString('zh-CN');
    $('stMeter').style.width = Math.min(100, (tin + tout) / 100) + '%';
    var d = S.stages[S.cur];
    $('stTime').textContent = d.ms ? '本阶段 ' + (d.ms / 1000).toFixed(1) + 's' : '';
�  }
  var throttledRender = NB.debounce(renderStage, 150);

  function gotoStage(i) {
    S.cur = Math.max(0, Math.min(STAGES.length - 1, i));
    $('phaseIdea').hidden = true; $('phaseStage').hidden = false;
    renderStepper(); renderStage();
    $('chatLog').innerHTML = '';
  }
  async function runStage(i) {
    if (S.busy || !S.flowId) return;
    S.busy = true; S.cur = i; S.stages[i].status = 'running'; S.stages[i].content = '';
    gotoStage(i); renderStepper();
    $('btnRegen').disabled = $('btnNext').disabled = $('btnChat').disabled = true;
    var t0 = Date.now();
    try {
      await NB.stream('/api/ai/flow/' + S.flowId + '/stage/' + STAGES[i].id + '/run', {}, {
        onEvent: function (ev) {
          if (ev.type === 'delta') { S.stages[i].content += ev.text || ''; throttledRender(); }
          else if (ev.type === 'usage') { S.stages[i].tin += ev.tokens_in || 0; S.stages[i].tout += ev.tokens_out || 0; setStatus(); }
          else if (ev.type === 'stage_done') {
            S.stages[i].status = 'done';
            if (ev.content) S.stages[i].content = ev.content;
            S.stages[i].ms = ev.ms || (Date.now() - t0);
          } else if (ev.type === 'error') { NB.toast(ev.message || '本阶段生成失败', 'err'); }
        },
        onError: function (e) { NB.toast(e.message || '请求失败', 'err'); },
        onDone: function () {
          if (S.stages[i].status !== 'done') S.stages[i].status = 'pending';
          S.busy = false;
          $('btnRegen').disabled = $('btnChat').disabled = false;
          renderStepper(); renderStage(); renderOut();
          if (S.mode === 'full' && S.stages[i].status === 'done' && i < STAGES.length - 1) {
            NB.toast(STAGES[i].label + ' 完成，自动进入下一阶段');
            setTimeout(function () { runStage(i + 1); }, 600);
          } else {
            $('btnNext').disabled = !(S.stages[i].status === 'done' && i < STAGES.length - 1);
          }
        }
      });
    } catch (e) { S.busy = false; renderStepper(); }
  }

  $('btnStart').onclick = async function () {
    if (!NB.auth.token()) { $('loginGate').hidden = false; return; }
    var idea = $('ideaInput').value.trim();
    if (idea.length < 5) { NB.toast('想法至少写 5 个字哦', 'warn'); return; }
    S.mode = document.querySelector('input[name=mode]:checked').value;
    this.disabled = true; this.textContent = '创建流程中…';
    try {
      var r = await NB.api('/api/ai/flow/start', { method: 'POST', body: { idea: idea, mode: S.mode } });
      S.flowId = r.flow_id; S.idea = idea; S.model = r.model || '';
      localStorage.setItem('nb_flow', JSON.stringify({ flow_id: r.flow_id, idea: idea }));
      runStage(0);
    } catch (e) { NB.toast(e.message || '创建失败', 'err'); }
    this.disabled = false; this.textContent = '开始全流程 →';
  };
  $('btnRegen').onclick = function () { runStage(S.cur); };
  $('btnNext').onclick = function () { runStage(S.cur + 1); };
  $('btnEdit').onclick = function () {
    var on = $('editArea').hidden;
    $('editArea').hidden = !on; $('editActions').hidden = !on; $('stageContent').hidden = on;
    if (on) $('editArea').value = S.stages[S.cur].content;
  };
  $('btnCancel').onclick = function () { $('editArea').hidden = $('editActions').hidden = true; $('stageContent').hidden = false; };
  $('btnSave').onclick = async function () {
    var val = $('editArea').value;
    try {
      await NB.api('/api/ai/flow/' + S.flowId + '/stage/' + STAGES[S.cur].id, { method: 'PUT', body: { content: val } });
      S.stages[S.cur].content = val;
      $('btnCancel').click(); renderStage(); NB.toast('已保存');
    } catch (e) { NB.toast(e.message || '保存失败', 'err'); }
  };
  function chatBubble(role, text) {
    var d = document.createElement('div');
    d.className = 'chat-msg chat-msg--' + role; d.textContent = text;
    $('chatLog').appendChild(d); $('chatLog').scrollTop = 1e9;
    return d;
  }
  $('btnChat').onclick = async function () {
    var msg = $('chatInput').value.trim();
    if (!msg || S.busy || !S.flowId) return;
    $('chatInput').value = ''; chatBubble('u', msg);
    var bubble = chatBubble('a', '…'); S.busy = true; this.disabled = true;
    try {
      await NB.stream('/api/ai/flow/' + S.flowId + '/stage/' + STAGES[S.cur].id + '/chat', { message: msg }, {
        onEvent: function (ev) {
          if (ev.type === 'delta') { bubble.textContent += ev.text || ''; $('chatLog').scrollTop = 1e9; }
          else if (ev.type === 'usage') { S.stages[S.cur].tin += ev.tokens_in || 0; S.stages[S.cur].tout += ev.tokens_out || 0; setStatus(); }
          else if (ev.type === 'stage_done') {
            if (ev.content) { S.stages[S.cur].content = ev.content; renderStage(); }
            bubble.textContent = '已并入本阶段产出 ↑';
          }
        },
        onError: function (e) { bubble.textContent = '失败：' + e.message; },
        onDone: function () { S.busy = false; $('btnChat').disabled = false; }
      });
    } catch (e) { bubble.textContent = '失败：' + (e.message || '未知错误'); S.busy = false; this.disabled = false; }
  };
  $('chatInput').addEventListener('keydown', function (e) { if (e.key === 'Enter') $('btnChat').click(); });

  function download(blob, name) {
    var a = document.createElement('a');
    a.href = URL.createObjectURL(blob); a.download = name;
    a.click(); setTimeout(function () { URL.revokeObjectURL(a.href); }, 4000);
  }
  function clientMd() {
    var out = ['# AI 全流程产出 · ' + (S.idea || ''), '', '> 模型: ' + (S.model || '–') + ' · 导出时间: ' + new Date().toLocaleString('zh-CN'), ''];
    STAGES.forEach(function (st, i) { if (S.stages[i].content) { out.push('## ' + (i + 1) + '. ' + st.label, '', S.stages[i].content, ''); } });
    return out.join('\n');
  }
  $('btnExportZip').onclick = async function () {
    try { var res = await NB.api('/api/ai/flow/' + S.flowId + '/export?format=zip', { raw: true });
      download(await res.blob(), 'nova-ai-flow-' + S.flowId + '.zip'); }
    catch (e) { NB.toast('zip 导出失败，已改用 md 导出', 'warn'); download(new Blob([clientMd()], { type: 'text/markdown' }), 'nova-ai-flow.md'); }
  };
  $('btnExportMd').onclick = function () { download(new Blob([clientMd()], { type: 'text/markdown' }), 'nova-ai-flow.md'); };
  $('btnNewFlow').onclick = function () { if (S.busy) return;
    localStorage.removeItem('nb_flow'); location.href = '/ai-studio.html'; };

  (async function init() {
    renderStepper(); renderOut();
    NB.bindUserChip($('userChip'));
    if (!NB.auth.token()) { $('loginGate').hidden = false; $('btnStart').disabled = true; return; }
    var saved = null;
    try { saved = JSON.parse(localStorage.getItem('nb_flow') || 'null'); } catch (e) {}
    if (saved && saved.flow_id) {
      try {
        var f = await NB.api('/api/ai/flow/' + saved.flow_id);
        S.flowId = f.flow_id; S.idea = f.idea; S.model = f.model; S.mode = f.mode || 'full';
        (f.stages || []).forEach(function (fs) {
          var i = STAGES.findIndex(function (x) { return x.id === fs.id; });
          if (i > -1) { S.stages[i] = Object.assign(S.stages[i], fs); }
        });
        var next = S.stages.findIndex(function (x) { return x.status !== 'done'; });
        $('ideaInput').value = S.idea;
        gotoStage(next === -1 ? STAGES.length - 1 : next);
        renderOut(); NB.toast('已恢复上次的流程');
      } catch (e) { localStorage.removeItem('nb_flow'); }
    }
  })();
})();
