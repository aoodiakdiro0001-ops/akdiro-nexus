/* AKDIRO-NEXUS UI — Baran © 2026 — تعامل جاسم مع نقاط نهايات API */
const $ = (s)=>document.querySelector(s);
const esc = (s)=>String(s??'').replace(/[&<>"']/g, c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));

async function api(p,o){
  const r = await fetch(p,o);
  if(!r.ok) throw new Error((await r.text()).slice(0,150));
  return r.json();
}
function toast(msg){
  const t=$('#toast'); t.textContent=msg; t.style.display='block';
  clearTimeout(t._h); t._h=setTimeout(()=>t.style.display='none',2800);
}

async function loadScope(){
  try{
    const s = await api('/api/scope');
    const items = (s.domains||[]).map(d=>
      `<span class="pill p2">${esc(d)}</span>`).join(' ') || '<span class="muted">لا نطاقات متكفَّع بها — أضف بتعهد</span>';
    $('#honestyBox').insertAdjacentHTML('afterend',
      `<div class="sidecard" id="scopeListed"><b>✅ النطاقات المتكفَّع بها</b><p>${items}</p></div>`);
    $('#scopeListed').id='';
  }catch(e){ toast('⛔ '+e.message); }
}

window.addScope = async () => {
  const dom = $('#scopeDom').value.trim();
  if(!dom){ toast('اكتب النطاق أولًا'); return; }
  if(!$('#attest').checked){ toast('⛔ التعهد مطلوب — الأداة تعمل بالإذن لا بالغموض'); return; }
  try{
    await api('/api/scope',{method:'POST',headers:{'Content-Type':'application/json'},
      body: JSON.stringify({domains:[dom], phrase:"أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب"})});
    toast('أُضيف للنصاب ✅ (التعهد مسجل)'); location.reload();
  }catch(e){ toast('⛔ '+e.message); }
};

window.probe = async () => {
  const u = $('#probeUrl').value.trim();
  if(!u){ toast('اكتب رابطًا أولًا'); return; }
  $('#probeBox').innerHTML='<em>جارٍ الخديل الاستكشافي…</em>';
  try{
    const r = await api('/api/profile?url='+encodeURIComponent(u));
    const cls='p'+r.risk_level;
    $('#probeBox').innerHTML = `
      <div><span class="pill ${cls}">درجة الخطر ${r.risk_level}/5</span> — يُقَترح: <b>${esc(r.recommend_tier)}</b></div>
      <div class="muted" style="margin-top:6px">الحالة ${r.status} · ${esc(r.server||'خادم غير معلن')} · زمن ${r.latency_ms}ms</div>
      <div class="muted">مؤشرات: ${r.markers&&r.markers.length? r.markers.map(m=>esc(m)).join('، ') : 'لا مؤشرات دفاعية مرصدة'}</div>
      <div class="muted" style="margin-top:6px">${esc(r.honesty_note)}</div>`;
  }catch(e){
    $('#probeBox').innerHTML = `<div class="muted" style="color:var(--bad)">${esc(e.message)}</div>`;
  }
};

async function loadCaps(){
  try{
    const c = await api('/api/capabilities');
    $('#capsBox').innerHTML =
      `TLS بصمة: ${c.tls_fingerprint?'✅':'⛔'} · h2: ${c.h2_fingerprint?'✅':'⛔'} · كابشا: ${c.captcha_solver?'✅':'⛔ (نهائيًا)'}<br>
       <div class="muted">${esc(c.honesty_note)}</div>`;
  }catch{ /* تجاهل */ }
}

window.egress = async () => {
  try{
    const e = await api('/api/egress');
    $('#egressBox').innerHTML = e.error? `⛔ ${esc(e.error)}`
      : `IP الخارجي: <b>${esc(e.ip)}</b> · الطبقة: ${esc(e.tier)} · ${e.latency_ms}ms`;
  }catch(e){ toast('⛔ '+e.message); }
};

async function loadFindings(){
  try{
    const f = await api('/api/findings');
    $('#findBody').innerHTML = (f.findings||[]).slice(0,80).map(x=>
      `<tr><td>${esc(x.target)}</td><td>${esc(x.kind)}</td><td><span class="pill p${x.severity}">${x.severity}</span></td>
       <td>${esc(x.detail).slice(0,90)}</td><td>${esc(x.source||'')}</td><td class="muted">${new Date(x.ts).toLocaleString('ar-EG',{hour12:false})}</td></tr>`
    ).join('') || '<tr><td colspan="6" class="muted">لا ثغرات مسجلة بعد</td></tr>';
  }catch{ /* تجاهل */ }
}

async function loadLicense(){
  try{
    const l = await api('/api/license');
    $('#licBox').innerHTML = `<b>${esc(l.product)}</b> — ${esc(l.license)} <span class="pill p3" style="background:#1c3428;color:var(--ok)">مفعّل</span>
    <div class="muted">المالك: ${esc(l.owner)} · المفتاح: <code dir="ltr">${esc(l.key)}</code> · الدعم: <a class="muted" href="mailto:${esc(l.support.email)}">${esc(l.support.email)}</a> · ${esc(l.support.phone)}</div>`;
  }catch{ /* تجاهل */ }
}

loadScope(); loadCaps(); loadLicense(); loadFindings();
setInterval(loadFindings, 5000);
