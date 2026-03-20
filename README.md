# Overleaf + GitHub Sync

Self-hosted Overleaf Community Edition with a custom sync service for manually pushing and pulling projects to/from GitHub.

## Overview

Overleaf CE has no built-in Git integration. This project adds it via:

- **sync-service** — a Go HTTP service with direct access to the Overleaf data volume
- **browser extension** — a Chrome/Firefox popup to trigger push/pull and configure repos

Sync is always **manual and explicit** — you decide when to push or pull. No real-time sync, no merge conflicts.

```
[Browser Extension Popup]
        │
        │  POST /sync/push  or  /sync/pull
        ▼
[Your Overleaf domain]   (Coolify routes /sync/* → sync-service)
        │
   ┌────┴─────────────────────┐
   │                          │
[sharelatex]          [sync-service]
[mongo][redis]     reads/writes Overleaf
                   data volume + git repos
```

---

## Requirements

- Docker + Docker Compose
- Coolify (or any reverse proxy that can route by path prefix)
- A GitHub personal access token (PAT) with `repo` scope
- Chrome or a Chromium-based browser for the extension

---

## Setup

### 1. Clone and configure

```bash
git clone https://github.com/Juliiiian/overleaf--with-git-sync.git
cd overleaf--with-git-sync
cp .env.example .env
```

Edit `.env`:

```bash
OVERLEAF_SITE_URL=https://latex.yourdomain.com
OVERLEAF_APP_NAME=My Overleaf
OVERLEAF_ADMIN_EMAIL=admin@example.com
OVERLEAF_ADMIN_PASSWORD=a-strong-password
```

### 2. Configure Coolify routing

In Coolify, add a path-prefix rule so the sync service is reachable under the same domain as Overleaf:

```
https://latex.yourdomain.com/sync/*  →  sync-service:8080
```

This avoids CORS entirely — the extension communicates with the same origin as Overleaf.

> If you use a different reverse proxy (nginx, Caddy, Traefik), add an equivalent location block:
> ```nginx
> location /sync/ {
>     proxy_pass http://sync-service:8080;
> }
> ```

### 3. Start the stack

```bash
docker compose up -d
```

On first boot Overleaf will take a minute to initialize. Then open your domain and register an admin account via the `/launchpad` page.

### 4. Verify the data path

After creating a test project and uploading a file, confirm where Overleaf stores files:

```bash
docker exec <sharelatex-container-name> find /var/lib/overleaf -maxdepth 4 -type d
```

The default expected path is `/var/lib/overleaf/data/user_files/<project_id>/`. If your version uses a different path, update `OVERLEAF_DATA_PATH` in `docker-compose.yml`.

---

## Browser Extension

### Install in Chrome

1. Open `chrome://extensions`
2. Enable **Developer mode** (toggle in the top-right corner)
3. Click **Load unpacked**
4. Select the `browser-extension/` folder from this repo

The extension icon will appear in your toolbar.

### Install in Firefox

1. Open `about:debugging#/runtime/this-firefox`
2. Click **Load Temporary Add-on...**
3. Select `browser-extension/manifest.json`

> Note: Temporary add-ons are removed when Firefox restarts. For a persistent install, package the extension as a `.xpi` or use Firefox Developer Edition.

### Configure the extension

1. Click the extension icon → click **Settings** (bottom of the popup)
2. Enter your Overleaf base URL, e.g. `https://latex.yourdomain.com`
3. Click **Save**

### Connect a project to GitHub

1. Open a project in Overleaf
2. Click the extension icon
3. Click **Configure Project**
4. Fill in:
   - **GitHub Repo URL** — e.g. `https://github.com/you/my-thesis.git`
     (Create the repo on GitHub first; it can be empty)
   - **Branch** — e.g. `main`
   - **GitHub Token** — a PAT with `repo` scope ([create one here](https://github.com/settings/tokens))
5. Click **Save Config**

### Push and pull

- **Push to GitHub** — copies your current Overleaf files to the repo, prompts for a commit message, and pushes
- **Pull from GitHub** — pulls the latest commit and overwrites Overleaf files (you will be asked to confirm)

The popup shows the repo URL, branch, last sync time, and latest commit SHA.

---

## API Reference

The sync service exposes these endpoints under `/sync/`:

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/sync/projects/:id/config` | Set GitHub repo, branch, and token for a project |
| `GET`  | `/sync/projects/:id/status` | Get config and last sync info |
| `POST` | `/sync/push` | Commit Overleaf files and push to GitHub |
| `POST` | `/sync/pull` | Pull from GitHub and overwrite Overleaf files |
| `GET`  | `/sync/health` | Health check |

**POST /sync/push**
```json
{ "project_id": "abc123", "message": "Update introduction" }
```

**POST /sync/pull**
```json
{ "project_id": "abc123" }
```

**POST /sync/projects/:id/config**
```json
{ "repo_url": "https://github.com/you/repo.git", "branch": "main", "github_token": "ghp_..." }
```

---

## Project Structure

```
overleaf--with-git-sync/
├── docker-compose.yml
├── .env.example
├── sync-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── handlers/        # HTTP handlers
│   ├── gitops/          # git shell operations
│   ├── storage/         # metadata.json store
│   └── middleware/      # request logging
└── browser-extension/
    ├── manifest.json
    ├── popup.html / popup.js
    ├── options.html / options.js
    ├── content.js       # scaffold for future toolbar button
    └── icons/
```

---

## Sync Rules

- **Push** — Overleaf is the source of truth. GitHub is overwritten.
- **Pull** — GitHub is the source of truth. Overleaf files are overwritten.
- There is no merge. Last write wins.
- Save your work in Overleaf before pushing (the extension does not force a compile/save).

---

## License

MIT
