/* 后台共享：adminGate 权限门 / adminNav 导航高亮 */
(function(){
  var NB = window.NB;
  NB.adminGate = function(cb){
    NB.auth.me().then(function(u){
      if(!u){ location.href = '/login.html?next=' + encodeURIComponent(location.pathname); return; }
      if(!NB.auth.isAdmin(u)){
        var m = document.querySelector('.nb-main');
        if(m) m.innerHTML = '<div class="nb-gate"><h2>403</h2><p>该区域需要管理员权限。</p>' +
          '<a class="nb-btn nb-btn--ghost nb-btn--sm" href="/">返回首页</a></div>';
        return;
      }
      cb && cb();
    });
  };
  NB.adminNav = function(){
    var p = location.pathname.split('/').pop() || 'admin.html';
    var links = document.querySelectorAll('.nb-side a');
    for(var i = 0; i < links.length; i++){
      if(links[i].getAttribute('href') === '/' + p) links[i].classList.add('on');
    }
  };
})();
