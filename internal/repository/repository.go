package repository

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Repository struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Editor string `json:"editor,omitempty"`
}

func FindRepositories() ([]Repository, error) {
	githubFolder := filepath.Join(os.Getenv("HOME"), "Documents", "GitHub")

	if runtime.GOOS == "windows" {
		githubFolder = filepath.Join(os.Getenv("USERPROFILE"), "Documents", "GitHub")
	}

	entries, err := os.ReadDir(githubFolder)
	if err != nil {
		return nil, fmt.Errorf("error reading GitHub folder: %w", err)
	}

	var repos []Repository
	for _, entry := range entries {
		if entry.IsDir() {
			repoPath := filepath.Join(githubFolder, entry.Name())
			repos = append(repos, Repository{
				Path: repoPath,
				Name: entry.Name(),
			})
		}
	}

	return repos, nil
}

func (r *Repository) Open(editor string) error {
	if editor == "" {
		editor = getDefaultEditor()
	}

	cmd := exec.Command(editor, r.Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getDefaultEditor() string {
	switch runtime.GOOS {
	case "windows":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "notepad.exe"
		}
		return editor
	case "linux", "darwin":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}
		return editor
	default:
		return "nano"
	}
}
