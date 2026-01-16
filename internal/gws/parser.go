package gws

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gogws/internal/log"
)

type Parser struct {
	root string
}

func NewParser(root string) *Parser {
	return &Parser{root: root}
}

func (p *Parser) Root() string {
	return p.root
}

func (p *Parser) HasProjectsFile() bool {
	path := filepath.Join(p.root, ProjectsFileName)
	_, err := os.Stat(path)
	return err == nil
}

func (p *Parser) HasWorkspacesFile() bool {
	path := filepath.Join(p.root, WorkspacesFileName)
	_, err := os.Stat(path)
	return err == nil
}

func (p *Parser) HasIgnoreFile() bool {
	path := filepath.Join(p.root, IgnoreFileName)
	_, err := os.Stat(path)
	return err == nil
}

func (p *Parser) ParseProjects() ([]Project, error) {
	projectsPath := filepath.Join(p.root, ProjectsFileName)
	log.Debug("Reading projects", "path", projectsPath)

	file, err := os.Open(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open projects file: %w", err)
	}
	defer file.Close()

	var projects []Project
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		project, err := parseProjectLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing line %d: %w", lineNum, err)
		}

		projects = append(projects, project)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading projects file: %w", err)
	}

	ignorePatterns, err := p.ParseIgnorePatterns()
	if err == nil && len(ignorePatterns) > 0 {
		projects = filterIgnoredProjects(projects, ignorePatterns)
	}

	log.Debug("Found projects", "count", len(projects), "root", p.root)
	return projects, nil
}

func (p *Parser) ParseWorkspaces() ([]WorkspaceRef, error) {
	workspacesPath := filepath.Join(p.root, WorkspacesFileName)
	log.Debug("Reading workspaces", "path", workspacesPath)

	file, err := os.Open(workspacesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []WorkspaceRef{}, nil
		}
		return nil, fmt.Errorf("failed to open workspaces file: %w", err)
	}
	defer file.Close()

	var workspaces []WorkspaceRef
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		ws, err := parseWorkspaceLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing line %d in %s: %w", lineNum, WorkspacesFileName, err)
		}

		workspaces = append(workspaces, ws)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading workspaces file: %w", err)
	}

	log.Debug("Found workspaces", "count", len(workspaces), "root", p.root)
	return workspaces, nil
}

func (p *Parser) ParseIgnorePatterns() ([]string, error) {
	ignorePath := filepath.Join(p.root, IgnoreFileName)
	file, err := os.Open(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open ignore file: %w", err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ignore file: %w", err)
	}

	return patterns, nil
}

func parseProjectLine(line string) (Project, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return Project{}, fmt.Errorf("invalid format: expected 'path | url [name] [| url2 name2 ...]'")
	}

	path := strings.TrimSpace(parts[0])
	if path == "" {
		return Project{}, fmt.Errorf("empty project path")
	}

	project := Project{
		Path:    path,
		Remotes: make([]Remote, 0),
	}

	for i := 1; i < len(parts); i++ {
		remotePart := strings.TrimSpace(parts[i])
		if remotePart == "" {
			continue
		}

		remote, err := parseRemote(remotePart, i-1)
		if err != nil {
			return Project{}, err
		}
		project.Remotes = append(project.Remotes, remote)
	}

	if len(project.Remotes) == 0 {
		return Project{}, fmt.Errorf("no remotes defined for project %s", path)
	}

	return project, nil
}

func parseWorkspaceLine(line string) (WorkspaceRef, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return WorkspaceRef{}, fmt.Errorf("invalid format: expected 'path | url [name]'")
	}

	path := strings.TrimSpace(parts[0])
	if path == "" {
		return WorkspaceRef{}, fmt.Errorf("empty workspace path")
	}

	remotePart := strings.TrimSpace(parts[1])
	if remotePart == "" {
		return WorkspaceRef{}, fmt.Errorf("empty remote URL for workspace %s", path)
	}

	remote, err := parseRemote(remotePart, 0)
	if err != nil {
		return WorkspaceRef{}, err
	}

	return WorkspaceRef{
		Path:   path,
		Remote: remote,
	}, nil
}

func parseRemote(remotePart string, index int) (Remote, error) {
	fields := strings.Fields(remotePart)
	if len(fields) == 0 {
		return Remote{}, fmt.Errorf("empty remote definition")
	}

	url := fields[0]
	name := "origin"

	if index == 0 && len(fields) > 1 {
		name = fields[1]
	} else if index == 1 {
		name = "upstream"
		if len(fields) > 1 {
			name = fields[1]
		}
	} else if len(fields) > 1 {
		name = fields[1]
	}

	return Remote{
		Name: name,
		URL:  url,
	}, nil
}

func filterIgnoredProjects(projects []Project, patterns []string) []Project {
	if len(patterns) == 0 {
		return projects
	}

	regexps := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			regexps = append(regexps, re)
		}
	}

	filtered := make([]Project, 0, len(projects))
	for _, project := range projects {
		ignored := false
		for _, re := range regexps {
			if re.MatchString(project.Path) {
				ignored = true
				log.Debug("Ignoring project", "path", project.Path, "reason", "matches pattern")
				break
			}
		}
		if !ignored {
			filtered = append(filtered, project)
		}
	}

	return filtered
}
