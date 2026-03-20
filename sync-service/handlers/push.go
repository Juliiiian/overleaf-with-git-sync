package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/gitops"
	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/storage"
)

type pushRequest struct {
	ProjectID string `json:"project_id"`
	Message   string `json:"message"`
}

func PushHandler(store *storage.Store, dataPath, reposPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req pushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.ProjectID == "" {
			jsonError(w, "project_id is required", http.StatusBadRequest)
			return
		}
		if req.Message == "" {
			req.Message = "Sync from Overleaf"
		}

		cfg, ok := store.Load(req.ProjectID)
		if !ok {
			jsonError(w, "project not configured — POST /sync/projects/"+req.ProjectID+"/config first", http.StatusConflict)
			return
		}

		srcDir := filepath.Join(dataPath, req.ProjectID)
		repoDir := filepath.Join(reposPath, req.ProjectID)

		sha, fileCount, err := gitops.Push(repoDir, srcDir, cfg.RepoURL, cfg.Branch, cfg.GitHubToken, req.Message)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cfg.LastCommit = sha
		cfg.LastSync = time.Now().UTC()
		cfg.LastSyncDirection = "push"
		_ = store.Save(req.ProjectID, cfg)

		jsonOK(w, map[string]any{
			"commit_sha":   sha,
			"pushed_files": fileCount,
			"branch":       cfg.Branch,
		})
	}
}
