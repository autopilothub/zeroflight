const state = {
  lastStatus: null,
  pollMs: 1000,
};

const els = {
  pillConnected: document.getElementById('pill-connected'),
  pillArmed: document.getElementById('pill-armed'),
  pillMode: document.getElementById('pill-mode'),
  pillGcsNav: document.getElementById('pill-gcsnav'),
  battery: document.getElementById('battery'),
  gpsFix: document.getElementById('gps-fix'),
  altitude: document.getElementById('altitude'),
  speed: document.getElementById('speed'),
  hdop: document.getElementById('hdop'),
  updated: document.getElementById('updated'),
  horizon: document.getElementById('horizon'),
  roll: document.getElementById('roll'),
  pitch: document.getElementById('pitch'),
  yaw: document.getElementById('yaw'),
  lat: document.getElementById('lat'),
  lon: document.getElementById('lon'),
  home: document.getElementById('home'),
  imuSection: document.getElementById('imu-section'),
  imuData: document.getElementById('imu-data'),
  droneDot: document.getElementById('drone-dot'),
  preflightList: document.getElementById('preflight-list'),
  gotoForm: document.getElementById('goto-form'),
  gotoLat: document.getElementById('goto-lat'),
  gotoLon: document.getElementById('goto-lon'),
  gotoAlt: document.getElementById('goto-alt'),
  gotoForce: document.getElementById('goto-force'),
  hoverAlt: document.getElementById('hover-alt'),
  message: document.getElementById('message'),
};

function radToDeg(rad) {
  return (rad * 180) / Math.PI;
}

function setPill(el, text, level) {
  el.textContent = text;
  el.className = 'pill' + (level ? ` ${level}` : '');
}

function showMessage(text, level) {
  els.message.textContent = text;
  els.message.className = 'message' + (level ? ` ${level}` : '');
}

async function api(path, options) {
  const res = await fetch(path, options);
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return body;
}

function updateStatus(data) {
  state.lastStatus = data;

  setPill(els.pillConnected, data.connected ? 'MAVLink active' : (data.link_open ? 'Serial open' : 'Disconnected'),
    data.connected ? 'ok' : (data.parse_errors > 0 ? 'bad' : 'warn'));
  setPill(els.pillArmed, data.armed ? 'Armed' : 'Disarmed', data.armed ? 'warn' : 'ok');
  setPill(els.pillMode, `Mode ${data.mode || '—'}`, '');
  setPill(
    els.pillGcsNav,
    data.gcs_nav_active ? 'GCS NAV on' : 'GCS NAV off',
    data.gcs_nav_active ? 'ok' : 'warn',
  );

  const bat = data.battery || {};
  els.battery.textContent = `${bat.voltage_v?.toFixed?.(1) ?? '—'} V / ${bat.remaining_pct ?? '—'}%`;

  const gps = data.gps || {};
  els.gpsFix.textContent = `fix ${gps.fix_type ?? 0}, ${gps.satellites ?? 0} sats`;
  els.altitude.textContent = `${gps.rel_alt_m?.toFixed?.(1) ?? '—'} m`;
  els.speed.textContent = `${gps.ground_speed?.toFixed?.(1) ?? '—'} m/s`;
  els.hdop.textContent = gps.hdop?.toFixed?.(1) ?? '—';
  els.updated.textContent = data.time ? new Date(data.time).toLocaleTimeString() : '—';

  const att = data.attitude || {};
  const rollDeg = radToDeg(att.roll || 0);
  const pitchDeg = radToDeg(att.pitch || 0);
  const yawDeg = radToDeg(att.yaw || 0);
  els.roll.textContent = rollDeg.toFixed(1);
  els.pitch.textContent = pitchDeg.toFixed(1);
  els.yaw.textContent = yawDeg.toFixed(1);
  els.horizon.style.transform = `rotate(${rollDeg}deg)`;
  els.horizon.querySelector('.horizon-sky').style.transform = `translateY(${pitchDeg * 1.2}px)`;
  els.horizon.querySelector('.horizon-ground').style.transform = `translateY(${pitchDeg * 1.2}px)`;

  els.lat.textContent = gps.lat ? gps.lat.toFixed(7) : '—';
  els.lon.textContent = gps.lon ? gps.lon.toFixed(7) : '—';

  const home = data.home || {};
  if (home.valid) {
    els.home.textContent = `${home.lat.toFixed(7)}, ${home.lon.toFixed(7)}`;
  } else {
    els.home.textContent = 'waiting…';
  }

  updateMap(gps, home);

  const imu = data.raw_imu || {};
  if (imu.available) {
    els.imuSection.hidden = false;
    els.imuData.textContent = `accel ${JSON.stringify(imu.accel)}\ngyro  ${JSON.stringify(imu.gyro)}\nmag   ${JSON.stringify(imu.mag)}`;
  } else {
    els.imuSection.hidden = true;
  }
}

function updateMap(gps, home) {
  const cx = 200;
  const cy = 200;
  const scale = 2; // meters per pixel at ring edge (~120px ~ 60m)

  if (!home.valid || !gps.lat || !gps.lon) {
    els.droneDot.setAttribute('cx', String(cx));
    els.droneDot.setAttribute('cy', String(cy));
    return;
  }

  const dLat = (gps.lat - home.lat) * 111320;
  const dLon = (gps.lon - home.lon) * 111320 * Math.cos((home.lat * Math.PI) / 180);
  const x = cx + dLon / scale;
  const y = cy - dLat / scale;
  els.droneDot.setAttribute('cx', String(Math.max(20, Math.min(380, x))));
  els.droneDot.setAttribute('cy', String(Math.max(20, Math.min(380, y))));
}

async function pollStatus() {
  try {
    const data = await api('/api/v1/status');
    updateStatus(data);
  } catch (err) {
    setPill(els.pillConnected, 'API error', 'bad');
    showMessage(err.message, 'error');
  }
}

async function runPreflight() {
  try {
    const data = await api('/api/v1/preflight');
    els.preflightList.innerHTML = '';
    (data.checks || []).forEach((check) => {
      const li = document.createElement('li');
      li.textContent = `${check.passed ? 'PASS' : 'FAIL'} — ${check.name}: ${check.message}`;
      li.className = check.passed ? 'pass' : 'fail';
      els.preflightList.appendChild(li);
    });
    showMessage(data.passed ? 'Preflight passed' : 'Preflight failed', data.passed ? 'ok' : 'error');
  } catch (err) {
    showMessage(err.message, 'error');
  }
}

document.getElementById('btn-preflight').addEventListener('click', runPreflight);

document.getElementById('btn-use-position').addEventListener('click', () => {
  const gps = state.lastStatus?.gps;
  if (!gps?.lat || !gps?.lon) {
    showMessage('No GPS position yet', 'error');
    return;
  }
  els.gotoLat.value = gps.lat;
  els.gotoLon.value = gps.lon;
  if (!els.gotoAlt.value) {
    els.gotoAlt.value = gps.rel_alt_m || 10;
  }
  showMessage('Filled goto form from current GPS', 'ok');
});

els.gotoForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  try {
    await api('/api/v1/goto', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        lat: Number(els.gotoLat.value),
        lon: Number(els.gotoLon.value),
        alt: Number(els.gotoAlt.value),
        force: els.gotoForce.checked,
      }),
    });
    showMessage('Goto command sent', 'ok');
  } catch (err) {
    showMessage(err.message, 'error');
  }
});

document.getElementById('btn-hover').addEventListener('click', async () => {
  try {
    await api('/api/v1/hover', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        alt: Number(els.hoverAlt.value),
        force: false,
      }),
    });
    showMessage('Hover command sent', 'ok');
  } catch (err) {
    showMessage(err.message, 'error');
  }
});

pollStatus();
setInterval(pollStatus, state.pollMs);
