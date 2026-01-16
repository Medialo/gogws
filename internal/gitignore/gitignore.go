package gitignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	StartMarker = "# === GWS START ==="
	EndMarker   = "# === GWS END ==="
)

func Generate() (string, error) {
	return renderTemplate(DefaultData())
}

func GenerateWithData(data TemplateData) (string, error) {
	return renderTemplate(data)
}

func GenerateWithMarkers() (string, error) {
	content, err := Generate()
	if err != nil {
		return "", err
	}
	return wrapWithMarkers(content), nil
}

func GenerateWithMarkersAndData(data TemplateData) (string, error) {
	content, err := GenerateWithData(data)
	if err != nil {
		return "", err
	}
	return wrapWithMarkers(content), nil
}

func wrapWithMarkers(content string) string {
	return fmt.Sprintf("%s\n%s%s\n", StartMarker, content, EndMarker)
}

func EnsureGWSSection(dir string) error {
	return EnsureGWSSectionWithData(dir, DefaultData())
}

func EnsureGWSSectionWithData(dir string, data TemplateData) error {
	filePath := filepath.Join(dir, ".gitignore")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return CreateGitignore(dir)
	}

	hasSection, err := HasGWSSection(filePath)
	if err != nil {
		return err
	}

	if hasSection {
		return UpdateGWSSection(filePath, data)
	}

	return AppendGWSSection(filePath, data)
}

func HasGWSSection(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == StartMarker {
			return true, nil
		}
	}

	return false, scanner.Err()
}

func CreateGitignore(dir string) error {
	return CreateGitignoreWithData(dir, DefaultData())
}

func CreateGitignoreWithData(dir string, data TemplateData) error {
	filePath := filepath.Join(dir, ".gitignore")

	content, err := GenerateWithMarkersAndData(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

func AppendGWSSection(filePath string, data TemplateData) error {
	content, err := GenerateWithMarkersAndData(data)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	prefix := "\n"
	if len(existingContent) > 0 && !strings.HasSuffix(string(existingContent), "\n") {
		prefix = "\n\n"
	} else if len(existingContent) > 0 {
		prefix = "\n"
	}

	_, err = file.WriteString(prefix + content)
	return err
}

func UpdateGWSSection(filePath string, data TemplateData) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	newSection, err := GenerateWithMarkersAndData(data)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var result []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == StartMarker {
			inSection = true
			continue
		}

		if trimmed == EndMarker {
			inSection = false
			result = append(result, strings.TrimSuffix(newSection, "\n"))
			continue
		}

		if !inSection {
			result = append(result, line)
		}
	}

	return os.WriteFile(filePath, []byte(strings.Join(result, "\n")), 0644)
}

func RemoveGWSSection(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var result []string
	inSection := false
	removedNewlineBefore := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == StartMarker {
			inSection = true
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" && len(result) > 0 {
				result = result[:len(result)-1]
				removedNewlineBefore = true
			}
			continue
		}

		if trimmed == EndMarker {
			inSection = false
			continue
		}

		if !inSection {
			result = append(result, line)
		}
	}

	_ = removedNewlineBefore

	finalContent := strings.Join(result, "\n")
	finalContent = strings.TrimRight(finalContent, "\n") + "\n"

	return os.WriteFile(filePath, []byte(finalContent), 0644)
}
