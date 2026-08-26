package discovery

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

type DockerContainerInfo struct {
	ID      string `json:"id"`
	Image   string `json:"image"`
	Command string `json:"command"`
	Created string `json:"createdAt"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Names   string `json:"names"`
	Ports   string `json:"ports"`
}

type DockerDiscovery struct{}

func NewDockerDiscovery() *DockerDiscovery {
	return &DockerDiscovery{}
}

func (d *DockerDiscovery) DiscoverContainers(ctx context.Context) ([]DockerContainerInfo, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxTimeout, "docker", "ps", "-a", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return []DockerContainerInfo{}, nil
	}

	var list []DockerContainerInfo
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		var c DockerContainerInfo
		if err := json.Unmarshal([]byte(l), &c); err == nil {
			list = append(list, c)
		}
	}
	return list, nil
}
