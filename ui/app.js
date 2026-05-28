/* ── Cipher Scope Dashboard — app.js ── */

const CIRCUMFERENCE = 471.24; // 2 * π * 75

// ── State ─────────────────────────────────────────────────────────────────────
let genHashType   = 'md5';
let crackHashType = 'md5';
let attackMode    = 'auto';
let isRunning     = false;
let evtSource     = null;
let genDebounce   = null;
let maxWorkers    = navigator.hardwareConcurrency || 4;

// ── Element refs ──────────────────────────────────────────────────────────────
const $ = id => document.getElementById(id);

const genPassword   = $('gen-password');
const genHashValue  = $('gen-hash-value');
const genCopyBtn    = $('gen-copy-btn');
const crackHashEl   = $('crack-hash');
const crackLenEl    = $('crack-length');
const lengthDisplay = $('length-display');
const crackWorkersEl= $('crack-workers');
const workersDisplay= $('workers-display');
const bfOptions     = $('bf-options');
const crackBtn      = $('crack-btn');
const ringCircle    = $('ring-circle');
const ringPercent   = $('ring-percent');
const ringPhase     = $('ring-phase');
const statusBadge   = $('status-badge');
const statusText    = $('status-text');
const workersDots   = $('workers-dots');
const workersLabel  = $('workers-label');
const statHashrate  = $('stat-hashrate');
const statEta       = $('stat-eta');
const statElapsed   = $('stat-elapsed');
const statProgress  = $('stat-progress');
const statTotalLbl  = $('stat-total-label');
const resultBanner  = $('result-banner');
const resultLabel   = $('result-label');
const resultValue   = $('result-value');
const resultIcon    = $('result-icon');
const logBody       = $('log-body');

// ── Init ──────────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  initWorkers();
  initTabs();
  initSliders();
  initGenerator();
  initCopyBtn();
  initCrackBtn();
  initLogClear();
  connectSSE();
  fetchInfo();
  log('info', 'Cipher Scope iniciado.');
});

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
      genCopyBtn.classList.add('copied');
      genCopyBtn.querySelector('span') && (genCopyBtn.innerHTML = `
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none">
          <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg> Copiado!`);
      setTimeout(() => {
        genCopyBtn.classList.remove('copied');
        genCopyBtn.innerHTML = `
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none">
            <rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" stroke-width="2"/>
            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="currentColor" stroke-width="2"/>
          </svg> Copiar`;
      }, 2000);
    });
  });
}

// ── Crack Button ──────────────────────────────────────────────────────────────
function initCrackBtn() {
  crackBtn.addEventListener('click', async () => {
    if (isRunning) {
      await stopAttack();
      return;
    }

    const hash = crackHashEl.value.trim();
    if (!hash) {
      log('error', 'Cole um hash antes de iniciar.');
      return;
    }

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
    hashType:   crackHashType,
    mode:       attackMode,
    workers:    parseInt(crackWorkersEl.value),
    maxLength:  parseInt(crackLenEl.value),
    charset,
    wordlist:   'wordlist.txt',
    johnRules:  $('john-rules-check') ? $('john-rules-check').checked : false,
  };

  try {
    const res = await fetch('/api/crack', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      const msg = await res.text();
      log('error', 'Erro ao iniciar: ' + msg);
      return;
    }

    isRunning = true;
    crackBtn.textContent = '⏹ Parar Ataque';
    crackBtn.classList.add('running');
    resetStats();
    clearResult();
    log('info', `Ataque iniciado — modo: ${attackMode}, hash: ${crackHashType}`);
    connectSSE();
  } catch (e) {
    log('error', 'Falha na conexão: ' + e.message);
  }
}

async function stopAttack() {
  try {
    await fetch('/api/stop', { method: 'POST' });
    log('warn', 'Ataque interrompido pelo usuário.');
  } catch {}
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
        evtSource.close();
        evtSource = null;
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
  // Status badge
  statusBadge.className = `status-badge ${s.status}`;
  const labels = { idle: 'Idle', running: 'Em execução', found: 'Senha Encontrada!', not_found: 'Não Encontrada' };
  statusText.textContent = labels[s.status] || s.status;

  // Progress ring
  const pct = Math.min(s.percent || 0, 100);
  const offset = CIRCUMFERENCE * (1 - pct / 100);
  ringCircle.style.strokeDashoffset = offset;
  ringPercent.textContent = pct.toFixed(1) + '%';

  const phaseLabels = {
    dictionary: '📖 Dictionary',
    rules:      '🔀 Rules',
    bruteforce: '💥 Brute Force',
    john:       '🗡️ John',
    '':         '—'
  };
  ringPhase.textContent = phaseLabels[s.phase] || s.phase || '—';

  const isPulsing = s.status === 'running';
  ringCircle.classList.toggle('pulsing', isPulsing);

  // Stats
  statHashrate.textContent = s.hashRate > 0 ? formatNum(Math.round(s.hashRate)) : '—';
  statEta.textContent      = s.eta      || '—';
  statElapsed.textContent  = s.elapsed  || '0s';
  statProgress.textContent = formatNum(s.progress || 0);
  statTotalLbl.textContent = s.total > 0 ? `de ${formatNum(s.total)}` : 'de —';

  // Result banner
  if (s.status === 'found') {
    resultBanner.className    = 'result-banner found';
    resultLabel.textContent   = '🎉 Senha Encontrada';
    resultValue.textContent   = s.result;
    resultValue.className     = 'result-value success-text';
    resultIcon.textContent    = '🔓';
    log('success', `Senha encontrada: ${s.result}`);
    setIdle();
  } else if (s.status === 'not_found') {
    resultBanner.className    = 'result-banner not-found';
    resultLabel.textContent   = '❌ Não Encontrada';
    resultValue.textContent   = 'Senha não encontrada no ataque.';
    resultValue.className     = 'result-value';
    resultIcon.textContent    = '🔒';
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

// ── Info (john + workers) ─────────────────────────────────────────────────────
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
      badge.textContent       = '❌ John não instalado — use o script scripts/setup-john-arm64.sh';
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

// ── Helpers ───────────────────────────────────────────────────────────────────
function formatNum(n) {
  return n.toLocaleString('pt-BR');
}

function escapeHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}
