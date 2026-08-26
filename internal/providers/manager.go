package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"vibeterm/internal/models"
)

type DiscoveryCache struct {
	LastUpdated time.Time                        `json:"lastUpdated"`
	Results     []models.ProviderDiscoveryResult `json:"results"`
	Aliases     map[string]string                `json:"aliases"`
	Favorites   map[string]bool                  `json:"favorites"`
}

type DiscoveryManager struct {
	mu        sync.RWMutex
	cachePath string
	cache     DiscoveryCache
}

var (
	globalDiscoveryMgr *DiscoveryManager
	discoveryMgrOnce   sync.Once
)

func GetDiscoveryManager() *DiscoveryManager {
	discoveryMgrOnce.Do(func() {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, ".vibeterm")
		_ = os.MkdirAll(dir, 0700)
		cachePath := filepath.Join(dir, "discovery_cache.json")

		mgr := &DiscoveryManager{
			cachePath: cachePath,
			cache: DiscoveryCache{
				Aliases:   make(map[string]string),
				Favorites: make(map[string]bool),
			},
		}
		mgr.loadCache()
		globalDiscoveryMgr = mgr
	})
	return globalDiscoveryMgr
}

func (dm *DiscoveryManager) loadCache() {
	data, err := os.ReadFile(dm.cachePath)
	if err == nil {
		_ = json.Unmarshal(data, &dm.cache)
	}
	if dm.cache.Aliases == nil {
		dm.cache.Aliases = make(map[string]string)
	}
	if dm.cache.Favorites == nil {
		dm.cache.Favorites = make(map[string]bool)
	}
}

func (dm *DiscoveryManager) saveCacheLocked() {
	data, err := json.MarshalIndent(dm.cache, "", "  ")
	if err == nil {
		_ = os.WriteFile(dm.cachePath, data, 0600)
	}
}

// RefreshAll executes discovery across all providers in parallel with failure isolation
func (dm *DiscoveryManager) RefreshAll(ctx context.Context) []models.ProviderDiscoveryResult {
	reg := GetRegistry()
	allProviders := reg.GetAll()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []models.ProviderDiscoveryResult

	for _, p := range allProviders {
		wg.Add(1)
		go func(prov InfrastructureProvider) {
			defer wg.Done()

			res, err := prov.Discover(ctx)
			status := models.StatusReady
			errMsg := ""

			if err != nil {
				status = models.StatusDegraded
				errMsg = err.Error()
			} else if len(res) == 0 {
				status = models.StatusReady
			}

			result := models.ProviderDiscoveryResult{
				ProviderID: prov.ID(),
				Name:       prov.Name(),
				Status:     status,
				Resources:  res,
				Error:      errMsg,
				LastSync:   time.Now(),
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	dm.mu.Lock()
	dm.cache.Results = results
	dm.cache.LastUpdated = time.Now()
	dm.saveCacheLocked()
	dm.mu.Unlock()

	return results
}

// GetUnifiedTree builds the normalized hierarchy tree for the UI
func (dm *DiscoveryManager) GetUnifiedTree(ctx context.Context) []models.InfrastructureNode {
	dm.mu.RLock()
	cachedResults := dm.cache.Results
	aliases := dm.cache.Aliases
	dm.mu.RUnlock()

	// If cache is empty, run refresh
	if len(cachedResults) == 0 {
		cachedResults = dm.RefreshAll(ctx)
	}

	var rootNodes []models.InfrastructureNode

	for _, provResult := range cachedResults {
		provNode := models.InfrastructureNode{
			ID:         fmt.Sprintf("provider-%s", provResult.ProviderID),
			NodeType:   models.NodeProvider,
			ProviderID: provResult.ProviderID,
			Name:       provResult.Name,
			Status:     string(provResult.Status),
			Capabilities: models.ResourceCapabilities{
				CanInspect: true,
			},
			Children: []models.InfrastructureNode{},
		}

		// Group resources by Folder/Grouping
		folderMap := make(map[string]*models.InfrastructureNode)

		for _, res := range provResult.Resources {
			alias := aliases[res.ID]
			displayName := res.Name
			if alias != "" {
				displayName = alias
			}

			// ONLY set HostID when resource is actually a Host or has a HostRef
			hostID := res.HostRef
			if res.Type == models.ResourceServer {
				hostID = res.ID
			}

			resNode := models.InfrastructureNode{
				ID:           fmt.Sprintf("res-%s", res.ID),
				ParentID:     provNode.ID,
				NodeType:     models.NodeResource,
				ProviderID:   res.ProviderID,
				ResourceID:   res.ID,
				HostID:       hostID,
				Name:         displayName,
				Alias:        alias,
				Status:       res.Status,
				Capabilities: res.Capabilities,
				Metadata:     res.Metadata,
				Children:     []models.InfrastructureNode{},
			}

			// Add Connections as child nodes
			for _, conn := range res.Connections {
				resNode.Children = append(resNode.Children, models.InfrastructureNode{
					ID:           fmt.Sprintf("conn-%s", conn.ID),
					ParentID:     resNode.ID,
					NodeType:     models.NodeConnection,
					ProviderID:   res.ProviderID,
					ResourceID:   res.ID,
					HostID:       conn.HostID,
					ConnectionID: conn.ID,
					Name:         conn.Name,
					Status:       res.Status,
					Capabilities: models.ResourceCapabilities{
						CanConnect:      true,
						CanOpenTerminal: true,
					},
				})
			}

			// Add Services as child nodes
			for _, svc := range res.Services {
				resNode.Children = append(resNode.Children, models.InfrastructureNode{
					ID:         fmt.Sprintf("svc-%s", svc.ID),
					ParentID:   resNode.ID,
					NodeType:   models.NodeService,
					ProviderID: res.ProviderID,
					ResourceID: res.ID,
					HostID:     svc.HostID,
					ServiceID:  svc.ID,
					Name:       fmt.Sprintf("%s (:%d)", svc.Name, svc.RemotePort),
					Status:     "available",
					Capabilities: models.ResourceCapabilities{
						CanOpenService:  true,
						CanCreateTunnel: true,
					},
					Metadata: map[string]string{
						"remoteHost": svc.RemoteHost,
						"remotePort": fmt.Sprintf("%d", svc.RemotePort),
						"localPort":  fmt.Sprintf("%d", svc.LocalPort),
						"path":       svc.Path,
						"strategy":   string(svc.Strategy),
					},
				})
			}

			// If resource has a folder grouping
			folder := res.Folder
			if folder == "" {
				provNode.Children = append(provNode.Children, resNode)
			} else {
				if _, ok := folderMap[folder]; !ok {
					folderNode := &models.InfrastructureNode{
						ID:       fmt.Sprintf("group-%s-%s", provResult.ProviderID, folder),
						ParentID: provNode.ID,
						NodeType: models.NodeGroup,
						Name:     folder,
						Status:   "online",
						Children: []models.InfrastructureNode{},
					}
					folderMap[folder] = folderNode
					provNode.Children = append(provNode.Children, *folderNode)
				}
				// Append to folder node
				for i := range provNode.Children {
					if provNode.Children[i].Name == folder && provNode.Children[i].NodeType == models.NodeGroup {
						provNode.Children[i].Children = append(provNode.Children[i].Children, resNode)
						break
					}
				}
			}
		}

		rootNodes = append(rootNodes, provNode)
	}

	return rootNodes
}

func (dm *DiscoveryManager) SetResourceAlias(resourceID, alias string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if alias == "" {
		delete(dm.cache.Aliases, resourceID)
	} else {
		dm.cache.Aliases[resourceID] = alias
	}
	dm.saveCacheLocked()
}

func (dm *DiscoveryManager) ToggleFavorite(resourceID string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.cache.Favorites[resourceID] = !dm.cache.Favorites[resourceID]
	dm.saveCacheLocked()
	return dm.cache.Favorites[resourceID]
}
