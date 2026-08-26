package actions

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"vibeterm/internal/models"
	"vibeterm/internal/providers"
	"vibeterm/internal/services"
)

type ActionRegistry struct {
	mu sync.RWMutex
}

var (
	globalActionRegistry *ActionRegistry
	actionOnce           sync.Once
)

func GetActionRegistry() *ActionRegistry {
	actionOnce.Do(func() {
		globalActionRegistry = &ActionRegistry{}
	})
	return globalActionRegistry
}

// Execute dispatches the action to the corresponding provider or subsystem
func (ar *ActionRegistry) Execute(ctx context.Context, payload models.ActionPayload) (models.ActionResult, error) {
	provReg := providers.GetRegistry()
	prov, ok := provReg.Get(payload.ProviderID)
	if !ok {
		// Fallback check based on ActionID
		if payload.ActionID == "service.launch" {
			return ar.executeServiceLaunch(ctx, payload)
		}
		return models.ActionResult{Success: false, Error: fmt.Sprintf("provider %s not found", payload.ProviderID)}, nil
	}

	switch payload.ActionID {
	case "resource.start":
		if dp, ok := prov.(*providers.DockerProvider); ok {
			err := dp.StartContainer(ctx, payload.ResourceID)
			if err != nil {
				return models.ActionResult{Success: false, Error: err.Error()}, nil
			}
			return models.ActionResult{Success: true, Output: "Started container successfully"}, nil
		}

	case "resource.stop":
		if dp, ok := prov.(*providers.DockerProvider); ok {
			err := dp.StopContainer(ctx, payload.ResourceID)
			if err != nil {
				return models.ActionResult{Success: false, Error: err.Error()}, nil
			}
			return models.ActionResult{Success: true, Output: "Stopped container successfully"}, nil
		}

	case "resource.restart":
		if dp, ok := prov.(*providers.DockerProvider); ok {
			err := dp.RestartContainer(ctx, payload.ResourceID)
			if err != nil {
				return models.ActionResult{Success: false, Error: err.Error()}, nil
			}
			return models.ActionResult{Success: true, Output: "Restarted container successfully"}, nil
		}

	case "resource.delete":
		if dp, ok := prov.(*providers.DockerProvider); ok {
			err := dp.RemoveContainer(ctx, payload.ResourceID)
			if err != nil {
				return models.ActionResult{Success: false, Error: err.Error()}, nil
			}
			return models.ActionResult{Success: true, Output: "Deleted container successfully"}, nil
		}

	case "resource.logs":
		if dp, ok := prov.(*providers.DockerProvider); ok {
			tail := 100
			if tStr, ok := payload.Params["tail"]; ok {
				if t, err := strconv.Atoi(tStr); err == nil {
					tail = t
				}
			}
			logs, err := dp.GetLogs(ctx, payload.ResourceID, tail)
			if err != nil {
				return models.ActionResult{Success: false, Error: err.Error()}, nil
			}
			return models.ActionResult{Success: true, Output: logs}, nil
		}

	case "service.launch":
		return ar.executeServiceLaunch(ctx, payload)
	}

	return models.ActionResult{Success: false, Error: fmt.Sprintf("unsupported action %s for provider %s", payload.ActionID, payload.ProviderID)}, nil
}

func (ar *ActionRegistry) executeServiceLaunch(ctx context.Context, payload models.ActionPayload) (models.ActionResult, error) {
	svcMgr := services.GetServiceManager()
	remotePort, _ := strconv.Atoi(payload.Params["remotePort"])
	localPort, _ := strconv.Atoi(payload.Params["localPort"])

	svc := models.RemoteService{
		ID:         payload.Params["serviceId"],
		HostID:     payload.HostID,
		ResourceID: payload.ResourceID,
		Name:       payload.Params["name"],
		Type:       models.ServiceType(payload.Params["type"]),
		Strategy:   models.ServiceAccessStrategy(payload.Params["strategy"]),
		RemoteHost: payload.Params["remoteHost"],
		RemotePort: remotePort,
		LocalPort:  localPort,
		AutoTunnel: true,
		Path:       payload.Params["path"],
	}

	status, err := svcMgr.LaunchServiceWithStrategy(ctx, payload.HostID, svc)
	if err != nil {
		return models.ActionResult{Success: false, Error: err.Error()}, nil
	}

	return models.ActionResult{Success: true, Output: status.LocalURL}, nil
}
