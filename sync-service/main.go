package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/handlers"
	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/middleware"
	"github.com/Juliiiian/overleaf--with-git-sync/sync-service/storage"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dataPath := getenv("OVERLEAF_DATA_PATH", "/var/lib/overleaf/data/user_files")
	configPath := getenv("SYNC_CONFIG_PATH", "/sync/config/metadata.json")
	reposPath := getenv("SYNC_REPOS_PATH", "/sync/repos")
	listenAddr := getenv("SYNC_LISTEN_ADDR", ":8080")

	store, err := storage.NewStore(configPath)
	if err != nil {
		log.Fatalf("failed to load metadata store: %v", err)
	}

	mux := http.NewServeMux()

	// Config & status
	mux.HandleFunc("POST /sync/projects/{id}/config", handlers.ConfigHandler(store))
	mux.HandleFunc("GET /sync/projects/{id}/status", handlers.StatusHandler(store))

	// Sync operations
	mux.HandleFunc("POST /sync/push", handlers.PushHandler(store, dataPath, reposPath))
	mux.HandleFunc("POST /sync/pull", handlers.PullHandler(store, dataPath, reposPath))

	// Health check
	mux.HandleFunc("GET /sync/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := middleware.Logging(mux)

	log.Printf("sync-service listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
