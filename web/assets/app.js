/* Nova Tools 前端 v2 —— 自包含精简版（无模板串、无正则，规避解析歧义） */
(function(){
'use strict';
function $(s){return document.querySelector(s)}
function esc(s){s=String(s==null?'':s);return s.split('&').join('&amp;').split('<').join('&lt;').split('>').join('&gt;').split('"').join('&quot;')}
function box(m){return '<div class="empty-state">'+esc(m)+'</div>'}
function show(j){if(j&&j.ok===false)return box(j.error||j.message||'执行失败');return '<pre class="code">'+esc(JSON.stringify(j,null,2))+'</pre>'}
function isFile(t){return t.kind==='file'||t.type==='file'||t.input==='file'||t.file===true}
function apiList(){
  return fetch('/api/tools').then(function(r){return r.json()}).then(function(j){
    if(j&&j.data&&j.data.tools)return j.data.tools;
    if(j&&j.tools)return j.tools;
    if(Array.isArray(j))return j;
    if(j&&Array.isArray(j.data))return j.data;
    return [];
  });
}
function header(){
  var h='<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:26px">';
  h+='<a href="#/" style="text-decoration:none;font-weight:800;font-size:17px">✳ Nova Tools</a>';
  h+='<div style="display:flex;gap:14px">';
  h+='<a class="muted" href="#/" style="text-decoration:none;font-size:14px">工具箱</a>';
  h+='<a class="muted" href="#/login" style="text-decoration:none;font-size:14px">登录 / 注册</a>';
  h+='</div></div>';
  return h;
}
function home(list){
  var h='<div class="eyebrow">NOVA TOOLS</div>';
  h+='<h1 style="margin:6px 0 4px">工具箱</h1>';
  h+='<p class="muted" style="margin:0 0 22px">共 '+list.length+' 个工具 · 文件类工具仅在服务器内存处理</p>';
  h+='<div class="workgrid">';
  for(var i=0;i<list.length;i++){
    var t=list[i],tid=t.id||t.key||'';
    h+='<a class="panel" style="display:block;text-decoration:none" href="#/tools/'+encodeURIComponent(tid)+'">';
    h+='<strong>'+esc(t.name||tid)+'</strong>';
    h+='<p class="muted" style="margin:6px 0 0;font-size:13px">'+esc(t.desc||t.description||'')+'</p>';
    h+='</a>';
  }
  h+='</div>';
  return h;
}
function detail(t){
  var tid=t.id||t.key||'';
  var h='<a class="secondary" href="#/" style="text-decoration:none;display:inline-block;margin-bottom:14px">&larr; 返回工具箱</a>';
  h+='<div class="panel">';
  h+='<div class="eyebrow">TOOL</div>';
  h+='<h2 style="margin:6px 0 4px">'+esc(t.name||tid)+'</h2>';
  h+='<p class="muted" style="margin:0 0 16px">'+esc(t.desc||t.description||'')+'</p>';
  if(isFile(t)){
    h+='<div class="dropzone" id="dz"><div class="eyebrow" style="color:inherit">点击选择或拖入文件</div>';
    h+='<p class="muted" style="font-size:12.5px;margin:6px 0 0">不超过 10MB · 内存处理不落盘</p>';
    h+='<input type="file" id="fp" hidden></div><div id="res" style="margin-top:14px"></div>';
  }else{
    h+='<textarea id="gin" rows="4" placeholder="输入要处理的内容（可选）"></textarea>';
    h+='<div style="margin-top:14px"><button class="primary" id="runbtn">执行</button></div>';
    h+='<div id="res" style="margin-top:14px"></div>';
  }
  h+='</div>';
  return h;
}
function bindRun(tid){
  var b=$('#runbtn'),ta=$('#gin'),res=$('#res');
  if(!b||!res)return;
  b.addEventListener('click',function(){
    var body={};
    if(ta&&ta.value.trim())body.input=ta.value.trim();
    res.innerHTML=box('执行中…');
    fetch('/api/tools/'+encodeURIComponent(tid)+'/run',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)})
      .then(function(r){return r.json()}).then(function(j){res.innerHTML=show(j)})
      .catch(function(e){res.innerHTML=box('网络错误：'+e.message)});
  });
}
function bindUpload(tid){
  var dz=$('#dz'),fp=$('#fp'),res=$('#res');
  if(!dz||!fp||!res)return;
  function send(){
    var f=fp.files&&fp.files[0];
    if(!f)return;
    if(f.size>10485760){res.innerHTML=box('文件超过 10MB 上限');return}
    res.innerHTML=box('上传处理中…');
    var fd=new FormData();fd.append('file',f);
    fetch('/api/tools/'+encodeURIComponent(tid)+'/run',{method:'POST',body:fd})
      .then(function(r){return r.json()}).then(function(j){res.innerHTML=show(j)})
      .catch(function(e){res.innerHTML=box('网络错误：'+e.message)});
  }
  dz.addEventListener('click',function(){fp.click()});
  fp.addEventListener('change',send);
  dz.addEventListener('dragover',function(e){e.preventDefault();dz.classList.add('drag')});
  dz.addEventListener('dragleave',function(){dz.classList.remove('drag')});
  dz.addEventListener('drop',function(e){e.preventDefault();dz.classList.remove('drag');if(e.dataTransfer&&e.dataTransfer.files.length){fp.files=e.dataTransfer.files;send()}});
}
function authView(){
  var h='<div class="panel" style="max-width:420px;margin:24px auto 0">';
  h+='<div class="eyebrow">ACCOUNT</div><h2 style="margin:6px 0 16px">登录 / 注册</h2>';
  h+='<input type="text" id="au" placeholder="用户名" style="margin-bottom:10px">';
  h+='<input type="password" id="ap" placeholder="密码" style="margin-bottom:14px">';
  h+='<div style="display:flex;gap:10px"><button class="primary" id="lg">登录</button><button id="rg">注册</button></div>';
  h+='<div id="ares" style="margin-top:14px"></div></div>';
  return h;
}
function bindAuth(){
  var a=$('#ares');if(!a)return;
  function go(path){
    var u=$('#au').value.trim(),p=$('#ap').value;
    if(!u||!p){a.innerHTML=box('请输入用户名和密码');return}
    a.innerHTML=box('提交中…');
    fetch(path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:u,password:p})})
      .then(function(r){return r.json()}).then(function(j){
        if(j&&j.ok===false){a.innerHTML=box(j.error||j.message||'失败');return}
        try{localStorage.setItem('nova.auth',JSON.stringify(j||{}))}catch(e){}
        a.innerHTML=box('成功，正在返回首页…');
        setTimeout(function(){location.hash='#/'},800);
      }).catch(function(e){a.innerHTML=box('网络错误：'+e.message)});
  }
  var lg=$('#lg'),rg=$('#rg');
  if(lg)lg.addEventListener('click',function(){go('/api/login')});
  if(rg)rg.addEventListener('click',function(){go('/api/register')});
}
function render(){
  var app=$('#app');
  if(!app)return;
  var h=location.hash||'';
  app.innerHTML=header()+'<div class="muted" style="text-align:center;padding:40px 0">加载中…</div>';
  if(h==='#/login'){app.innerHTML=header()+authView();bindAuth();return}
  if(h.indexOf('#/tools/')===0){
    var id=decodeURIComponent(h.slice(8));
    apiList().then(function(list){
      var t=null;
      for(var i=0;i<list.length;i++){if((list[i].id||list[i].key)===id)t=list[i]}
      if(!t){app.innerHTML=header()+'<div class="panel">'+box('工具不存在或未开启')+'</div>';return}
      app.innerHTML=header()+detail(t);
      if(isFile(t))bindUpload(t.id||t.key);else bindRun(t.id||t.key);
    }).catch(function(e){app.innerHTML=header()+'<div class="panel">'+box('加载失败：'+e.message)+'</div>'});
    return;
  }
  apiList().then(function(list){app.innerHTML=header()+home(list)})
    .catch(function(e){app.innerHTML=header()+'<div class="panel">'+box('加载失败：'+e.message)+'</div>'});
}
window.addEventListener('hashchange',render);
render();
})();
