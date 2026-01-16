package export

import (
	"encoding/json"
	"fmt"

	"gogws/internal/git"

	"gopkg.in/yaml.v3"
)

type StatusOutput struct {
	Total        int                      `json:"total" yaml:"total"`
	Clean        int                      `json:"clean" yaml:"clean"`
	Changed      int                      `json:"changed" yaml:"changed"`
	Missing      int                      `json:"missing" yaml:"missing"`
	Errors       int                      `json:"errors" yaml:"errors"`
	Repositories []RepositoryStatusOutput `json:"repositories" yaml:"repositories"`
}

type RepositoryStatusOutput struct {
	Path        string `json:"path" yaml:"path"`
	Exists      bool   `json:"exists" yaml:"exists"`
	Clean       bool   `json:"clean" yaml:"clean"`
	Branch      string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Ahead       int    `json:"ahead" yaml:"ahead"`
	Behind      int    `json:"behind" yaml:"behind"`
	Uncommitted int    `json:"uncommitted" yaml:"uncommitted"`
	Untracked   int    `json:"untracked" yaml:"untracked"`
	HasRemote   bool   `json:"has_remote" yaml:"has_remote"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

func ToJSON(statuses []git.RepositoryStatus) (string, error) {
	output := buildOutput(statuses)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ToYAML(statuses []git.RepositoryStatus) (string, error) {
	output := buildOutput(statuses)
	data, err := yaml.Marshal(output)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildOutput(statuses []git.RepositoryStatus) StatusOutput {
	output := StatusOutput{
		Total:        len(statuses),
		Repositories: make([]RepositoryStatusOutput, len(statuses)),
	}

	for i, status := range statuses {
		repoOutput := RepositoryStatusOutput{
			Path:        status.Path,
			Exists:      status.Exists,
			Clean:       status.Clean,
			Branch:      status.Branch,
			Ahead:       status.Ahead,
			Behind:      status.Behind,
			Uncommitted: status.Uncommitted,
			Untracked:   status.Untracked,
			HasRemote:   status.HasRemote,
		}

		if status.Error != nil {
			repoOutput.Error = status.Error.Error()
			output.Errors++
		}

		if !status.Exists {
			output.Missing++
		} else if status.Clean && status.Ahead == 0 && status.Behind == 0 {
			output.Clean++
		} else {
			output.Changed++
		}

		output.Repositories[i] = repoOutput
	}

	return output
}

func Format(statuses []git.RepositoryStatus, format string) (string, error) {
	switch format {
	case "json":
		return ToJSON(statuses)
	case "yaml":
		return ToYAML(statuses)
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: json, yaml)", format)
	}
}
