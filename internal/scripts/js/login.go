package scripts

const Login = `
(function(u,p){
  function visible(el){
    if(!el) return false;
    var s = window.getComputedStyle(el);
    var r = el.getBoundingClientRect();
    return s.display!=='none' && s.visibility!=='hidden' && r.width>0 && r.height>0;
  }
  function setVal(el, txt){
    if(!el) return false;
    el.focus();
    // Если contenteditable/role="textbox"
    if (el.isContentEditable || el.getAttribute('role')==='textbox'){
      // очистка
      el.textContent = '';
      el.dispatchEvent(new Event('input',{bubbles:true}));
      // вставим как ввод текста
      try { document.execCommand('insertText', false, txt); } catch(e){ el.textContent = txt; }
      el.dispatchEvent(new Event('input',{bubbles:true}));
      el.dispatchEvent(new Event('change',{bubbles:true}));
      return true;
    }
    // Обычный input: используем "native setter", чтобы триггерить реактивность
    var proto = el.__proto__ || Object.getPrototypeOf(el);
    var desc = Object.getOwnPropertyDescriptor(proto, 'value');
    if (desc && desc.set){
      desc.set.call(el, txt);
    } else {
      el.value = txt;
    }
    el.dispatchEvent(new Event('input',{bubbles:true}));
    el.dispatchEvent(new Event('change',{bubbles:true}));
    return true;
  }

  var pwd = document.querySelector("input[type='password'], input[name='password'], #password");
  if(!pwd || !visible(pwd)) return "NO_PASSWORD";

  // Кандидаты на логин
  var selectors = [
    "input[autocomplete='username']",
    "input[name='username']",
    "input[id*='user' i]",
    "input[name*='user' i]",
    "input[type='email']",
    "input[type='text']",
    "input:not([type])",
    "textarea",
    "[role='textbox']",
    "[contenteditable='true']",
    "[contenteditable='']"
  ];

  // Ищем ближайший видимый к паролю (выше по экрану)
  var login = null;
  var pwdTop = pwd.getBoundingClientRect().top;


  function pick(cands){
    var best = null, bestDist = 1e9;
    for (var i=0;i<cands.length;i++){
      var el = cands[i];
      if(!el || el===pwd) continue;
      // игнорим password/hidden/disabled
      var t = (el.getAttribute('type')||'').toLowerCase();
      if (t==='password' || el.disabled || el.readOnly) continue;
      if (!visible(el)) continue;
      var r = el.getBoundingClientRect();
      var dist = (pwdTop - r.top >= -5) ? Math.abs(pwdTop - r.top) : Math.abs(r.top - pwdTop) + 1000; // выше — приоритетнее
      if (dist < bestDist){ bestDist = dist; best = el; }
    }
    return best;
  }

  // Сначала — ищем в пределах одного контейнера
  var root = pwd.closest("form, .dialog, .window, .gwt-DialogBox, body") || document;
  var cands = [];
  for (var s of selectors){
    root.querySelectorAll(s).forEach(function(el){ cands.push(el); });
  }
  login = pick(cands);

  // Если не нашли — ищем по всему документу
  if(!login){
    cands = [];
    for (var s of selectors){
      document.querySelectorAll(s).forEach(function(el){ cands.push(el); });
    }
    login = pick(cands);
  }

  if(!login) return "NO_LOGIN";

  // Заполняем
  setVal(login, u);
  setVal(pwd, p);

  // Ищем кнопку входа
  function findBtn(){
    var texts = ["login","sign in","log in","войти","anmelden","entrar","acceder"];
    var all = Array.from(document.querySelectorAll("button, input[type='submit'], [role='button'], .x-btn, .btn"));
    // по тексту
    for (var el of all){
      if (!visible(el)) continue;
      var t = (el.innerText || el.value || el.getAttribute('aria-label') || "").toLowerCase();
      for (var k of texts){ if (t.indexOf(k)>=0) return el; }
    }
    // хоть что-то кликабельное
    return all.find(visible) || null;
  }
  var btn = findBtn();
  if (btn){ btn.focus(); btn.click(); return "OK_CLICK"; }

  // если нет кнопки — жмём Enter в пароле
  pwd.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',keyCode:13,which:13,bubbles:true}));
  pwd.dispatchEvent(new KeyboardEvent('keyup',{key:'Enter',keyCode:13,which:13,bubbles:true}));
  return "OK_ENTER";
})(arguments[0], arguments[1]);
`
