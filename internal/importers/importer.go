package importers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"vibeterm/internal/models"
)

type ImportResult struct {
	ImportedCount int      `json:"importedCount"`
	Hosts         []models.Host `json:"hosts"`
	Errors        []string `json:"errors,omitempty"`
}

type SessionImporter struct{}

func NewSessionImporter() *SessionImporter {
	return &SessionImporter{}
}

// ImportTermiusJSON imports hosts from Termius export JSON
func (si *SessionImporter) ImportTermiusJSON(data []byte) (ImportResult, error) {
	var raw struct {
		Hosts []struct {
			Label    string `json:"label"`
			Address  string `json:"address"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Group    string `json:"group"`
		} `json:"hosts"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return ImportResult{}, fmt.Errorf("invalid Termius JSON: %w", err)
	}

	var hosts []models.Host
	for _, h := range raw.Hosts {
		port := h.Port
		if port == 0 {
			port = 22
		}
		hosts = append(hosts, models.Host{
			ID:          fmt.Sprintf("termius-%s-%d", h.Address, port),
			Name:        h.Label,
			Hostname:    h.Address,
			Port:        port,
			Protocol:    models.ProtocolSSH,
			Username:    h.Username,
			AuthMethod:  models.AuthAgent,
			Environment: "production",
			Folder:      h.Group,
			Health:      models.HealthUnknown,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Tags:        []string{"imported", "termius"},
		})
	}

	return ImportResult{
		ImportedCount: len(hosts),
		Hosts:         hosts,
	}, nil
}

// ImportMobaXtermSessions parses MobaXterm session export format
func (si *SessionImporter) ImportMobaXtermSessions(text string) ImportResult {
	lines := strings.Split(text, "\n")
	var hosts []models.Host

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}

		sessionName := strings.TrimSpace(parts[0])
		params := strings.Split(parts[1], "%")

		if len(params) >= 3 && strings.Contains(params[0], "#") {
			// Format: #109#1%hostname%port%username%...
			hostIP := params[1]
			port := 22
			username := "root"
			if len(params) > 2 && params[2] != "" {
				_, _ = fmt.Sscanf(params[2], "%d", &port)
			}
			if len(params) > 3 && params[3] != "" {
				username = params[3]
			}

			hosts = append(hosts, models.Host{
				ID:          fmt.Sprintf("moba-%s-%d", hostIP, port),
				Name:        sessionName,
				Hostname:    hostIP,
				Port:        port,
				Protocol:    models.ProtocolSSH,
				Username:    username,
				AuthMethod:  models.AuthAgent,
				Environment: "production",
				Folder:      "MobaXterm",
				Health:      models.HealthUnknown,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Tags:        []string{"imported", "mobaxterm"},
			})
		}
	}

	return ImportResult{
		ImportedCount: len(hosts),
		Hosts:         hosts,
	}
}
