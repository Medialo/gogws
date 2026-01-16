package gws

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspaceRoot_ProjectsOnly(t *testing.T) {
	tempDir := t.TempDir()

	projectsFile := filepath.Join(tempDir, ProjectsFileName)
	if err := os.WriteFile(projectsFile, []byte("test | url"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	ws, err := FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("FindWorkspaceRoot failed: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Expected workspace root %s, got %s", tempDir, ws.Root)
	}
	if !ws.HasProjectsFile {
		t.Error("Expected HasProjectsFile to be true")
	}
	if ws.HasWorkspacesFile {
		t.Error("Expected HasWorkspacesFile to be false")
	}
}

func TestFindWorkspaceRoot_WorkspacesOnly(t *testing.T) {
	tempDir := t.TempDir()

	workspacesFile := filepath.Join(tempDir, WorkspacesFileName)
	if err := os.WriteFile(workspacesFile, []byte("ws | url"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	ws, err := FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("FindWorkspaceRoot failed: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Expected workspace root %s, got %s", tempDir, ws.Root)
	}
	if ws.HasProjectsFile {
		t.Error("Expected HasProjectsFile to be false")
	}
	if !ws.HasWorkspacesFile {
		t.Error("Expected HasWorkspacesFile to be true")
	}
}

func TestFindWorkspaceRoot_Both(t *testing.T) {
	tempDir := t.TempDir()

	projectsFile := filepath.Join(tempDir, ProjectsFileName)
	if err := os.WriteFile(projectsFile, []byte("test | url"), 0644); err != nil {
		t.Fatal(err)
	}

	workspacesFile := filepath.Join(tempDir, WorkspacesFileName)
	if err := os.WriteFile(workspacesFile, []byte("ws | url"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	ws, err := FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("FindWorkspaceRoot failed: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Expected workspace root %s, got %s", tempDir, ws.Root)
	}
	if !ws.HasProjectsFile {
		t.Error("Expected HasProjectsFile to be true")
	}
	if !ws.HasWorkspacesFile {
		t.Error("Expected HasWorkspacesFile to be true")
	}
}

func TestFindWorkspaceRoot_NoFile(t *testing.T) {
	tempDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	_, err = FindWorkspaceRoot()
	if err == nil {
		t.Error("Expected error when no workspace files found, got nil")
	}
}
