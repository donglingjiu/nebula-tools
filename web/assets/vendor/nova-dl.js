/*! nova-dl v2 —— 统一保存层:文件名纠偏守卫 + revoke 缓刑 + FSA 另存为优先 + 保活兜底 */
(function(){
  /* 1) 乱码纠偏:UTF-8 字节被当 Latin-1 显示的字符串,反解回中文;解不动原样返回 */
  function fixMojibake(s){
    s=String(s==null?'':s);
    if(!/[\u0080-\u00ff]/.test(s))return s;
    try{
      var n=s.length,b=new Uint8Array(n),i,c;
      for(i=0;i<n;i++){c=s.charCodeAt(i);if(c>255)return s;b[i]=c}
      return new TextDecoder('utf-8',{fatal:true}).decode(b);
    }catch(e){return s}
  }
  function safeName(s){
    s=fixMojibake(s).trim();
    s=s.replace(/[\\\/:*?"<>|\u0000-\u001f\u007f]/g,'-');
    return s.replace(/^[-.]+/,'').slice(0,80)||'file';
  }
  /* 2) 兜底通道:a[download] + 60 秒保活(给下载管理器留足读取时间) */
  function legacy(data,name,mime){
    try{
      var blob=data instanceof Blob?data:new Blob([data],{type:mime||'application/octet-stream'});
      var url=URL.createObjectURL(blob);
      var a=document.createElement('a');
      a.href=url;a.download=safeName(name);a.rel='noopener';
      a.style.position='fixed';a.style.opacity='0';
      document.body.appendChild(a);a.click();
      setTimeout(function(){URL.revokeObjectURL(url);a.remove()},60000);
      return blob.size;
    }catch(e){return -1}
  }
  /* 3) FSA 另存为(https/localhost 生效):系统对话框,完全不经过下载管理器 */
  var MIME={pdf:'application/pdf',
    docx:'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    xlsx:'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    csv:'text/csv',svg:'image/svg+xml',png:'image/png',jpg:'image/jpeg',jpeg:'image/jpeg',
    webp:'image/webp',zip:'application/zip',txt:'text/plain',md:'text/markdown'};
  async function save(data,name,mime){
    name=safeName(name);
    var blob=data instanceof Blob?data:new Blob([data],{type:mime||'application/octet-stream'});
    if(window.showSaveFilePicker){
      try{
        var ext=(name.match(/\.([a-z0-9]{1,5})$/i)||[])[1]||'';
        var opts={suggestedName:name};
        if(ext){var m=MIME[ext.toLowerCase()]||blob.type||'application/octet-stream';
          var acc={};acc[m]=['.'+ext.toLowerCase()];
          opts.types=[{description:ext.toUpperCase()+' 文件',accept:acc}]}
        var h=await window.showSaveFilePicker(opts);
        var w=await h.createWritable();await w.write(blob);await w.close();
        return blob.size;
      }catch(e){if(e&&e.name==='AbortError')return 0}
    }
    return legacy(blob,name,mime);
  }
  /* 4a) 全局文件名守卫:任何 a[download] 点击都过 safeName(app.js 自有管线也治) */
  try{
    var oc=HTMLAnchorElement.prototype.click;
    HTMLAnchorElement.prototype.click=function(){
      if(this.download)this.download=safeName(this.download);
      return oc.apply(this,arguments);
    };
  }catch(e){}
  /* 4b) 全局 revoke 缓刑:任何 objectURL 回收延迟 60s,根绝「等待中 0B」类卡死 */
  try{
    var orv=URL.revokeObjectURL.bind(URL);
    URL.revokeObjectURL=function(u){setTimeout(function(){try{orv(u)}catch(e){}},60000)};
  }catch(e){}
  /* 5) 文本编码适配:BOM 探测 → UTF-8 严格 → GBK 回退(excel 等依赖) */
  function decodeBuf(buf){
    var u8=new Uint8Array(buf);
    if(u8.length>1&&u8[0]===0xFF&&u8[1]===0xFE)return new TextDecoder('utf-16le').decode(u8.subarray(2));
    if(u8.length>1&&u8[0]===0xFE&&u8[1]===0xFF){try{return new TextDecoder('utf-16be').decode(u8.subarray(2))}catch(e){}}
    if(u8.length>2&&u8[0]===0xEF&&u8[1]===0xBB&&u8[2]===0xBF)return new TextDecoder('utf-8').decode(u8.subarray(3));
    try{return new TextDecoder('utf-8',{fatal:true}).decode(u8)}
    catch(e){try{return new TextDecoder('gbk').decode(u8)}catch(e2){return new TextDecoder('utf-8').decode(u8)}}
  }
  window.novaDL=save;window.novaSave=save;window.novaSafeName=safeName;window.novaDecode=decodeBuf;
})();
