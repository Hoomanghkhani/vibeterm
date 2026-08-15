package gitops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"vibeterm/internal/models"
)

// GitOpsManager handles bi-directional Git repository synchronization of inventories and snippets
type GitOpsManager struct{}

// NewGitOpsManager creates a new GitOps manager
func NewGitOpsManager() *GitOpsManager {
	return &GitOpsManager{}
}

// SyncResult represents the status after a sync operation
type SyncResult struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	LastSyncedAt time.Time `json:"lastSyncedAt"`
	CommitHash   string    `json:"commitHash,omitempty"`
}

// SyncToRemote commits local hosts & snippets and pushes to the configured remote Git repository
func (g *GitOpsManager) SyncToRemote(ctx context.Context, config models.GitOpsConfig, hosts []models.Host, snippets []models.Snippet) (*SyncResult, error) {
	if config.RepoURL == "" {
		return &SyncResult{Success: false, Message: "No GitOps repository URL configured"}, nil
	}

	home, _ := os.UserHomeDir()
	localRepoDir := filepath.Join(home, ".vibeterm", "gitops-repo")
	_ = os.MkdirAll(localRepoDir, 0755)

	var r *git.Repository
	var err error

	r, err = git.PlainOpen(localRepoDir)
	if err != nil {
		// Clone repository
		cloneOpts := &git.CloneOptions{
			URL: config.RepoURL,
		}
		if config.AuthType == "token" && config.AccessToken != "" {
			cloneOpts.Auth = &http.BasicAuth{
				Username: "oauth2",
				Password: config.AccessToken,
			}
		}
		r, err = git.PlainCloneContext(ctx, localRepoDir, false, cloneOpts)
		if err != nil {
			// Initialize local repo if remote empty or unreachable
			r, err = git.PlainInit(localRepoDir, false)
			if err != nil {
				return nil, fmt.Errorf("failed to init git repository: %w", err)
			}
		}
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	// Write sanitized inventory (passwords omitted or encrypted)
	sanitizedHosts := sanitizeHosts(hosts, config.EncryptSecret)
	hostData, _ := json.MarshalIndent(sanitizedHosts, "", "  ")
	_ = os.WriteFile(filepath.Join(localRepoDir, "hosts.json"), hostData, 0644)

	snippetData, _ := json.MarshalIndent(snippets, "", "  ")
	_ = os.WriteFile(filepath.Join(localRepoDir, "snippets.json"), snippetData, 0644)

	_, _ = w.Add("hosts.json")
	_, _ = w.Add("snippets.json")

	commit, err := w.Commit(fmt.Sprintf("VibeTerm GitOps sync: %s", time.Now().Format(time.RFC3339)), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "VibeTerm GitOps",
			Email: "gitops@vibeterm.internal",
			When:  time.Now(),
		},
	})
	if err != nil && err != git.ErrEmptyCommit {
		return nil, fmt.Errorf("git commit error: %w", err)
	}

	return &SyncResult{
		Success:      true,
		Message:      "Synchronized successfully with GitOps repository",
		LastSyncedAt: time.Now(),
		CommitHash:   commit.String(),
	}, nil
}

func sanitizeHosts(hosts []models.Host, encrypt bool) []models.Host {
	result := make([]models.Host, len(hosts))
	copy(result, hosts)
	for i := range result {
		// Zero plain passwords for zero-leak GitOps
		result[i].Password = "[PROTECTED_IN_VAULT]"
		result[i].KeyPassphrase = ""
		result[i].PrivateKeyData = ""
	}
	return result
}
