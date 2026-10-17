/*! nova-dl v1 —— 统一下载 + 编码适配(全站工具共用) */
(function(){
  function safeName(s){
    s=String(s==null?'file':s).trim();
    s=s.replace(/[\\\/:*?"<>|\u0000-\u001f\u007f]/g,'-');
    return s.replace(/^[-.]+/,'').slice(0,80)||'file';
  }
  function dl(data,name,mime){
    try{
      var blob=data instanceof Blob?data:new Blob([data],{type:mime||'application/octet-stream'});
      var url=URL.createObjectURL(blob);
      var a=document.createElement('a');
      a.href=url;a.download=safeName(name);
      a.rel='noopener';a.style.position='fixed';a.style.opacity='0';
      document.body.appendChild(a);
      a.click();
      /* 60 秒后才回收:给下载管理器留足读取时间(治大文件红色感叹号/0 B/s 卡死) */
      setTimeout(function(){URL.revokeObjectURL(url);a.remove()},60000);
      return blob.size;
    }catch(e){
      try{var u=URL.createObjectURL(new Blob([data],{type:mime||'application/octet-stream'}));
        window.open(u,'_blank');return -1}catch(e2){return -1}
    }
  }
  /* 编码适配:BOM 探测 → UTF-8 严格校验 → GBK 回退 */
  function decodeBuf(buf){
    var u8=new Uint8Array(buf);
    if(u8.length>1&&u8[0]===0xFF&&u8[1]===0xFE)return new TextDecoder('utf-16le').decode(u8.subarray(2));
    if(u8.length>1&&u8[0]===0xFE&&u8[1]===0xFF){try{return new TextDecoder('utf-16be').decode(u8.subarray(2))}catch(e){}}
    if(u8.length>2&&u8[0]===0xEF&&u8[1]===0xBB&&u8[2]===0xBF)return new TextDecoder('utf-8').decode(u8.subarray(3));
    try{return new TextDecoder('utf-8',{fatal:true}).decode(u8)}
    catch(e){try{return new TextDecoder('gbk').decode(u8)}catch(e2){return new TextDecoder('utf-8').decode(u8)}}
  }
  window.novaDL=dl;window.novaSafeName=safeName;window.novaDecode=decodeBuf;
})();
