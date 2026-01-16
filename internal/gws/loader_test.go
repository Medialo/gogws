package gws

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot_ProjectsOnly(t *testing.T) {
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

	ws, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Expected workspace root %s, got %s", tempDir, ws.Root)
	}
}

func TestFindRoot_WorkspacesOnly(t *testing.T) {
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

	ws, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Expected workspace root %s, got %s", tempDir, ws.Root)
	}
}

func TestFindRoot_Both(t *testing.T) {
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

	ws, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Expected workspace root %s, got %s", tempDir, ws.Root)
	}
}

func TestFindRoot_NoFile(t *testing.T) {
	tempDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	_, err = FindRoot()
	if err == nil {
		t.Error("Expected error when no workspace files found, got nil")
	}
}

func TestLoader_Load(t *testing.T) {
	tempDir := t.TempDir()

	projectsFile := filepath.Join(tempDir, ProjectsFileName)
	content := "project1 | git@github.com:user/repo1.git\nproject2 | git@github.com:user/repo2.git"
	if err := os.WriteFile(projectsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ws, err := New(tempDir).Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ws.Projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(ws.Projects))
	}

	if ws.Projects[0].Path != "project1" {
		t.Errorf("Expected project1, got %s", ws.Projects[0].Path)
	}
}

func TestLoader_LoadRecursive(t *testing.T) {
	tempDir := t.TempDir()

	workspacesFile := filepath.Join(tempDir, WorkspacesFileName)
	if err := os.WriteFile(workspacesFile, []byte("child | git@github.com:user/child.git"), 0644); err != nil {
		t.Fatal(err)
	}

	childDir := filepath.Join(tempDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	childProjectsFile := filepath.Join(childDir, ProjectsFileName)
	if err := os.WriteFile(childProjectsFile, []byte("subproject | git@github.com:user/sub.git"), 0644); err != nil {
		t.Fatal(err)
	}

	ws, err := New(tempDir).Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ws.Children) != 1 {
		t.Fatalf("Expected 1 child workspace, got %d", len(ws.Children))
	}

	if !ws.Children[0].Exists {
		t.Error("Expected child workspace to exist")
	}

	if len(ws.Children[0].Projects) != 1 {
		t.Errorf("Expected 1 project in child, got %d", len(ws.Children[0].Projects))
	}
}

func TestLoader_NonRecursive(t *testing.T) {
	tempDir := t.TempDir()

	workspacesFile := filepath.Join(tempDir, WorkspacesFileName)
	if err := os.WriteFile(workspacesFile, []byte("child | git@github.com:user/child.git"), 0644); err != nil {
		t.Fatal(err)
	}

	childDir := filepath.Join(tempDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	childProjectsFile := filepath.Join(childDir, ProjectsFileName)
	if err := os.WriteFile(childProjectsFile, []byte("subproject | git@github.com:user/sub.git"), 0644); err != nil {
		t.Fatal(err)
	}

	ws, err := New(tempDir).Recursive(false).Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ws.Children) != 1 {
		t.Fatalf("Expected 1 child workspace ref, got %d", len(ws.Children))
	}

	if len(ws.Children[0].Projects) != 0 {
		t.Errorf("Expected 0 projects in non-recursive child, got %d", len(ws.Children[0].Projects))
	}
}
