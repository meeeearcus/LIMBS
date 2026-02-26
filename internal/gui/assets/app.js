const form = document.getElementById('exportForm');
const statusEl = document.getElementById('status');
const errorEl = document.getElementById('error');
const resultEl = document.getElementById('result');
const runButton = document.getElementById('runButton');
const guidanceEl = document.getElementById('guidance');
const previewEl = document.getElementById('preview');
const storageKey = 'limbs_gui_v1';

function normalizePathValue(value) {
  const trimmed = (value || '').toString().trim();
  if (
    (trimmed.startsWith('"') && trimmed.endsWith('"')) ||
    (trimmed.startsWith("'") && trimmed.endsWith("'"))
  ) {
    return trimmed.slice(1, -1).trim();
  }
  return trimmed;
}

function formToPayload() {
  const data = new FormData(form);
  return {
    sourceMount: normalizePathValue(data.get('sourceMount')),
    projectsRoot: normalizePathValue(data.get('projectsRoot')),
    samplesRoot: normalizePathValue(data.get('samplesRoot')),
    usbDrive: normalizePathValue(data.get('usbDrive')),
    projectName: (data.get('projectName') || '').toString().trim(),
    projectFile: normalizePathValue(data.get('projectFile')),
    destRoot: normalizePathValue(data.get('destRoot')),
    limbsRoot: (data.get('limbsRoot') || '').toString().trim(),
    zip: data.get('zip') === 'on',
    allowMissing: data.get('allowMissing') === 'on'
  };
}

function payloadToCommand(payload) {
  const args = ['limbs', 'export'];
  if (payload.projectName) args.push('--project-name', payload.projectName);
  if (payload.projectFile) args.push('--project-file', payload.projectFile);
  if (payload.destRoot) args.push('--dest-root', payload.destRoot);
  if (payload.sourceMount) args.push('--source-mount', payload.sourceMount);
  if (payload.projectsRoot) args.push('--projects-root', payload.projectsRoot);
  if (payload.samplesRoot) args.push('--samples-root', payload.samplesRoot);
  if (payload.usbDrive) args.push('--usb-drive', payload.usbDrive);
  if (payload.limbsRoot) args.push('--limbs-root', payload.limbsRoot);
  if (payload.zip) args.push('--zip');
  args.push(`--allow-missing=${payload.allowMissing}`);
  return args.join(' ');
}

function renderGuidance(payload) {
  if (
    payload.sourceMount.includes('"') || payload.projectsRoot.includes('"') ||
    payload.samplesRoot.includes('"') || payload.usbDrive.includes('"') ||
    payload.projectFile.includes('"') || payload.destRoot.includes('"')
  ) {
    guidanceEl.textContent = 'Do not use shell quoting in paths. Enter raw paths only.';
    return;
  }

  if (payload.projectName && payload.projectFile) {
    guidanceEl.textContent = 'Use either Project Name or Project File, not both.';
  } else if (!payload.projectName && !payload.projectFile) {
    guidanceEl.textContent = 'Set Project Name or Project File.';
  } else {
    guidanceEl.textContent = '';
  }
}

function updatePreview() {
  const payload = formToPayload();
  previewEl.textContent = payloadToCommand(payload);
  renderGuidance(payload);
  localStorage.setItem(storageKey, JSON.stringify(payload));
}

function loadSaved() {
  const raw = localStorage.getItem(storageKey);
  if (!raw) return;
  try {
    const payload = JSON.parse(raw);
    Object.entries(payload).forEach(([k, v]) => {
      const field = form.elements.namedItem(k);
      if (!field) return;
      if (field.type === 'checkbox') {
        field.checked = Boolean(v);
      } else {
        field.value = v || '';
      }
    });
  } catch {
    // ignore invalid stored values
  }
}

form.addEventListener('input', updatePreview);
form.addEventListener('submit', async (e) => {
  e.preventDefault();
  const payload = formToPayload();

  errorEl.textContent = '';
  resultEl.textContent = '';
  statusEl.textContent = 'Running export...';
  runButton.disabled = true;

  try {
    const res = await fetch('/api/export', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const body = await res.json();
    if (!res.ok) {
      throw new Error(body.error || 'Export failed');
    }

    statusEl.textContent = 'Export complete.';
    resultEl.innerHTML = `<pre>${JSON.stringify(body, null, 2)}</pre>`;
  } catch (err) {
    statusEl.textContent = '';
    errorEl.textContent = err.message;
  } finally {
    runButton.disabled = false;
  }
});

loadSaved();
updatePreview();
