package gitops

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// projectLocks holds a per-project mutex to prevent concurrent git operations.
var projectLocks sync.Map

func lockProject(projectID string) func() {
	v, _ := projectLocks.LoadOrStore(projectID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// authedURL injects the token into an HTTPS GitHub URL.
func authedURL(repoURL, token string) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-access-token", token)
	return u.String(), nil
}

// git runs a git command in dir, returning combined output or an error.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, buf.String())
	}
	return strings.TrimSpace(buf.String()), nil
}

// EnsureRepo clones the repo if it doesn't exist, otherwise fetches and resets to origin/<branch>.
func EnsureRepo(repoDir, repoURL, branch, token string) error {
	authed, err := authedURL(repoURL, token)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(repoDir), 0755); err != nil {
			return err
		}
		if _, err := git("", "clone", "--branch", branch, authed, repoDir); err != nil {
			// Branch may not exist yet (empty repo). Clone without --branch and create it.
			if err2 := cloneEmpty(repoDir, authed, branch); err2 != nil {
				return fmt.Errorf("clone failed: %w; empty-clone also failed: %v", err, err2)
			}
		}
		return nil
	}

	// Repo exists — update remote URL (token may have changed) then fetch+reset.
	if _, err := git(repoDir, "remote", "set-url", "origin", authed); err != nil {
		return err
	}
	if _, err := git(repoDir, "fetch", "origin"); err != nil {
		return err
	}
	// Check if remote branch exists.
	out, err := git(repoDir, "ls-remote", "--heads", "origin", branch)
	if err != nil {
		return err
	}
	if out == "" {
		// Remote branch doesn't exist yet; create it locally if needed.
		return nil
	}
	if _, err := git(repoDir, "reset", "--hard", "origin/"+branch); err != nil {
		return err
	}
	return nil
}

// cloneEmpty clones a potentially empty repo and sets up the branch.
func cloneEmpty(repoDir, authed, branch string) error {
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}
	if _, err := git(repoDir, "init"); err != nil {
		return err
	}
	if _, err := git(repoDir, "remote", "add", "origin", authed); err != nil {
		return err
	}
	if _, err := git(repoDir, "checkout", "-b", branch); err != nil {
		return err
	}
	return nil
}

// CopyToRepo copies all files from srcDir into repoDir, skipping .git.
// Returns the count of files copied.
func CopyToRepo(srcDir, repoDir string) (int, error) {
	count := 0
	return count, filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(repoDir, rel), 0755)
		}
		count++
		return copyFile(path, filepath.Join(repoDir, rel))
	})
}

// CopyFromRepo copies all files from repoDir into dstDir, skipping .git.
// Deletes files in dstDir that no longer exist in repoDir.
// Returns the count of files updated.
func CopyFromRepo(repoDir, dstDir string) (int, error) {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return 0, err
	}

	// Collect repo files (relative paths).
	repoFiles := make(map[string]struct{})
	count := 0
	err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repoDir, path)
		if rel == ".git" || strings.HasPrefix(rel, ".git/") || strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dstDir, rel), 0755)
		}
		repoFiles[rel] = struct{}{}
		count++
		return copyFile(path, filepath.Join(dstDir, rel))
	})
	if err != nil {
		return count, err
	}

	// Remove files in dst that are no longer in repo.
	_ = filepath.WalkDir(dstDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dstDir, path)
		if _, ok := repoFiles[rel]; !ok {
			_ = os.Remove(path)
		}
		return nil
	})

	return count, nil
}

// CommitAndPush stages all changes, commits with message, and pushes.
// Returns the commit SHA. If there is nothing to commit, returns ("", nil).
func CommitAndPush(repoDir, branch, message string) (string, error) {
	if _, err := git(repoDir, "add", "-A"); err != nil {
		return "", err
	}

	// Check if there's anything to commit.
	status, err := git(repoDir, "status", "--porcelain")
	if err != nil {
		return "", err
	}
	if status == "" {
		// Nothing to commit — return current HEAD SHA.
		sha, _ := git(repoDir, "rev-parse", "HEAD")
		return sha, nil
	}

	if _, err := git(repoDir, "commit", "-m", message); err != nil {
		return "", err
	}

	if _, err := git(repoDir, "push", "origin", branch); err != nil {
		return "", err
	}

	sha, err := git(repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return sha, nil
}

// HeadSHA returns the current HEAD commit SHA of the repo.
func HeadSHA(repoDir string) (string, error) {
	return git(repoDir, "rev-parse", "HEAD")
}

// Push performs the full push flow: ensure repo, copy files, commit+push.
func Push(repoDir, srcDir, repoURL, branch, token, message string) (sha string, fileCount int, err error) {
	unlock := lockProject(repoDir)
	defer unlock()

	if err = EnsureRepo(repoDir, repoURL, branch, token); err != nil {
		return
	}

	fileCount, err = CopyToRepo(srcDir, repoDir)
	if err != nil {
		return
	}

	sha, err = CommitAndPush(repoDir, branch, message)
	return
}

// Pull performs the full pull flow: ensure repo, copy files to dst.
func Pull(repoDir, dstDir, repoURL, branch, token string) (sha string, fileCount int, err error) {
	unlock := lockProject(repoDir)
	defer unlock()

	if err = EnsureRepo(repoDir, repoURL, branch, token); err != nil {
		return
	}

	fileCount, err = CopyFromRepo(repoDir, dstDir)
	if err != nil {
		return
	}

	sha, err = HeadSHA(repoDir)
	return
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
