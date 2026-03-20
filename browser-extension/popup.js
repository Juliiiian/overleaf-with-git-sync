let projectId = null;
let syncBaseUrl = null;

const projectIdDisplay = document.getElementById('projectIdDisplay');
const statusBox = document.getElementById('statusBox');
const toast = document.getElementById('toast');
const configSection = document.getElementById('configSection');

function showToast(msg, type = 'success') {
  toast.textContent = msg;
  toast.className = type;
  setTimeout(() => { toast.className = ''; toast.textContent = ''; }, 4000);
}

function extractProjectId(url) {
  const match = url && url.match(/\/project\/([a-f0-9]+)/i);
  return match ? match[1] : null;
}

async function loadStatus() {
  if (!projectId || !syncBaseUrl) return;
  try {
    const res = await fetch(`${syncBaseUrl}/sync/projects/${projectId}/status`);
    const data = await res.json();
    if (data.configured) {
      statusBox.innerHTML = `
        <strong>Repo:</strong> ${data.repo_url}<br>
        <strong>Branch:</strong> ${data.branch}<br>
        <strong>Last sync:</strong> ${data.last_sync ? data.last_sync + ' (' + data.last_sync_direction + ')' : 'never'}<br>
        <strong>Commit:</strong> ${data.last_commit || '—'}
      `;
    } else {
      statusBox.innerHTML = '<span>Not configured. Use "Configure Project" below.</span>';
    }
  } catch {
    statusBox.innerHTML = '<span style="color:red">Could not reach sync service.</span>';
  }
}

async function push() {
  const message = prompt('Commit message:', 'Update from Overleaf');
  if (message === null) return;
  if (!message.trim()) { showToast('Commit message cannot be empty.', 'error'); return; }

  try {
    const res = await fetch(`${syncBaseUrl}/sync/push`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ project_id: projectId, message: message.trim() }),
    });
    const data = await res.json();
    if (!res.ok) { showToast(data.error || 'Push failed.', 'error'); return; }
    showToast(`Pushed ${data.pushed_files} file(s). Commit: ${data.commit_sha?.slice(0, 7)}`);
    loadStatus();
  } catch (e) {
    showToast('Network error: ' + e.message, 'error');
  }
}

async function pull() {
  if (!confirm('Pull from GitHub? This will overwrite Overleaf files.')) return;
  try {
    const res = await fetch(`${syncBaseUrl}/sync/pull`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ project_id: projectId }),
    });
    const data = await res.json();
    if (!res.ok) { showToast(data.error || 'Pull failed.', 'error'); return; }
    showToast(`Pulled ${data.updated_files} file(s). Commit: ${data.commit_sha?.slice(0, 7)}`);
    loadStatus();
  } catch (e) {
    showToast('Network error: ' + e.message, 'error');
  }
}

async function saveConfig() {
  const repoUrl = document.getElementById('repoUrl').value.trim();
  const branch = document.getElementById('branch').value.trim() || 'main';
  const token = document.getElementById('token').value.trim();

  if (!repoUrl || !token) { showToast('Repo URL and token are required.', 'error'); return; }

  try {
    const res = await fetch(`${syncBaseUrl}/sync/projects/${projectId}/config`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ repo_url: repoUrl, branch, github_token: token }),
    });
    const data = await res.json();
    if (!res.ok) { showToast(data.error || 'Config save failed.', 'error'); return; }
    showToast('Config saved!');
    configSection.classList.remove('visible');
    document.getElementById('token').value = '';
    loadStatus();
  } catch (e) {
    showToast('Network error: ' + e.message, 'error');
  }
}

// Wire up events
document.getElementById('pushBtn').addEventListener('click', push);
document.getElementById('pullBtn').addEventListener('click', pull);
document.getElementById('saveConfigBtn').addEventListener('click', saveConfig);
document.getElementById('toggleConfigBtn').addEventListener('click', () => {
  configSection.classList.toggle('visible');
});
document.getElementById('optionsLink').addEventListener('click', () => {
  chrome.runtime.openOptionsPage();
});

// Init
chrome.storage.sync.get('syncBaseUrl', async ({ syncBaseUrl: url }) => {
  syncBaseUrl = url ? url.replace(/\/$/, '') : null;
  if (!syncBaseUrl) {
    statusBox.innerHTML = '<span style="color:red">Set your sync URL in <a href="#" id="settingsLink">Settings</a>.</span>';
    document.getElementById('settingsLink')?.addEventListener('click', (e) => {
      e.preventDefault();
      chrome.runtime.openOptionsPage();
    });
    return;
  }

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  projectId = extractProjectId(tab?.url);

  if (projectId) {
    projectIdDisplay.textContent = `Project: ${projectId}`;
  } else {
    projectIdDisplay.textContent = 'Not on an Overleaf project page.';
    document.getElementById('pushBtn').disabled = true;
    document.getElementById('pullBtn').disabled = true;
    document.getElementById('toggleConfigBtn').disabled = true;
  }

  await loadStatus();
});
