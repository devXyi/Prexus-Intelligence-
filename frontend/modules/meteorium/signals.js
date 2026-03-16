/**
 * modules/meteorium/signals.js
 * Prexus Intelligence — Climate Signal Intelligence
 * THE GREAT FILE · Phase 2
 */

import { store } from '../../js/store.js';
import { fPct } from '../../js/utils.js';
import { updateTopbar } from './meteorium.js';

const SIGNALS = [
  { var:'temp_anomaly_c',       val:1.8,   unit:'°C',   label:'Temperature Anomaly',      src:'ERA5 Reanalysis',  fresh:4,  color:'#F59E0B', norm:0.58 },
  { var:'fire_prob_100km',      val:0.42,  unit:'prob', label:'Fire Probability 100km',   src:'NASA FIRMS VIIRS', fresh:2,  color:'#EF4444', norm:0.42 },
  { var:'drought_index',        val:0.61,  unit:'0–1',  label:'Drought Index',            src:'Open-Meteo ERA5',  fresh:6,  color:'#F97316', norm:0.61 },
  { var:'heat_stress_prob_7d',  val:0.54,  unit:'prob', label:'Heat Stress (7-day)',       src:'ECMWF Forecast',   fresh:1,  color:'#F59E0B', norm:0.54 },
  { var:'flood_susceptibility', val:0.38,  unit:'0–1',  label:'Flood Susceptibility',     src:'ERA5 + Terrain',   fresh:6,  color:'#0EA5E9', norm:0.38 },
  { var:'co2_intensity_norm',   val:0.71,  unit:'norm', label:'CO₂ Intensity',           src:'Carbon Monitor',   fresh:18, color:'#8B5CF6', norm:0.71 },
  { var:'transition_risk',      val:0.48,  unit:'0–1',  label:'Transition Risk Score',    src:'Carbon Monitor',   fresh:18, color:'#8B5CF6', norm:0.48 },
  { var:'ndvi',                 val:0.31,  unit:'idx',  label:'Vegetation Health (NDVI)', src:'Sentinel-2 ESA',   fresh:72, color:'#10B981', norm:0.31 },
  { var:'precip_anomaly_pct',   val:-28.4, unit:'%',    label:'Precipitation Anomaly',    src:'ERA5 Reanalysis',  fresh:4,  color:'#0EA5E9', norm:0.64 },
  { var:'wind_speed_ms',        val:12.1,  unit:'m/s',  label:'Wind Speed',               src:'ECMWF Forecast',   fresh:1,  color:'#7BA4C0', norm:0.24 },
  { var:'fire_hazard_score',    val:0.55,  unit:'0–1',  label:'Fire Hazard Score',        src:'FIRMS + ERA5',     fresh:3,  color:'#EF4444', norm:0.55 },
  { var:'carbon_policy_risk',   val:0.64,  unit:'0–1',  label:'Carbon Policy Risk',       src:'Carbon Monitor',   fresh:18, color:'#8B5CF6', norm:0.64 },
];

const CORRELATIONS = [
  { a:'Drought Index',   b:'Fire Probability', strength:0.82, confirmed:true  },
  { a:'Temp Anomaly',    b:'Heat Stress',       strength:0.76, confirmed:true  },
  { a:'NDVI Decline',    b:'Drought Index',     strength:0.71, confirmed:true  },
  { a:'Precip Anomaly',  b:'Flood Risk',        strength:0.64, confirmed:false },
  { a:'CO₂ Intensity', b:'Policy Risk',         strength:0.88, confirmed:true  },
];

let _refreshTimer = null;

export function init(container) {
  updateTopbar({ signalCount: SIGNALS.length });
  container.innerHTML = `
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
      <div class="panel">
        <div class="panel-head">
          <span class="panel-title">Live Signal Feed · ${SIGNALS.length} variables</span>
          <div style="margin-left:auto"><span style="font-size:8px;color:var(--text-muted)">30s refresh</span></div>
        </div>
        <div style="display:flex;gap:10px;padding:5px 12px;border-bottom:1px solid var(--border);background:rgba(0,0,0,.2)">
          ${['Variable','Value','Level','Source','Age'].map((l,i)=>`<div style="font-size:7px;letter-spacing:.12em;text-transform:uppercase;color:var(--text-muted);${i===0?'flex:2.5':i===1?'min-width:65px;text-align:right':i===2?'flex:1':i===3?'width:110px':'width:32px;text-align:right'}">${l}</div>`).join('')}
        </div>
        <div id="signal-feed">
          ${SIGNALS.map(_row).join('')}
        </div>
      </div>
      <div style="display:flex;flex-direction:column;gap:12px">
        <div class="panel">
          <div class="panel-head"><span class="panel-title">Correlation Matrix</span></div>
          <div>${CORRELATIONS.map(_corr).join('')}</div>
        </div>
        <div class="panel">
          <div class="panel-head"><span class="panel-title">Source Freshness</span></div>
          <div style="padding:6px 12px">${_freshness()}</div>
        </div>
        <div class="panel">
          <div class="panel-head"><span class="panel-title">Signal Categories</span></div>
          <div style="padding:10px 12px">${_categories()}</div>
        </div>
      </div>
    </div>`;
  _refreshTimer = setInterval(_jitter, 30_000);
}

export function destroy() {
  if (_refreshTimer) { clearInterval(_refreshTimer); _refreshTimer = null; }
}

function _row(s) {
  const ageC=s.fresh<6?'var(--green)':s.fresh<24?'var(--amber)':'var(--text-muted)';
  const valStr=s.unit==='%'?`${s.val.toFixed(1)}%`:s.unit==='°C'?`${s.val>0?'+':''}${s.val.toFixed(1)}°C`:s.unit==='m/s'?`${s.val.toFixed(1)}`:fPct(s.norm);
  return `<div class="met-signal-row" style="padding:7px 12px;display:flex;gap:10px;align-items:center" id="sig-${s.var}">
    <div style="flex:2.5;font-size:10px;color:var(--text-secondary)">${s.label}</div>
    <div style="min-width:65px;text-align:right;color:${s.color};font-family:var(--font-display);font-size:14px">${valStr}</div>
    <div style="flex:1"><div class="met-risk-bar"><div class="met-risk-fill" style="width:${s.norm*100}%;background:${s.color};opacity:.8"></div></div></div>
    <div style="width:110px;font-size:8px;color:var(--text-muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${s.src}</div>
    <div style="width:32px;text-align:right;font-size:8px;color:${ageC}">${s.fresh}h</div>
  </div>`;
}

function _corr(c) {
  const col=c.strength>0.75?'var(--red)':c.strength>0.6?'var(--amber)':'var(--cobalt)';
  return `<div style="display:flex;align-items:center;gap:10px;padding:7px 12px;border-bottom:1px solid rgba(14,165,233,.07)">
    <div style="flex:1;font-size:10px"><span style="color:var(--text-secondary)">${c.a}</span><span style="color:var(--text-muted);margin:0 6px;font-size:9px">↔</span><span style="color:var(--text-secondary)">${c.b}</span></div>
    <div style="width:80px"><div class="met-risk-bar"><div class="met-risk-fill" style="width:${c.strength*100}%;background:${col}"></div></div></div>
    <div style="font-family:var(--font-display);font-size:15px;color:${col};min-width:36px;text-align:right">${c.strength.toFixed(2)}</div>
    <span class="tag ${c.confirmed?'tag-red':'tag-cobalt'}">${c.confirmed?'CONFIRMED':'PENDING'}</span>
  </div>`;
}

function _freshness() {
  return [['ECMWF Forecast',1,'live'],['NASA FIRMS VIIRS',2,'live'],['ERA5 Reanalysis',4,'live'],['ERA5 + Terrain',6,'live'],['Carbon Monitor',18,'nominal'],['Sentinel-2 ESA',72,'warn']]
    .map(([n,a,s])=>`<div style="display:flex;align-items:center;gap:8px;padding:5px 0;border-bottom:1px solid rgba(14,165,233,.07)">
      <span class="status-dot ${s==='live'?'live':s==='warn'?'warn':'dim'}"></span>
      <span style="flex:1;font-size:10px;color:var(--text-secondary)">${n}</span>
      <span style="font-family:var(--font-display);font-size:14px;color:${s==='warn'?'var(--amber)':'var(--cobalt)'}">${a<24?`${a}h`:`${Math.round(a/24)}d`}</span>
      <span style="font-size:8px;text-transform:uppercase;color:${s==='live'?'var(--green)':s==='warn'?'var(--amber)':'var(--text-muted)'}">${s==='live'?'LIVE':s==='warn'?'STALE':'OK'}</span>
    </div>`).join('');
}

function _categories() {
  return [['Physical Hazard',6,0.50,'#F59E0B'],['Satellite Intel',2,0.17,'#0EA5E9'],['Transition Risk',3,0.25,'#8B5CF6'],['Meteorological',1,0.08,'#10B981']]
    .map(([l,n,p,c])=>`<div style="margin-bottom:10px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px">
        <span style="font-size:10px;color:var(--text-secondary)">${l}</span>
        <span style="font-size:9px;color:var(--text-muted)">${n} signals · ${Math.round(p*100)}%</span>
      </div>
      <div class="met-risk-bar" style="height:4px"><div class="met-risk-fill" style="width:${p*100}%;background:${c}"></div></div>
    </div>`).join('');
}

function _jitter() {
  SIGNALS.forEach(s => {
    s.norm = Math.max(0, Math.min(1, s.norm + (Math.random()-0.5)*0.04));
    const row = document.getElementById(`sig-${s.var}`);
    if (!row) return;
    const bar = row.querySelector('.met-risk-fill');
    if (bar) bar.style.width = `${s.norm*100}%`;
  });
}
