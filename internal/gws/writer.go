package gws

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func AddProject(workspaceRoot string, project Project) error {
	location := getProjectsFileLocation(workspaceRoot)
	if location == nil {
		gwsDir := filepath.Join(workspaceRoot, ConfigDirName)
		if err := os.MkdirAll(gwsDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", ConfigDirName, err)
		}
		location = &FileLocation{
			Path:        filepath.Join(gwsDir, "projects."+FileExtension),
			IsConfigDir: true,
		}
	}

	file, err := os.OpenFile(location.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open projects file: %w", err)
	}
	defer file.Close()

	line := formatProjectLine(project)
	if _, err := file.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to projects file: %w", err)
	}

	return nil
}

func RemoveProject(workspaceRoot string, projectPath string) error {
	location := getProjectsFileLocation(workspaceRoot)
	if location == nil {
		return fmt.Errorf("no projects file found")
	}

	return removeLineFromFile(location.Path, func(line string) bool {
		parts := strings.Split(line, "|")
		if len(parts) < 1 {
			return false
		}
		linePath := strings.TrimSpace(parts[0])
		return linePath == projectPath
	})
}

func AddWorkspace(workspaceRoot string, ws *Workspace) error {
	location := getWorkspacesFileLocation(workspaceRoot)
	if location == nil {
		gwsDir := filepath.Join(workspaceRoot, ConfigDirName)
		if err := os.MkdirAll(gwsDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", ConfigDirName, err)
		}
		location = &FileLocation{
			Path:        filepath.Join(gwsDir, "workspaces."+FileExtension),
			IsConfigDir: true,
		}
	}

	file, err := os.OpenFile(location.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open workspaces file: %w", err)
	}
	defer file.Close()

	line := formatWorkspaceLine(ws)
	if _, err := file.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to workspaces file: %w", err)
	}

	return nil
}

func RemoveWorkspace(workspaceRoot string, workspacePath string) error {
	location := getWorkspacesFileLocation(workspaceRoot)
	if location == nil {
		return fmt.Errorf("no workspaces file found")
	}

	return removeLineFromFile(location.Path, func(line string) bool {
		parts := strings.Split(line, "|")
		if len(parts) < 1 {
			return false
		}
		linePath := strings.TrimSpace(parts[0])
		return linePath == workspacePath
	})
}

func formatProjectLine(project Project) string {
	var remoteParts []string
	for _, remote := range project.Remotes {
		if remote.Name == "origin" {
			remoteParts = append(remoteParts, remote.URL)
		} else {
			remoteParts = append(remoteParts, fmt.Sprintf("%s %s", remote.URL, remote.Name))
		}
	}
	return fmt.Sprintf("%s | %s", project.Path, strings.Join(remoteParts, " | "))
}

func formatWorkspaceLine(ws *Workspace) string {
	if ws.Remote.Name == "origin" || ws.Remote.Name == "" {
		return fmt.Sprintf("%s | %s", ws.Path, ws.Remote.URL)
	}
	return fmt.Sprintf("%s | %s %s", ws.Path, ws.Remote.URL, ws.Remote.Name)
}

func removeLineFromFile(filePath string, shouldRemove func(line string) bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}

		cleanLine := trimmed
		if idx := strings.Index(cleanLine, "#"); idx != -1 {
			cleanLine = strings.TrimSpace(cleanLine[:idx])
		}

		if !shouldRemove(cleanLine) {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		file.Close()
		return fmt.Errorf("error reading file: %w", err)
	}
	file.Close()

	output, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer output.Close()

	for _, line := range lines {
		if _, err := output.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}
