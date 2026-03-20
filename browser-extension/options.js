const input = document.getElementById('syncUrl');
const status = document.getElementById('status');

chrome.storage.sync.get('syncBaseUrl', ({ syncBaseUrl }) => {
  if (syncBaseUrl) input.value = syncBaseUrl;
});

document.getElementById('save').addEventListener('click', () => {
  const url = input.value.trim().replace(/\/$/, '');
  if (!url) {
    status.style.color = 'red';
    status.textContent = 'Please enter a URL.';
    return;
  }
  chrome.storage.sync.set({ syncBaseUrl: url }, () => {
    status.style.color = 'green';
    status.textContent = 'Saved!';
    setTimeout(() => { status.textContent = ''; }, 2000);
  });
});
