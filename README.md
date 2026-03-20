# Overleaf + GitHub Sync

Adds manual GitHub push/pull to self-hosted Overleaf CE via a browser extension and a small sync service running alongside Overleaf in Docker.

## Setup

```bash
cp .env.example .env   # fill in your domain and admin credentials
docker compose up -d
```

Configure path routing in Coolify so `/sync/*` proxies to `sync-service:8080`.

## Browser Extension

### Firefox

**[⬇ Click to install](https://github.com/Juliiiian/overleaf--with-git-sync/releases/latest/download/overleaf-github-sync-firefox.xpi)**

### Chrome

1. Download [`overleaf-github-sync-chrome.zip`](https://github.com/Juliiiian/overleaf--with-git-sync/releases/latest/download/overleaf-github-sync-chrome.zip) and unzip it
2. Open `chrome://extensions` → enable **Developer mode** → **Load unpacked** → select the folder

### First use

1. Click the extension icon → **Settings** → enter your Overleaf URL (e.g. `https://latex.example.com`)
2. Open a project → **Configure Project** → add your GitHub repo URL, branch, and a PAT token
3. Use **Push** / **Pull** to sync

## License

MIT
