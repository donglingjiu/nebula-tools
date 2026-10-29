/* AI 对话 —— 接真 LLM
   ============ 前后端契约（阶段5按此实现）============
   POST /api/ai/chat {message, history:[{role,content} 最多20条]} -> SSE:
     {type:"delta",text} {type:"usage",tokens_in,tokens_out} {type:"done"}
   错误码：401 未登录 / 402 付费门 / 410 功能下线 / 429 限流（core 统一处理）
   ==================================================== */
(function(){
  'use strict';
  var NB = window.NB;
  var $ = function(id){ return document.getElementById(id); };
  var S = { msgs: [], busy: false, tin: 0, tout: 0 };

  function load(){ try { S.msgs = JSON.parse(localStorage.getItem('nb_chat') || '[]'); } catch(e){ S.msgs = []; } }
  function save(){ try { localStorage.setItem('nb_chat', JSON.stringify(S.msgs.slice(-50))); } catch(e){} }

  function bubble(role, text){
    var d = document.createElement('div');
    d.className = 'nb-msg nb-msg--' + (role === 'user' ? 'u' : 'a');
    d.textContent = text;
    $('log').appendChild(d);
    $('log').scrollTop = 1e9;
    return d;
  }
  function renderAll(){ $('log').innerHTML = ''; S.msgs.forEach(function(m){ bubble(m.role, m.content); }); }
  function setStatus(){ $('stTok').textContent = (S.tin + S.tout).toLocaleString('zh-CN'); }

  $('btnSend').onclick = async function(){
    var txt = $('input').value.trim();
    if(!txt || S.busy) return;
    if(!NB.auth.token()){
      NB.modal({
        title: '需要登录',
        body: 'AI 对话需登录后使用（免费工具无需登录）。',
        actions: [
          { label: '去登录', cls: 'nb-btn--warm', onClick: function(){ location.href = '/login.html?next=/ai-chat.html'; } },
          { label: '暂不' }
        ]
      });
      return;
    }
    $('input').value = '';
    S.msgs.push({ role: 'user', content: txt });
    bubble('user', txt); save();
    var b = bubble('assistant', '…');
    S.busy = true; $('btnSend').disabled = true;
    var acc = '';
    try {
      await NB.stream('/api/ai/chat', { message: txt, history: S.msgs.slice(-21, -1) }, {
        onEvent: function(ev){
          if(ev.type === 'delta'){ acc += ev.text || ''; b.textContent = acc; $('log').scrollTop = 1e9; }
          else if(ev.type === 'usage'){ S.tin += ev.tokens_in || 0; S.tout += ev.tokens_out || 0; setStatus(); }
        },
        onError: function(e){ b.textContent = '失败：' + e.message; },
        onDone: function(){
          if(acc){ S.msgs.push({ role: 'assistant', content: acc }); save(); }
          else if(b.textContent === '…'){ b.textContent = '（无返回内容）'; }
          S.busy = false; $('btnSend').disabled = false;
        }
      });
    } catch(e){ S.busy = false; $('btnSend').disabled = false; }
  };
  $('input').addEventListener('keydown', function(e){
    if(e.key === 'Enter' && !e.shiftKey){ e.preventDefault(); $('btnSend').click(); }
  });
  $('btnNew').onclick = function(){
    if(S.busy) return;
    S.msgs = []; S.tin = 0; S.tout = 0;
    save(); $('log').innerHTML = ''; setStatus();
  };

  (function init(){
    load(); renderAll(); setStatus();
    NB.bindUserChip($('userChip'));
  })();
})();
