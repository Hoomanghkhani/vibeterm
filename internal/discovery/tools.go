package discovery

import (
	"os/exec"
	"strings"
)

type DetectedTool struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Version   string `json:"version"`
}

type ToolDetector struct{}

func NewToolDetector() *ToolDetector {
	return &ToolDetector{}
}

func (td *ToolDetector) DetectAll() []DetectedTool {
	tools := []string{"ssh", "scp", "sftp", "docker", "kubectl", "helm", "tailscale", "rsync", "git"}
	var results []DetectedTool

	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			results = append(results, DetectedTool{
				Name:      tool,
				Installed: false,
			})
			continue
		}

		// Try getting version
		version := "detected"
		if out, err := exec.Command(tool, "--version").Output(); err == nil {
			firstLine := strings.Split(strings.TrimSpace(string(out)), "\n")[0]
			if len(firstLine) > 50 {
				firstLine = firstLine[:50]
			}
			version = firstLine
		} else if out, err := exec.Command(tool, "version").Output(); err == nil {
			firstLine := strings.Split(strings.TrimSpace(string(out)), "\n")[0]
			if len(firstLine) > 50 {
				firstLine = firstLine[:50]
			}
			version = firstLine
		}

		results = append(results, DetectedTool{
			Name:      tool,
			Installed: true,
			Path:      path,
			Version:   version,
		})
	}

	return results
}
