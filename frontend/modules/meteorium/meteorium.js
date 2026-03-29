/**
 * modules/meteorium/meteorium.js
 * Prexus Intelligence — Meteorium Module Orchestrator
 *
 * Controls: sidebar nav, topbar, module mounting/unmounting.
 * Every sub-module calls updateTopbar() from here.
 */

import { store, subscribe }             from '../../js/store.js';
import { initPrefetch, destroyPrefetch } from '../../js/prefetch.js';
import { startPredictionScheduler, stopPredictionScheduler } from '../../js/predict.js';
import { logout }   from '../../js/auth.js';
import { navigate } from '../../js/router.js';
import { startClock } from '../../js/utils.js';

/* ── Lazy module imports ─────────────────────────────────── */
const MODULES = {
  dashboard: () => import('./dashboard.js'),
  portfolio: () => import('./portfolio.js'),
  analysis:  () => import('./analysis.js'),
  signals:   () => import('./signals.js'),
  sources:   () => import('./sources.js'),
  pipeline:  () => import('./pipeline.js'),
  ai:        () => import('./ai.js'),
};

const NAV = [
  { id:'dashboard', ico:'fa-chart-line',     label:'Overview'       },
  { id:'portfolio', ico:'fa-layer-group',    label:'Asset Portfolio' },
  { id:'analysis',  ico:'fa-bolt',           label:'Risk Analysis'   },
  { id:'signals',   ico:'fa-wave-square',    label:'Signals'        },
  { id:'sources',   ico:'fa-satellite-dish', label:'Data Sources'   },
  { id:'pipeline',  ico:'fa-sitemap',        label:'Pipeline'       },
  { id:'ai',        ico:'fa-microchip',      label:'AI Intelligence', badge:'AI' },
];

const VIEW_TITLES = {
  dashboard: 'Overview',
  portfolio: 'Asset Portfolio',
  analysis:  'Risk Analysis',
  signals:   'Signals',
  sources:   'Data Sources',
  pipeline:  'Pipeline',
  ai:        'AI Intelligence',
};

let _activeModule  = null;
let _activeId      = null;
let _clockStop     = null;

/* ══════════════════════════════════════════════════════════
   PUBLIC API
══════════════════════════════════════════════════════════ */

/** Called by app.js when navigating to Meteorium */
export async function initMeteorium() {
  _renderShell();
  startPredictionScheduler();
  initPrefetch();
  _startClock();
  await _mount('dashboard');
}

/** Destroy Meteorium — called on navigate away */
export function destroyMeteorium() {
  _unmountCurrent();
  stopPredictionScheduler();
  destroyPrefetch();
  if (_clockStop) { _clockStop(); _clockStop = null; }
}

/**
 * Sub-modules call this to push data into the topbar.
 * @param {object} opts - { signalCount, title }
 */
export function updateTopbar(opts = {}) {
  if (opts.title) {
    const el = document.getElementById('met-topbar-title');
    if (el) el.textContent = opts.title;
  }
  if (opts.signalCount !== undefined) {
    const el = document.getElementById('met-topbar-alerts');
    if (el) {
      el.textContent = opts.signalCount;
      el.style.display = opts.signalCount > 0 ? 'flex' : 'none';
    }
  }
}

/* ══════════════════════════════════════════════════════════
   SHELL RENDER
══════════════════════════════════════════════════════════ */

function _renderShell() {
  const page = document.getElementById('page-meteorium');
  if (!page) return;

  const user = store.get('user');
  const org  = store.get('org');

  page.innerHTML = `
    <!-- Classification banner -->
    <div class="met-cls">UNCLASSIFIED // FOR OFFICIAL USE ONLY (FOUO) // DISTRIBUTION C // METEORIUM PLATFORM v3.1.0</div>

    <!-- App shell -->
    <div class="met-shell">

      <!-- Sidebar -->
      <div class="met-sidebar">
        <div class="met-sidebar-logo">
          <div class="met-sidebar-icon">
            <i class="fa-solid fa-cloud-bolt" style="font-size:13px;color:var(--cobalt)"></i>
          </div>
          <div>
            <div class="met-sidebar-name">METEORIUM</div>
            <div class="met-sidebar-sub">by PREXUS</div>
          </div>
        </div>

        <div style="padding:8px 0;flex:1;overflow:auto">
          <div class="met-nav-section">Navigation</div>
          ${NAV.map(n => `
            <div class="met-nav-item ${n.id === 'dashboard' ? 'active' : ''}" id="nav-${n.id}" onclick="window._met_nav('${n.id}')">
              <span class="ico"><i class="fa-solid ${n.ico}"></i></span>
              ${n.label}
              ${n.badge ? `<span class="met-nav-badge">${n.badge}</span>` : ''}
            </div>`).join('')}

          <div class="met-nav-section" style="margin-top:14px">System</div>
          <div class="met-nav-item" onclick="window._met_back()">
            <span class="ico"><i class="fa-solid fa-grid-2"></i></span>
            Back to Hub
          </div>
        </div>

        <div class="met-sidebar-footer">
          <div style="display:flex;align-items:center;gap:5px;margin-bottom:6px">
            <span class="status-dot live pulse"></span>
            <span style="font-family:var(--font-data);font-size:8px;color:var(--text-muted)">Kernel ONLINE</span>
          </div>
          ${[['ORG', org?.orgName || user?.org_name || '—'],['ROLE', user?.role || 'ORG_ADMIN'],['USER', user?.email || '—']].map(([k,v]) =>
            `<div class="met-user-line">${k}: <span style="color:${k==='ROLE'?'var(--green)':'var(--text-secondary)'}">${v}</span></div>`).join('')}
          <button onclick="window._met_logout()"
            style="width:100%;margin-top:8px;padding:5px;background:transparent;border:1px solid var(--border);color:var(--text-secondary);border-radius:3px;cursor:pointer;font-family:var(--font-data);font-size:9px;letter-spacing:.1em;text-transform:uppercase;display:flex;align-items:center;justify-content:center;gap:6px;transition:all .2s"
            onmouseover="this.style.borderColor='var(--border-hi)';this.style.color='var(--text-primary)'"
            onmouseout="this.style.borderColor='var(--border)';this.style.color='var(--text-secondary)'">
            <i class="fa-solid fa-right-from-bracket"></i>Sign Out
          </button>
        </div>
      </div>

      <!-- Main -->
      <div class="met-main">
        <!-- Topbar -->
        <div class="met-topbar">
          <span class="met-topbar-title" id="met-topbar-title">Overview</span>
          <div class="met-topbar-right">
            <span id="met-loading-indicator" style="font-family:var(--font-data);font-size:8px;color:var(--text-muted);display:none">Loading…</span>
            <div style="display:flex;align-items:center;gap:5px">
              <span class="status-dot live pulse"></span>
              <span style="font-family:var(--font-data);font-size:8.5px;color:var(--text-muted)">LIVE</span>
            </div>
            <span id="met-utc-clock" style="font-family:var(--font-data);font-size:9px;color:var(--text-muted)"></span>
            <div id="met-topbar-alerts" style="display:none;align-items:center;gap:4px">
              <i class="fa-solid fa-bell" style="font-size:10px;color:var(--amber)"></i>
              <span style="font-family:var(--font-data);font-size:9.5px;color:var(--amber);font-weight:700">0</span>
            </div>
          </div>
        </div>

        <!-- Workspace -->
        <div class="met-workspace" id="met-workspace">
          <div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-muted);font-family:var(--font-data);font-size:10px">
            <span class="spinner"></span>&nbsp;Loading…
          </div>
        </div>
      </div>
    </div>

    <!-- Classification footer -->
    <div class="met-cls">UNCLASSIFIED // FOR OFFICIAL USE ONLY (FOUO) // DISTRIBUTION C // PROPRIETARY — PREXUS INC.</div>`;

  /* Wire global handlers */
  window._met_nav    = _navTo;
  window._met_back   = () => { if (window._met_onBack) window._met_onBack(); else navigate('hub'); };
  window._met_logout = () => logout();
}

/* ══════════════════════════════════════════════════════════
   MODULE MOUNTING
══════════════════════════════════════════════════════════ */

async function _mount(id) {
  if (_activeId === id) return;

  _unmountCurrent();
  stopPredictionScheduler();
  destroyPrefetch();
  _setNavActive(id);

  const workspace = document.getElementById('met-workspace');
  if (!workspace) return;

  // Update topbar title immediately
  const titleEl = document.getElementById('met-topbar-title');
  if (titleEl) titleEl.textContent = VIEW_TITLES[id] || id;

  // Show loader
  workspace.innerHTML = `<div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-muted);font-family:var(--font-data);font-size:10px">
    <span class="spinner"></span>&nbsp;Loading ${VIEW_TITLES[id]}…
  </div>`;

  const loader = MODULES[id];
  if (!loader) return;

  try {
    const mod = await loader();
    _activeModule = mod;
    _activeId = id;
    store.set('module', id);

    // Clear workspace and mount
    workspace.innerHTML = '';
    await mod.init(workspace);
  } catch (e) {
    console.error(`[meteorium] Failed to load module "${id}":`, e);
    workspace.innerHTML = `<div style="padding:20px;color:var(--red);font-family:var(--font-data);font-size:10px;background:var(--red-lo);border:1px solid rgba(239,68,68,.3);border-radius:3px;margin:14px">
      ⚠ Failed to load ${id}: ${e.message}
    </div>`;
  }
}

function _unmountCurrent() {
  if (_activeModule?.destroy) {
    try { _activeModule.destroy(); } catch {}
  }
  _activeModule = null;
  _activeId = null;
}

async function _navTo(id) {
  if (_activeId === id) return;
  await _mount(id);
}

function _setNavActive(id) {
  NAV.forEach(n => {
    const el = document.getElementById(`nav-${n.id}`);
    if (el) el.classList.toggle('active', n.id === id);
  });
}

/* ══════════════════════════════════════════════════════════
   CLOCK
══════════════════════════════════════════════════════════ */

function _startClock() {
  if (_clockStop) _clockStop();
  _clockStop = startClock(({ time }) => {
    const el = document.getElementById('met-utc-clock');
    if (el) el.textContent = time;
  });
}
