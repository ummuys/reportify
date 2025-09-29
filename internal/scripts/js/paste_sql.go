package scripts

const PasteSql = `
(function(sql){
  function isVisible(el){
    if (!el) return false;
    var s = getComputedStyle(el);
    if (s.visibility === 'hidden' || s.display === 'none') return false;
    var r = el.getBoundingClientRect();
    return r.width > 0 && r.height > 0;
  }

  function isSqlEditor(cm){
    try{
      var m = cm && cm.getOption && cm.getOption('mode');
      if (!m) return false;
      if (typeof m === 'string') return /sql|postgres|mysql|mariadb|ms\-?sql/i.test(m);
      if (m && m.name)    return /sql/i.test(m.name);
    }catch(e){}
    return false;
  }

  function findEditorsInDoc(doc, out){
    out = out || [];
    try{
      var wrappers = doc.querySelectorAll('.CodeMirror');
      for (var i=0;i<wrappers.length;i++){
        var w = wrappers[i];
        var cm = w.CodeMirror || w.cm;
        if (!cm){
          // на всякий случай поищем объект-редактор среди свойств
          for (var k in w){
            var v;
            try{ v = w[k]; }catch(e){ v = null; }
            if (v && typeof v.setValue==='function' && typeof v.focus==='function'){ cm = v; break; }
          }
        }
        if (cm && typeof cm.setValue==='function'){
          out.push({cm: cm, wrapper: w});
        }
      }
      // рекурсивно в те же-доменные iframe
      var frames = doc.querySelectorAll('iframe');
      for (var j=0;j<frames.length;j++){
        var f = frames[j];
        try{
          if (f.contentDocument) findEditorsInDoc(f.contentDocument, out);
        }catch(e){}
      }
    }catch(e){}
    return out;
  }

  function pasteInto(cm, value){
    try{
      if (cm.getOption && cm.getOption('readOnly')) cm.setOption('readOnly', false);
      if (cm.execCommand) cm.execCommand('selectAll');
      cm.setValue(value || '');
      if (cm.refresh) cm.refresh();
      if (cm.focus) cm.focus();
      return true;
    }catch(e){ return false; }
  }

  var editors = findEditorsInDoc(document, []);

  // 1) сначала ищем видимый SQL-редактор
  var target = null;
  for (var i=0;i<editors.length;i++){
    if (isVisible(editors[i].wrapper) && isSqlEditor(editors[i].cm)){ target = editors[i].cm; break; }
  }
  // 2) затем любой видимый редактор
  if (!target){
    for (var i2=0;i2<editors.length;i2++){
      if (isVisible(editors[i2].wrapper)){ target = editors[i2].cm; break; }
    }
  }
  // 3) если ничего видимого — берём первый попавшийся
  if (!target && editors.length){ target = editors[0].cm; }

  if (target){
    var ok = pasteInto(target, sql);
    return ok ? 'OK' : 'ERR_SET';
  }

  // запасной вариант: видимое textarea
  var tas = document.querySelectorAll('textarea');
  for (var t=0;t<tas.length;t++){
    var ta = tas[t];
    if (!isVisible(ta)) continue;
    try{
      ta.value = sql;
      ta.dispatchEvent(new Event('input', {bubbles:true}));
      ta.dispatchEvent(new Event('change', {bubbles:true}));
      ta.focus();
      return 'OK_TA';
    }catch(e){}
  }

  return 'NO_EDITOR';
})(arguments[0]);`
