package handlers

import (
	"net/http"

	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/storage"
)

type statusResponse struct {
	ProjectID         string `json:"project_id"`
	Configured        bool   `json:"configured"`
	RepoURL           string `json:"repo_url,omitempty"`
	Branch            string `json:"branch,omitempty"`
	LastCommit        string `json:"last_commit,omitempty"`
	LastSync          string `json:"last_sync,omitempty"`
	LastSyncDirection string `json:"last_sync_direction,omitempty"`
}

func StatusHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		if projectID == "" {
			jsonError(w, "missing project id", http.StatusBadRequest)
			return
		}

		cfg, ok := store.Load(projectID)
		if !ok {
			jsonOK(w, statusResponse{ProjectID: projectID, Configured: false})
			return
		}

		resp := statusResponse{
			ProjectID:         projectID,
			Configured:        true,
			RepoURL:           cfg.RepoURL,
			Branch:            cfg.Branch,
			LastCommit:        cfg.LastCommit,
			LastSyncDirection: cfg.LastSyncDirection,
		}
		if !cfg.LastSync.IsZero() {
			resp.LastSync = cfg.LastSync.UTC().Format("2006-01-02T15:04:05Z")
		}

		jsonOK(w, resp)
	}
}
