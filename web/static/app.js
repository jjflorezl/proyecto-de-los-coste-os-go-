'use strict';
const $ = (id) => document.getElementById(id);
let state = null;
let busy = false;
let initial = true;
let noticeUntil = 0;
const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (c) => ({'&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;'}[c]));
const initials = (name) => name.split(' ').slice(0, 2).map((part) => part[0]).join('');

function notice(text, kind = '', transient = true) {
  $('notice').textContent = text;
  $('notice').className = `notice ${kind}`;
  noticeUntil = transient ? Date.now() + 5500 : 0;
}

function renderPatients() {
  if (!state) return;
  const query = $('search').value.toLocaleLowerCase('es');
  const patients = state.patients.filter((p) => `${p.name} ${p.id}`.toLocaleLowerCase('es').includes(query));
  $('patient-rows').innerHTML = patients.length ? patients.map((p) => `<tr>
    <td><div class="patient-name"><span class="patient-avatar">${escapeHTML(initials(p.name))}</span><span><strong>${escapeHTML(p.name)}</strong><small>${escapeHTML(p.id)} · ${p.age} años · ${escapeHTML(p.level)}</small></span></div></td>
    <td><span class="badge ${p.state_code === 1 ? 'hallway' : p.state_code === 2 ? 'sleep' : ''}">${escapeHTML(p.state)}</span><span class="cell-detail">${escapeHTML(p.location)}</span></td>
    <td>${escapeHTML(p.doctor)}<span class="cell-detail">${p.room ? `Habitación ${p.room}` : 'Sin cama asignada'}</span></td>
    <td><button class="action-btn" data-action="${p.state_code === 0 ? 'attack' : p.state_code === 1 ? 'assign' : 'wake'}" data-patient="${escapeHTML(p.id)}" ${busy ? 'disabled' : ''}>${p.state_code === 0 ? 'Registrar episodio' : p.state_code === 1 ? 'Asignar cama' : 'Despertar'}</button></td>
  </tr>`).join('') : '<tr><td colspan="4" class="empty">No hay pacientes que coincidan con la búsqueda.</td></tr>';
}

function renderHistory() {
  if (!state) return;
  const filter = $('staff-filter').value;
  const episodes = state.episodes.filter((e) => !filter || e.attender_id === filter).slice(-40).reverse();
  $('history').innerHTML = episodes.length ? episodes.map((e) => `<div class="history-row"><span class="event-dot">${e.room ? '✓' : '!'}</span><div class="event-body"><strong>${escapeHTML(e.patient_name)} <span class="cell-detail">${escapeHTML(e.id)} · ${escapeHTML(e.patient)}</span></strong><p>${escapeHTML(e.location)} → ${e.room ? `Habitación ${e.room}` : 'Sin cama asignada'}<br>Atendió: ${escapeHTML(e.attender)}</p></div><time datetime="${escapeHTML(e.date)}">${new Date(e.date).toLocaleTimeString('es-CO', {hour: '2-digit', minute: '2-digit', second: '2-digit', timeZone: 'America/Bogota'})}</time></div>`).join('') : '<div class="empty"><span class="empty-icon">◷</span>Aún no hay episodios en esta vista.<br>Registra un ataque o inicia la simulación.</div>';
}

function render() {
  const hallway = state.patients.filter((p) => p.state_code === 1).length;
  $('total-patients').textContent = String(state.patients.length).padStart(2, '0');
  $('free-beds').textContent = String(state.free_beds).padStart(2, '0');
  $('bed-detail').textContent = `De ${state.rooms.reduce((sum, room) => sum + room.capacity, 0)} camas registradas`;
  $('hallway-count').textContent = String(hallway).padStart(2, '0');
  $('hallway-detail').textContent = hallway ? 'Requieren asignación de cama' : 'Todos están fuera del pasillo';
  $('episode-count').textContent = String(state.episodes.length).padStart(2, '0');
  $('simulation').innerHTML = state.simulation_running ? '<span>□</span> Detener simulación' : '<span>▷</span> Iniciar simulación';
  $('simulation').className = `button primary ${state.simulation_running ? 'running' : ''}`;
  $('simulation').disabled = busy;
  $('rooms').innerHTML = state.rooms.map((r) => `<article class="room ${r.free ? '' : 'occupied'}"><div class="room-left"><span class="bed-icon">▤</span><div><strong>Habitación ${r.number}</strong><small>${r.free} de ${r.capacity} camas libres</small></div></div><span class="room-label">${r.occupants.length ? r.occupants.map(escapeHTML).join(', ') : 'Lista para recibir'}</span></article>`).join('');
  $('severity').innerHTML = state.patients.filter((p) => Object.hasOwn(state.severe, p.id)).map((p) => `<div class="severity-row"><span>${escapeHTML(p.name)}<small>${escapeHTML(p.id)}</small></span><strong>${state.severe[p.id]}</strong></div>`).join('');
  if (initial) {
    $('staff-filter').innerHTML = '<option value="">Todo el personal</option>' + state.staff.map((a) => `<option value="${escapeHTML(a.id)}">${escapeHTML(a.name)}</option>`).join('');
    initial = false;
  }
  renderPatients(); renderHistory();
  if (Date.now() >= noticeUntil) {
    if (state.simulation_error) notice(state.simulation_error, 'error', false);
    else if (state.simulation_running) notice('Simulación en curso: los pacientes duermen y despiertan a intervalos aleatorios. Se detiene a los 30 segundos.', '', false);
    else if (hallway) notice(`${hallway} paciente(s) esperan una cama. Despierta a un paciente en cama y asigna el cupo liberado.`, 'warn', false);
    else notice('Hospital conectado. Puedes registrar episodios, despertar pacientes o iniciar la simulación.', '', false);
  }
}

async function refresh() {
  try {
    const response = await fetch('/api/state');
    if (!response.ok) throw new Error('No se pudo consultar el hospital');
    state = await response.json();
    $('connection-label').textContent = 'En línea';
    $('connection-dot').className = 'live';
    render();
  } catch (error) {
    $('connection-label').textContent = 'Sin conexión';
    $('connection-dot').className = '';
    notice('No hay conexión con el servidor. Comprueba que go run ./cmd/web sigue en ejecución.', 'error');
  }
}

async function act(path, body, successText) {
  if (busy) return;
  busy = true;
  if (state) render();
  try {
    const response = await fetch(path, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)});
    const result = await response.json();
    if (!response.ok) notice(result.error || 'No se pudo realizar la operación', result.episode_recorded ? 'warn' : 'error');
    else notice(successText || result.message);
  } catch (error) { notice('No se pudo conectar con el hospital.', 'error'); }
  finally { busy = false; await refresh(); }
}

$('patient-rows').addEventListener('click', (event) => {
  const button = event.target.closest('button[data-action]');
  if (!button) return;
  const action = button.dataset.action;
  const body = {patient_id: button.dataset.patient};
  if (action === 'attack') body.location = $('location').value;
  const messages = {attack: 'Episodio registrado. El paciente recibió atención y una cama.', wake: 'El paciente despertó y su cama quedó disponible.', assign: 'Cama asignada al paciente que esperaba en el pasillo.'};
  act(`/api/${action}`, body, messages[action]);
});
$('simulation').addEventListener('click', () => { if (state) act(`/api/simulation/${state.simulation_running ? 'stop' : 'start'}`, {}); });
$('search').addEventListener('input', renderPatients);
$('staff-filter').addEventListener('change', renderHistory);
document.querySelectorAll('.nav-item').forEach((link) => link.addEventListener('click', () => { document.querySelectorAll('.nav-item').forEach((item) => item.classList.remove('active')); link.classList.add('active'); }));
$('today').textContent = new Date().toLocaleDateString('es-CO', {day: 'numeric', month: 'short', year: 'numeric', timeZone: 'America/Bogota'});
async function poll() { await refresh(); setTimeout(poll, 1000); }
poll();
