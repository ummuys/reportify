package scripts

const FindLabel = `
(function(){
  var needleRaw = (arguments && arguments.length ? arguments[0] : '');
  var needle = String(needleRaw == null ? '' : needleRaw);
  var targetLC = needle.replace(/\s+/g,' ').trim().toLowerCase();
  if (!targetLC) return 'ERR:NO_LABEL';

  var deadline = Date.now() + 5000; // ждём до 5 сек
  var token = 'goway-' + (Date.now().toString(36) + Math.random().toString(36).slice(2));

  function vis(el){
    if(!el) return false;
    var s = el.ownerDocument.defaultView.getComputedStyle(el);
    if (s.visibility==='hidden' || s.display==='none' || s.pointerEvents==='none') return false;
    var r = el.getBoundingClientRect();
    return r.width>0 && r.height>0;
  }
  function q(doc, sel){ try{ return [].slice.call(doc.querySelectorAll(sel)); }catch(_){ return []; } }
  function norm(t){ return (t||'').replace(/\s+/g,' ').trim().toLowerCase(); }

  function markFrameChainFrom(el){
    var d = el.ownerDocument;
    while (d && d.defaultView && d.defaultView.frameElement){
      var fe = d.defaultView.frameElement;
      try{ fe.setAttribute('data-goway-frame', token); }catch(_){}
      d = fe.ownerDocument;
    }
  }

  function searchOnceIn(doc, stat){
    // 0) пропустить лоадер, если есть
    try{
      var ld = doc.getElementById('loading');
      if (ld && vis(ld)) { stat.loading++; return null; }
    }catch(_){}

    // 1) grid/tree/ячейки — сначала точное совпадение текста
    var cellsSel = '.x-grid-cell-inner,.x-tree-node-text,[role="treeitem"],[role="row"],a,button,span,div,li';
    var cells = q(doc, cellsSel);
    stat.cells += cells.length;
    for (var i=0;i<cells.length;i++){
      var e = cells[i]; if(!vis(e)) continue;
      var t = norm(e.innerText || e.textContent);
      if (t && t === targetLC){ try{ e.setAttribute('data-goway-hit', token); }catch(_){}
        markFrameChainFrom(e); return e; }
    }
    // 2) grid/tree — подстрока
    for (var j=0;j<cells.length;j++){
      var e2 = cells[j]; if(!vis(e2)) continue;
      var t2 = norm(e2.innerText || e2.textContent);
      if (t2 && t2.indexOf(targetLC) >= 0){ try{ e2.setAttribute('data-goway-hit', token); }catch(_){}
        markFrameChainFrom(e2); return e2; }
    }

    // 3) инпуты — точное value
    var fields = q(doc, 'input:not([type="hidden"]),textarea,[role="textbox"],[contenteditable="true"]');
    stat.fields += fields.length;
    for (var k=0;k<fields.length;k++){
      var f = fields[k]; if(!vis(f)) continue;
      var v = norm(f.value || f.getAttribute && f.getAttribute('value'));
      if (v && v === targetLC){ try{ f.setAttribute('data-goway-hit', token); }catch(_){}
        markFrameChainFrom(f); return f; }
    }
    // 4) инпуты — подстрока value
    for (var m=0;m<fields.length;m++){
      var f2 = fields[m]; if(!vis(f2)) continue;
      var v2 = norm(f2.value || f2.getAttribute && f2.getAttribute('value'));
      if (v2 && v2.indexOf(targetLC) >= 0){ try{ f2.setAttribute('data-goway-hit', token); }catch(_){}
        markFrameChainFrom(f2); return f2; }
    }

    // 5) Shadow DOM (первый уровень)
    var hosts = q(doc, '*');
    for (var h=0; h<hosts.length; h++){
      var host = hosts[h];
      if (host.shadowRoot){
        var res = searchOnceIn(host.shadowRoot, stat);
        if (res) return res;
      }
    }

    // 6) iframe
    var ifr = q(doc, 'iframe');
    stat.iframes += ifr.length;
    for (var n=0;n<ifr.length;n++){
      var frame = ifr[n];
      try{
        var d = frame.contentDocument;
        if (d){
          var r = searchOnceIn(d, stat);
          if (r){ try{ frame.setAttribute('data-goway-frame', token); }catch(_){}
            return r; }
        } else {
          stat.xorigin++;
        }
      }catch(e){
        stat.xorigin++;
      }
    }
    return null;
  }

  var stat = {cells:0, fields:0, iframes:0, xorigin:0, loading:0};
  var hit = null;
  while (!hit && Date.now() < deadline){
    hit = searchOnceIn(document, stat);
    if (!hit) { try { var t = setTimeout; } catch(_){ } ; // no-op
      // небольшой микрослип
      var start = Date.now(); while (Date.now()-start < 100) {}
    }
  }
  if (!hit){
    return 'ERR:NF cells='+stat.cells+' fields='+stat.fields+' ifr='+stat.iframes+' xo='+stat.xorigin+' loadingHits='+stat.loading;
  }
  return token;
})();


`
