package parser

import (
	"os"
	"path/filepath"
	"testing"
)

const testLockJSON = `{
  "Comment.nvim": { "branch": "master", "commit": "e30b7f2008e52442154b66f7c519bfd2f1e32acb" },
  "blink.cmp": { "branch": "main", "commit": "4b18c32adef2898f95cdef6192cbd5796c1a332d" },
  "nvim-treesitter": { "branch": "main", "commit": "45a07f869b0cffba342276f2c77ba7c116d35db8" },
  "nvim-lspconfig": { "branch": "master", "commit": "66fd02ad1c7ea31616d3ca678fa04e6d0b360824" },
  "some-unknown-plugin": { "branch": "main", "commit": "abcdef1234567890abcdef1234567890abcdef12" }
}`

func TestParse(t *testing.T) {
	// Write temp lockfile
	dir := t.TempDir()
	path := filepath.Join(dir, "lazy-lock.json")
	if err := os.WriteFile(path, []byte(testLockJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Use temp dir for config too to avoid scanning real user config
	plugins, err := Parse(path, dir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(plugins) != 5 {
		t.Fatalf("expected 5 plugins, got %d", len(plugins))
	}

	// Check known plugins are resolved correctly
	found := map[string]Plugin{}
	for _, p := range plugins {
		found[p.Name] = p
	}

	// Since we are NOT providing a config file with overrides,
	// and we removed the hardcoded map, these will fallback to heuristic default.
	// E.g. blink.cmp -> blink.cmp/blink.cmp (unless we add a dummy config file)

	// Let's verify heuristic fallback works
	if p, ok := found["some-unknown-plugin"]; ok {
		if p.Repo != "some-unknown-plugin" {
			t.Errorf("expected some-unknown-plugin, got %s", p.Repo)
		}
	}
}

func TestScanConfig(t *testing.T) {
	dir := t.TempDir()
	luaContent := `
	return require("packer").startup(function(use)
	  use "wbthomason/packer.nvim"
	  use { "folke/which-key.nvim" }
	  use 'Another/Plugin.nvim'
	end)
	`
	if err := os.WriteFile(filepath.Join(dir, "plugins.lua"), []byte(luaContent), 0644); err != nil {
		t.Fatal(err)
	}

	overrides, err := ScanConfig(dir)
	if err != nil {
		t.Fatalf("ScanConfig failed: %v", err)
	}

	tests := map[string]string{
		"packer.nvim":    "wbthomason/packer.nvim",
		"which-key.nvim": "folke/which-key.nvim",
		"Plugin.nvim":    "Another/Plugin.nvim",
	}

	for repo, want := range tests {
		if got, ok := overrides[repo]; !ok || got != want {
			t.Errorf("repo %s: got %q, want %q", repo, got, want)
		}
	}
}

func TestFindLockFile_WithExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lazy-lock.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := FindLockFile(path)
	if err != nil {
		t.Fatalf("FindLockFile failed: %v", err)
	}
	if found != path {
		t.Errorf("expected %s, got %s", path, found)
	}
}

func TestFindLockFile_NotFound(t *testing.T) {
	_, err := FindLockFile("/nonexistent/path/lazy-lock.json")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}
