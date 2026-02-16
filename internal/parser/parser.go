package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Plugin represents a single entry from lazy-lock.json.
type Plugin struct {
	Name   string 
	Branch string 
	Commit string 
	Owner  string 
	Repo   string 
}

type lockEntry struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

// FindLockFile locates the lazy-lock.json file. It checks in order:
// 1. The provided path (if non-empty)
// 2. $XDG_CONFIG_HOME/nvim/lazy-lock.json
// 3. ~/.config/nvim/lazy-lock.json
func FindLockFile(path string) (string, error) {
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("lockfile not found at %s: %w", path, err)
		}
		return path, nil
	}

	candidates := []string{}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "nvim", "lazy-lock.json"))
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "nvim", "lazy-lock.json"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("lazy-lock.json not found; tried: %v", candidates)
}

func inferGitHubRepo(name string, overrides map[string]string) (owner, repo string) {
	// Check overrides first (from config scan)
	if override, ok := overrides[name]; ok {
		parts := strings.SplitN(override, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}


	repo = name
	owner = name
	// Strip common suffixes for the owner guess
	owner = strings.TrimSuffix(owner, ".nvim")
	owner = strings.TrimSuffix(owner, ".lua")

	return owner, repo
}

// Parse reads and parses a lazy-lock.json file, returning a slice of Plugins
// with inferred GitHub owner/repo information.
// configDir is the path to the neovim configuration directory (e.g. ~/.config/nvim).
// If empty, it defaults to ~/.config/nvim.
func Parse(lockPath string, configDir string) ([]Plugin, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("reading lockfile: %w", err)
	}

	var entries map[string]lockEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing lockfile JSON: %w", err)
	}

	// Scan config for plugin definitions
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "nvim")
	}

	overrides, err := ScanConfig(configDir)
	if err != nil {
		// 
	}

	plugins := make([]Plugin, 0, len(entries))
	for name, entry := range entries {
		owner, repo := inferGitHubRepo(name, overrides)
		plugins = append(plugins, Plugin{
			Name:   name,
			Branch: entry.Branch,
			Commit: entry.Commit,
			Owner:  owner,
			Repo:   repo,
		})
	}

	return plugins, nil
}

func ScanConfig(root string) (map[string]string, error) {
	overrides := make(map[string]string)

	re := regexp.MustCompile(`["']([a-zA-Z0-9_\-\.]+\/[a-zA-Z0-9_\-\.]+)["']`)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".lua") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		matches := re.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			if len(m) > 1 {
				full := m[1]
				parts := strings.Split(full, "/")
				if len(parts) == 2 {
					repo := parts[1]
			
					overrides[repo] = full
				}
			}
		}
		return nil
	})

	return overrides, err
}
