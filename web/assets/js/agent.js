/* Agent 工具代理 —— cookie 会话 + POST /api/agent/run (JSON 非流式)
   契约: {input, maxRounds, conversationId} -> {ok, conversationId, rounds, calls[{tool,input,output,ok,ms}], answer, profile, tokens, ms}
   错误: 401 未登录 / 422 无可用档案 / 502 全档案失败 */
(function(){
'use strict';
var NB = window.NB;
var $ = function(id){ return document.getElementById(id); };
var S = { convId: 0, busy: false, tokens: 0 };

function load(){ try { S.convId = parseInt(localStorage.getItem('nb_agent_conv')||'0',10)||0; } catch(e){} }
function save(){ try { localStorage.setItem('nb_agent_conv', String(S.convId)); } catch(e){} }
function ensureMeta(){
  var m = $('stMeta');
  if(!m){ m = document.createElement('span'); m.id = 'stMeta'; m.style.cssText = 'margin-left:10px;opacity:.75';
    var t = $('stTok'); if(t && t.parentNode){ t.parentNode.insertBefore(m, t.nextSibling); } }
  return m;
}
function bubble(role, text){
  var d = document.createElement('div');
  d.className = 'nb-msg nb-msg--' + (role==='user'?'u':'a');
  d.textContent = text;
  $('log').appendChild(d); $('log').scrollTop = 1e9;
  return d;
}
function callCard(c){
  var d = document.createElement('div');
  d.className = 'nb-agent-call' + (c.ok ? '' : ' bad');
  var h = document.createElement('div');
  h.className = 'nb-agent-call-h';
  h.textContent = (c.ok?'✓ ':'✗ ') + c.tool + ' · ' + (c.ms||0) + 'ms';
  var p = document.createElement('pre');
  p.className = 'nb-agent-call-o';
  p.textContent = c.output || '';
  d.appendChild(h); d.appendChild(p);
  $('log').appendChild(d); $('log').scrollTop = 1e9;
}
function needLogin(){
  if(NB && NB.modal){ NB.modal({ title:'需要登录', body:'Agent 需登录后使用。',
    actions:[{label:'去登录',cls:'nb-btn--warm',onClick:function(){ location.href='/login.html?next=/agent.html'; }},{label:'暂不'}] }); }
  else { location.href = '/login.html?next=/agent.html'; }
}

$('btnSend').onclick = function(){
  var txt = $('input').value.trim();
  if(!txt || S.busy) return;
  S.busy = true; $('btnSend').disabled = true; $('input').value = '';
  bubble('user', txt);
  var b = bubble('assistant', '…思考中(最多4轮工具调用)');
  fetch('/api/agent/run', {
    method:'POST', credentials:'same-origin',
    headers:{'Content-Type':'application/json'},
    body: JSON.stringify({ input: txt, maxRounds: 4, conversationId: S.convId })
  }).then(function(r){ return r.json().then(function(j){ return {status:r.status, j:j||{}}; }); })
    .then(function(x){
      if(b.parentNode) b.remove();
      if(x.status === 401){ needLogin(); return; }
      if(x.j && x.j.ok){
        S.convId = x.j.conversationId || 0; save();
        (x.j.calls || []).forEach(callCard);
        bubble('assistant', x.j.answer || '(无返回内容)');
        S.tokens += x.j.tokens || 0;
        $('stTok').textContent = S.tokens.toLocaleString('zh-CN');
        ensureMeta().textContent = '档案 '+(x.j.profile||'-')+' · '+(x.j.rounds||0)+'轮 · '+(x.j.ms||0)+'ms';
      } else {
        bubble('assistant', '失败：' + ((x.j && x.j.error) || ('HTTP '+x.status)));
      }
    }).catch(function(e){
      if(b.parentNode) b.remove();
      bubble('assistant', '网络错误：' + e.message);
    }).then(function(){ S.busy = false; $('btnSend').disabled = false; });
};
$('input').addEventListener('keydown', function(e){
  if(e.key === 'Enter' && !e.shiftKey){ e.preventDefault(); $('btnSend').click(); }
});
$('btnNew').onclick = function(){
  if(S.busy) return;
  S.convId = 0; S.tokens = 0; save();
  $('log').innerHTML = ''; $('stTok').textContent = '0'; ensureMeta().textContent = '';
};
(function init(){
  load(); ensureMeta();
  $('stTok').textContent = '0';
  if(NB && NB.bindUserChip) NB.bindUserChip($('userChip'));
})();
})();
