package providers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"vibeterm/internal/discovery"
	"vibeterm/internal/models"
	"vibeterm/internal/transport"
)

type DockerProvider struct {
	disc *discovery.DockerDiscovery
}

func NewDockerProvider() *DockerProvider {
	return &DockerProvider{
		disc: discovery.NewDockerDiscovery(),
	}
}

func (dp *DockerProvider) ID() string {
	return "provider-docker"
}

func (dp *DockerProvider) Name() string {
	return "Docker Engine & Containers"
}

func (dp *DockerProvider) Type() models.ProviderType {
	return models.ProviderDocker
}

func (dp *DockerProvider) Discover(ctx context.Context) ([]models.Resource, error) {
	containers, err := dp.disc.DiscoverContainers(ctx)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, c := range containers {
		status := "stopped"
		if strings.HasPrefix(strings.ToLower(c.Status), "up") {
			status = "running"
		}

		name := c.Names
		if name == "" {
			name = c.Image
		}

		resources = append(resources, models.Resource{
			ID:         c.ID,
			ProviderID: dp.ID(),
			Type:       models.ResourceContainer,
			Name:       name,
			Status:     status,
			Metadata: map[string]string{
				"image":   c.Image,
				"command": c.Command,
				"ports":   c.Ports,
				"state":   c.State,
				"status":  c.Status,
			},
		})
	}

	return resources, nil
}

func (dp *DockerProvider) GetResource(ctx context.Context, id string) (*models.Resource, error) {
	resources, err := dp.Discover(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range resources {
		if r.ID == id || strings.HasPrefix(r.ID, id) {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("container %s not found", id)
}

func (dp *DockerProvider) CreateTransport(ctx context.Context, res models.Resource, sessionID string, cols, rows int) (transport.TerminalTransport, error) {
	return transport.NewDockerExecTransport(sessionID, res.ID), nil
}

// Container Lifecycle Actions
func (dp *DockerProvider) StartContainer(ctx context.Context, containerID string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctxTimeout, "docker", "start", containerID).Run()
}

func (dp *DockerProvider) StopContainer(ctx context.Context, containerID string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctxTimeout, "docker", "stop", containerID).Run()
}

func (dp *DockerProvider) RestartContainer(ctx context.Context, containerID string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctxTimeout, "docker", "restart", containerID).Run()
}

func (dp *DockerProvider) RemoveContainer(ctx context.Context, containerID string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctxTimeout, "docker", "rm", "-f", containerID).Run()
}

func (dp *DockerProvider) GetLogs(ctx context.Context, containerID string, tail int) (string, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tailStr := fmt.Sprintf("%d", tail)
	if tail <= 0 {
		tailStr = "100"
	}
	out, err := exec.CommandContext(ctxTimeout, "docker", "logs", "--tail", tailStr, containerID).CombinedOutput()
	return string(out), err
}
