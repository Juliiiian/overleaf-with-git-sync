package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/storage"
)

type configRequest struct {
	RepoURL     string `json:"repo_url"`
	Branch      string `json:"branch"`
	GitHubToken string `json:"github_token"`
}

func ConfigHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		if projectID == "" {
			jsonError(w, "missing project id", http.StatusBadRequest)
			return
		}

		var req configRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.RepoURL == "" || req.Branch == "" || req.GitHubToken == "" {
			jsonError(w, "repo_url, branch, and github_token are required", http.StatusBadRequest)
			return
		}

		cfg, _ := store.Load(projectID)
		cfg.RepoURL = req.RepoURL
		cfg.Branch = req.Branch
		cfg.GitHubToken = req.GitHubToken

		if err := store.Save(projectID, cfg); err != nil {
			jsonError(w, "failed to save config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		jsonOK(w, map[string]string{"project_id": projectID, "status": "configured"})
	}
}
