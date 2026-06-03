/* ── Cipher Scope Dashboard — app.js ── */

const CIRCUMFERENCE = 471.24; // 2 * π * 75

// ── State ─────────────────────────────────────────────────────────────────────
let genHashType   = 'md5';
let crackHashType = 'md5';
let attackMode    = 'auto';
let encodeFormat  = 'base64';
let isRunning     = false;
let evtSource     = null;
let genDebounce   = null;
let maxWorkers    = navigator.hardwareConcurrency || 4;

// Attack metadata for log export
let lastAttackMeta = {};

// Chrome data for CSV export
let chromeEntries = [];
let chromeFile    = null;

// ── Element refs ──────────────────────────────────────────────────────────────
const $ = id => document.getElementById(id);

const genPassword    = $('gen-password');
const genHashValue   = $('gen-hash-value');
const genCopyBtn     = $('gen-copy-btn');
const pwdLenBadge    = $('pwd-len-badge');
const crackHashEl    = $('crack-hash');
const crackLenEl     = $('crack-length');
const lengthDisplay  = $('length-display');
const crackWorkersEl = $('crack-workers');
const workersDisplay = $('workers-display');
const bfOptions      = $('bf-options');
const crackBtn       = $('crack-btn');
const ringCircle     = $('ring-circle');
const ringPercent    = $('ring-percent');
const ringPhase      = $('ring-phase');
const statusBadge    = $('status-badge');
const statusText     = $('status-text');
const workersDots    = $('workers-dots');
const workersLabel   = $('workers-label');
const statHashrate   = $('stat-hashrate');
const statEta        = $('stat-eta');
const statElapsed    = $('stat-elapsed');
const statProgress   = $('stat-progress');
const statTotalLbl   = $('stat-total-label');
const resultBanner   = $('result-banner');
const resultLabel    = $('result-label');
const resultValue    = $('result-value');
const resultIcon     = $('result-icon');
const logBody        = $('log-body');

// ── Init ──────────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  initNavTabs();
  initWorkers();
  initTabs();
  initSliders();
  initGenerator();
  initCopyBtn();
  initCrackBtn();
  initLogClear();
  initLogExport();
  initEncodeTool();
  initIdentifyTool();
  initChromeTool();
  connectSSE();
  fetchInfo();
  log('info', 'Cipher Scope v3.0 iniciado — Swiss Army Knife 🔐');
});

// ── Nav Tabs ──────────────────────────────────────────────────────────────────
function initNavTabs() {
  document.querySelectorAll('.nav-tab').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
      document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
      btn.classList.add('active');
      $(btn.dataset.page).classList.add('active');
    });
  });
}

function initWorkers() {
  crackWorkersEl.max = maxWorkers;
  crackWorkersEl.value = maxWorkers;
  workersDisplay.textContent = maxWorkers;
  renderWorkerDots(maxWorkers, maxWorkers);
}

// ── Tabs ──────────────────────────────────────────────────────────────────────
function initTabs() {
  // Generator type tabs
  $('gen-type-tabs').querySelectorAll('.tab').forEach(btn => {
    btn.addEventListener('click', () => {
      $('gen-type-tabs').querySelectorAll('.tab').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      genHashType = btn.dataset.type;
      triggerGenHash();
    });
  });

  // Crack type tabs
  $('crack-type-tabs').querySelectorAll('.tab').forEach(btn => {
    btn.addEventListener('click', () => {
      $('crack-type-tabs').querySelectorAll('.tab').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      crackHashType = btn.dataset.type;
    });
  });

  // Mode tabs
  $('mode-tabs').querySelectorAll('.mode-tab').forEach(btn => {
    btn.addEventListener('click', () => {
      $('mode-tabs').querySelectorAll('.mode-tab').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      attackMode = btn.dataset.mode;
      const showBf   = attackMode === 'bruteforce' || attackMode === 'auto';
      const showJohn = attackMode === 'john';
      bfOptions.classList.toggle('hidden', !showBf);
      $('john-status-badge').style.display  = showJohn ? 'block' : 'none';
      $('john-rules-label').style.display   = showJohn ? 'flex'  : 'none';
    });
  });
}

// ── Sliders ───────────────────────────────────────────────────────────────────
function initSliders() {
  crackLenEl.addEventListener('input', () => {
    lengthDisplay.textContent = crackLenEl.value;
  });

  crackWorkersEl.addEventListener('input', () => {
    const n = parseInt(crackWorkersEl.value);
    workersDisplay.textContent = n;
    renderWorkerDots(n, maxWorkers);
  });
}

function renderWorkerDots(active, total) {
  workersDots.innerHTML = '';
  const show = Math.min(total, 8);
  for (let i = 0; i < show; i++) {
    const d = document.createElement('div');
    d.className = 'w-dot' + (i < active ? ' active' : '');
    workersDots.appendChild(d);
  }
  workersLabel.textContent = `${active} worker${active !== 1 ? 's' : ''}`;
}

// ── Hash Generator ────────────────────────────────────────────────────────────
function initGenerator() {
  genPassword.addEventListener('input', () => {
    // Password length counter
    const len = genPassword.value.length;
    if (len > 0) {
      pwdLenBadge.textContent = `${len} char${len !== 1 ? 's' : ''}`;
      pwdLenBadge.style.opacity = '1';
    } else {
      pwdLenBadge.style.opacity = '0';
    }

    clearTimeout(genDebounce);
    genDebounce = setTimeout(triggerGenHash, 200);
  });
}

async function triggerGenHash() {
  const pwd = genPassword.value;
  if (!pwd) { genHashValue.textContent = '—'; return; }

  try {
    const res = await fetch('/api/hash', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: pwd, type: genHashType }),
    });
    const data = await res.json();
    genHashValue.textContent = data.hash || '—';
  } catch {
    genHashValue.textContent = 'erro';
  }
}

function initCopyBtn() {
  genCopyBtn.addEventListener('click', () => {
    const hash = genHashValue.textContent;
    if (!hash || hash === '—') return;
    navigator.clipboard.writeText(hash).then(() => {
      genCopyBtn.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg> Copiado!`;
      genCopyBtn.classList.add('copied');
      setTimeout(() => {
        genCopyBtn.classList.remove('copied');
        genCopyBtn.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none"><rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" stroke-width="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="currentColor" stroke-width="2"/></svg> Copiar`;
      }, 2000);
    });
  });
}

// ── Hash Auto-Identify on paste ───────────────────────────────────────────────
crackHashEl && crackHashEl.addEventListener('input', () => {
  const h = crackHashEl.value.trim();
  if (!h) { $('hash-guess').innerHTML = ''; return; }
  autoIdentifyHash(h);
});

async function autoIdentifyHash(hash) {
  try {
    const res  = await fetch('/api/identify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ hash }),
    });
    const data = await res.json();
    const guessEl = $('hash-guess');
    guessEl.innerHTML = '';
    if (data.candidates && data.candidates.length > 0) {
      data.candidates.forEach(type => {
        const span = document.createElement('span');
        span.className = 'hash-guess-item';
        span.textContent = type.toUpperCase();
        span.title = 'Clique para selecionar';
        span.style.cursor = 'pointer';
        span.addEventListener('click', () => {
          $('crack-type-tabs').querySelectorAll('.tab').forEach(b => {
            b.classList.toggle('active', b.dataset.type === type);
          });
          crackHashType = type;
          log('info', `Tipo de hash selecionado automaticamente: ${type.toUpperCase()}`);
        });
        guessEl.appendChild(span);
      });
      // Auto-select first candidate
      if (data.candidates.length === 1) {
        const t = data.candidates[0];
        $('crack-type-tabs').querySelectorAll('.tab').forEach(b => {
          b.classList.toggle('active', b.dataset.type === t);
        });
        crackHashType = t;
      }
    }
  } catch { /* silencioso */ }
}

// ── Crack Button ──────────────────────────────────────────────────────────────
function initCrackBtn() {
  crackBtn.addEventListener('click', async () => {
    if (isRunning) {
      await stopAttack();
      return;
    }
    const hash = crackHashEl.value.trim();
    if (!hash) { log('error', 'Cole um hash antes de iniciar.'); return; }
    await startAttack(hash);
  });
}

async function startAttack(hash) {
  const charset = buildCharset();
  if (!charset && (attackMode === 'bruteforce' || attackMode === 'auto')) {
    log('error', 'Selecione ao menos um charset.');
    return;
  }

  const body = {
    hash,
    hashType:  crackHashType,
    mode:      attackMode,
    workers:   parseInt(crackWorkersEl.value),
    maxLength: parseInt(crackLenEl.value),
    charset,
    wordlist:  'wordlist.txt',
    johnRules: $('john-rules-check') ? $('john-rules-check').checked : false,
  };

  lastAttackMeta = { hash, hashType: crackHashType, mode: attackMode, startedAt: new Date().toISOString() };

  try {
    const res = await fetch('/api/crack', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!res.ok) { log('error', 'Erro ao iniciar: ' + await res.text()); return; }
    isRunning = true;
    crackBtn.textContent = '⏹ Parar Ataque';
    crackBtn.classList.add('running');
    resetStats();
    clearResult();
    log('info', `Ataque iniciado — modo: ${attackMode}, tipo: ${crackHashType}`);
    connectSSE();
  } catch (e) {
    log('error', 'Falha na conexão: ' + e.message);
  }
}

async function stopAttack() {
  try { await fetch('/api/stop', { method: 'POST' }); } catch {}
  log('warn', 'Ataque interrompido pelo usuário.');
  setIdle();
}

function buildCharset() {
  let cs = '';
  if ($('cs-lower').checked)   cs += 'abcdefghijklmnopqrstuvwxyz';
  if ($('cs-upper').checked)   cs += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  if ($('cs-digits').checked)  cs += '0123456789';
  if ($('cs-symbols').checked) cs += '!@#$%^&*()-_=+[]{}|;:,.<>?';
  return cs;
}

// ── SSE ───────────────────────────────────────────────────────────────────────
function connectSSE() {
  if (evtSource) { evtSource.close(); evtSource = null; }
  evtSource = new EventSource('/api/progress');

  evtSource.onmessage = (e) => {
    try {
      const s = JSON.parse(e.data);
      updateUI(s);
      if (s.status === 'found' || s.status === 'not_found') {
        evtSource.close(); evtSource = null;
        isRunning = false;
        crackBtn.textContent = '⚡ Iniciar Ataque';
        crackBtn.classList.remove('running');
      }
    } catch {}
  };

  evtSource.onerror = () => {
    if (!isRunning) { evtSource.close(); evtSource = null; }
  };
}

// ── UI Update ─────────────────────────────────────────────────────────────────
function updateUI(s) {
  statusBadge.className = `status-badge ${s.status}`;
  const labels = { idle: 'Idle', running: 'Em execução', found: 'Senha Encontrada!', not_found: 'Não Encontrada' };
  statusText.textContent = labels[s.status] || s.status;

  const pct = Math.min(s.percent || 0, 100);
  ringCircle.style.strokeDashoffset = CIRCUMFERENCE * (1 - pct / 100);
  ringPercent.textContent = pct.toFixed(1) + '%';

  const phaseLabels = { dictionary: '📖 Dictionary', rules: '🔀 Rules', bruteforce: '💥 Brute Force', john: '🗡️ John', '': '—' };
  ringPhase.textContent = phaseLabels[s.phase] || s.phase || '—';
  ringCircle.classList.toggle('pulsing', s.status === 'running');

  statHashrate.textContent = s.hashRate > 0 ? formatNum(Math.round(s.hashRate)) : '—';
  statEta.textContent      = s.eta      || '—';
  statElapsed.textContent  = s.elapsed  || '0s';
  statProgress.textContent = formatNum(s.progress || 0);
  statTotalLbl.textContent = s.total > 0 ? `de ${formatNum(s.total)}` : 'de —';

  if (s.status === 'found') {
    resultBanner.className  = 'result-banner found';
    resultLabel.textContent = '🎉 Senha Encontrada';
    resultValue.textContent = s.result;
    resultValue.className   = 'result-value success-text';
    resultIcon.textContent  = '🔓';
    lastAttackMeta.result   = s.result;
    lastAttackMeta.elapsed  = s.elapsed;
    log('success', `Senha encontrada: ${s.result}`);
    setIdle();
  } else if (s.status === 'not_found') {
    resultBanner.className  = 'result-banner not-found';
    resultLabel.textContent = '❌ Não Encontrada';
    resultValue.textContent = 'Senha não encontrada no ataque.';
    resultValue.className   = 'result-value';
    resultIcon.textContent  = '🔒';
    lastAttackMeta.result   = 'Não encontrada';
    lastAttackMeta.elapsed  = s.elapsed;
    log('error', 'Senha não encontrada.');
    setIdle();
  }
}

function setIdle() {
  isRunning = false;
  crackBtn.textContent = '⚡ Iniciar Ataque';
  crackBtn.classList.remove('running');
}

function resetStats() {
  ringCircle.style.strokeDashoffset = CIRCUMFERENCE;
  ringPercent.textContent = '0%';
  ringPhase.textContent = '—';
  statHashrate.textContent = '—';
  statEta.textContent  = '—';
  statElapsed.textContent = '0s';
  statProgress.textContent = '0';
  statTotalLbl.textContent = 'de —';
}

function clearResult() {
  resultBanner.className  = 'result-banner';
  resultLabel.textContent = 'Resultado';
  resultValue.textContent = 'Ataque em andamento...';
  resultValue.className   = 'result-value';
  resultIcon.textContent  = '⚙️';
}

// ── Info ──────────────────────────────────────────────────────────────────────
async function fetchInfo() {
  try {
    const res  = await fetch('/api/info');
    const info = await res.json();
    maxWorkers = info.workers || navigator.hardwareConcurrency || 4;
    crackWorkersEl.max   = maxWorkers;
    crackWorkersEl.value = maxWorkers;
    workersDisplay.textContent = maxWorkers;
    renderWorkerDots(maxWorkers, maxWorkers);

    const badge = $('john-status-badge');
    if (info.john && info.john.available) {
      badge.style.borderColor = 'var(--success)';
      badge.style.color       = 'var(--success)';
      badge.style.background  = 'var(--success-dim)';
      badge.textContent       = '✅ John disponível — ' + (info.john.version || '');
      log('info', 'John the Ripper detectado: ' + info.john.version);
    } else {
      badge.style.borderColor = 'var(--error)';
      badge.style.color       = 'var(--error)';
      badge.style.background  = 'var(--error-dim)';
      badge.textContent       = '❌ John não instalado';
    }
  } catch { /* silencioso */ }
}

// ── Log ───────────────────────────────────────────────────────────────────────
function log(level, msg) {
  const entry = document.createElement('div');
  entry.className = 'log-entry';
  const now = new Date().toLocaleTimeString('pt-BR', { hour12: false });
  entry.innerHTML = `
    <span class="log-time">${now}</span>
    <span class="log-msg ${level}">${escapeHtml(msg)}</span>`;
  logBody.appendChild(entry);
  logBody.scrollTop = logBody.scrollHeight;
}

function initLogClear() {
  $('log-clear-btn').addEventListener('click', () => { logBody.innerHTML = ''; });
}

function initLogExport() {
  $('log-export-btn').addEventListener('click', () => {
    const lines = [];
    lines.push('=== Cipher Scope — Relatório de Ataque ===');
    lines.push(`Data: ${new Date().toLocaleString('pt-BR')}`);
    if (lastAttackMeta.hash) {
      lines.push(`Hash alvo: ${lastAttackMeta.hash}`);
      lines.push(`Tipo: ${lastAttackMeta.hashType}`);
      lines.push(`Modo: ${lastAttackMeta.mode}`);
      lines.push(`Resultado: ${lastAttackMeta.result || '—'}`);
      lines.push(`Tempo: ${lastAttackMeta.elapsed || '—'}`);
    }
    lines.push('');
    lines.push('--- Log ---');
    logBody.querySelectorAll('.log-entry').forEach(e => {
      const time = e.querySelector('.log-time')?.textContent || '';
      const msg  = e.querySelector('.log-msg')?.textContent || '';
      lines.push(`[${time}] ${msg}`);
    });
    const blob = new Blob([lines.join('\n')], { type: 'text/plain' });
    const url  = URL.createObjectURL(blob);
    const a    = document.createElement('a');
    a.href = url; a.download = `cipher-scope-report-${Date.now()}.txt`;
    a.click(); URL.revokeObjectURL(url);
    log('info', 'Relatório exportado.');
  });
}

// ── Encode/Decode Tool ────────────────────────────────────────────────────────
function initEncodeTool() {
  document.querySelectorAll('.format-tab').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.format-tab').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      encodeFormat = btn.dataset.fmt;
    });
  });

  $('encode-btn').addEventListener('click', () => runEncode('encode'));
  $('decode-btn').addEventListener('click', () => runEncode('decode'));

  $('encode-copy-btn').addEventListener('click', () => {
    const val = $('encode-output').value;
    if (!val) return;
    navigator.clipboard.writeText(val).then(() => {
      $('encode-copy-btn').textContent = '✅ Copiado!';
      setTimeout(() => { $('encode-copy-btn').textContent = '📋 Copiar'; }, 2000);
    });
  });
}

async function runEncode(action) {
  const value = $('encode-input').value;
  if (!value) return;
  try {
    const res  = await fetch('/api/encode', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ value, format: encodeFormat, action }),
    });
    const data = await res.json();
    if (data.error) {
      $('encode-output').value = '❌ Erro: ' + data.error;
    } else {
      $('encode-output').value = data.result;
    }
  } catch (e) {
    $('encode-output').value = 'Erro: ' + e.message;
  }
}

// ── Identify Tool ─────────────────────────────────────────────────────────────
function initIdentifyTool() {
  $('identify-btn').addEventListener('click', async () => {
    const hash = $('identify-input').value.trim();
    if (!hash) return;
    try {
      const res  = await fetch('/api/identify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ hash }),
      });
      const data = await res.json();
      const result = $('identify-result');
      const cands  = $('identify-candidates');
      const meta   = $('identify-meta');
      result.style.display = 'block';
      cands.innerHTML = '';
      if (data.candidates && data.candidates.length > 0) {
        data.candidates.forEach(t => {
          const span = document.createElement('span');
          span.className = 'hash-guess-item';
          span.textContent = t.toUpperCase();
          cands.appendChild(span);
        });
        meta.textContent = `Comprimento: ${data.length} chars | Hex válido: ${data.isHex ? 'Sim' : 'Não'}`;
      } else {
        cands.innerHTML = '<span style="color:var(--text-3);font-size:12px;">Tipo não reconhecido</span>';
        meta.textContent = `Comprimento: ${data.length} | Hex: ${data.isHex ? 'Sim' : 'Não'}`;
      }
    } catch (e) {
      log('error', 'Erro ao identificar hash: ' + e.message);
    }
  });
}

// ── Chrome Tool ───────────────────────────────────────────────────────────────
function initChromeTool() {
  const zone      = $('chrome-upload-zone');
  const fileInput = $('chrome-file-input');

  zone.addEventListener('click', () => fileInput.click());
  fileInput.addEventListener('change', () => {
    if (fileInput.files.length > 0) {
      chromeFile = fileInput.files[0];
      zone.classList.add('has-file');
      zone.querySelector('.upload-zone-text').textContent = `✅ ${chromeFile.name}`;
      zone.querySelector('.upload-zone-sub').textContent  = `${(chromeFile.size / 1024).toFixed(1)} KB`;
    }
  });

  // Drag and drop
  zone.addEventListener('dragover', e => { e.preventDefault(); zone.classList.add('drag-over'); });
  zone.addEventListener('dragleave', () => zone.classList.remove('drag-over'));
  zone.addEventListener('drop', e => {
    e.preventDefault();
    zone.classList.remove('drag-over');
    const f = e.dataTransfer.files[0];
    if (f) {
      chromeFile = f;
      zone.classList.add('has-file');
      zone.querySelector('.upload-zone-text').textContent = `✅ ${f.name}`;
      zone.querySelector('.upload-zone-sub').textContent  = `${(f.size / 1024).toFixed(1)} KB`;
    }
  });

  $('chrome-decrypt-btn').addEventListener('click', runChromeDecrypt);

  $('chrome-export-btn').addEventListener('click', () => {
    if (!chromeEntries.length) return;
    const rows = ['URL,Usuário,Senha'];
    chromeEntries.forEach(e => {
      rows.push([
        `"${e.url}"`,
        `"${e.username}"`,
        `"${e.password || e.error || ''}"`,
      ].join(','));
    });
    const blob = new Blob([rows.join('\n')], { type: 'text/csv' });
    const url  = URL.createObjectURL(blob);
    const a    = document.createElement('a');
    a.href = url; a.download = `chrome-passwords-${Date.now()}.csv`;
    a.click(); URL.revokeObjectURL(url);
    log('info', `CSV exportado com ${chromeEntries.length} senhas.`);
  });
}

async function runChromeDecrypt() {
  if (!chromeFile) { log('error', 'Selecione o arquivo Login Data do Chrome primeiro.'); return; }

  const btn = $('chrome-decrypt-btn');
  btn.textContent = '⏳ Processando...';
  btn.disabled = true;

  const formData = new FormData();
  formData.append('db', chromeFile);
  formData.append('masterPassword', $('chrome-master-pwd').value || 'peanuts');

  try {
    const res  = await fetch('/api/chrome', { method: 'POST', body: formData });
    if (!res.ok) {
      const msg = await res.text();
      log('error', 'Erro: ' + msg);
      return;
    }
    const data = await res.json();
    chromeEntries = data.entries || [];

    // Summary
    $('chrome-count').textContent = data.count;
    $('chrome-summary').style.display = 'flex';

    // Table
    const tbody = $('chrome-table-body');
    tbody.innerHTML = '';
    chromeEntries.forEach(e => {
      const tr = document.createElement('tr');
      const pwdClass = e.error ? 'pwd-err' : 'pwd-ok';
      const pwdText  = e.error ? `erro: ${e.error}` : (e.password || '(vazio)');
      tr.innerHTML = `
        <td title="${escapeHtml(e.url)}">${escapeHtml(e.url.replace(/^https?:\/\//, '').substring(0, 40))}</td>
        <td>${escapeHtml(e.username)}</td>
        <td class="${pwdClass}">${escapeHtml(pwdText)}</td>`;
      tbody.appendChild(tr);
    });

    $('chrome-table-wrap').style.display = 'block';
    log('success', `${data.count} senhas extraídas do Chrome!`);

    // Navigate to chrome page (already there)
  } catch (e) {
    log('error', 'Falha na comunicação: ' + e.message);
  } finally {
    btn.textContent = '🔓 Descriptografar Senhas';
    btn.disabled = false;
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function formatNum(n) {
  return n.toLocaleString('pt-BR');
}

function escapeHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}
