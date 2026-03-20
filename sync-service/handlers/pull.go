package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/gitops"
	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/storage"
)

type pullRequest struct {
	ProjectID string `json:"project_id"`
}

func PullHandler(store *storage.Store, dataPath, reposPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req pullRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.ProjectID == "" {
			jsonError(w, "project_id is required", http.StatusBadRequest)
			return
		}

		cfg, ok := store.Load(req.ProjectID)
		if !ok {
			jsonError(w, "project not configured — POST /sync/projects/"+req.ProjectID+"/config first", http.StatusConflict)
			return
		}

		dstDir := filepath.Join(dataPath, req.ProjectID)
		repoDir := filepath.Join(reposPath, req.ProjectID)

		sha, fileCount, err := gitops.Pull(repoDir, dstDir, cfg.RepoURL, cfg.Branch, cfg.GitHubToken)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cfg.LastCommit = sha
		cfg.LastSync = time.Now().UTC()
		cfg.LastSyncDirection = "pull"
		_ = store.Save(req.ProjectID, cfg)

		jsonOK(w, map[string]any{
			"commit_sha":    sha,
			"updated_files": fileCount,
			"branch":        cfg.Branch,
		})
	}
}
