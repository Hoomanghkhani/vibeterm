import React, { useState, useEffect, useMemo, useRef } from 'react';
import {
    Server,
    Box,
    Terminal,
    Folder,
    FolderOpen,
    ChevronRight,
    ChevronDown,
    RefreshCw,
    Search,
    Play,
    Square,
    RotateCcw,
    FileText,
    ExternalLink,
    Star,
    Edit3,
    Trash2,
    Activity,
    Plus,
    FoldHorizontal,
    Globe,
    Cpu,
    MoreVertical,
    Layers,
    X
} from 'lucide-react';
import {
    GetUnifiedInfrastructureTree,
    RefreshDiscovery,
    TriggerBackgroundRefresh,
    ExecuteResourceAction,
    SetResourceAlias,
    ToggleResourceFavorite
} from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

export interface InfrastructureNode {
    id: string;
    parentId?: string;
    nodeType: string;
    providerId?: string;
    resourceId?: string;
    hostId?: string;
    connectionId?: string;
    serviceId?: string;
    name: string;
    alias?: string;
    status: string;
    icon?: string;
    capabilities: {
        canConnect?: boolean;
        canOpenTerminal?: boolean;
        canBrowseFiles?: boolean;
        canOpenLogs?: boolean;
        canStart?: boolean;
        canStop?: boolean;
        canRestart?: boolean;
        canInspect?: boolean;
        canCreateTunnel?: boolean;
        canOpenService?: boolean;
        canDelete?: boolean;
    };
    children?: InfrastructureNode[];
    metadata?: { [key: string]: string };
}

interface InfrastructureTreeProps {
    onOpenTerminal: (node: InfrastructureNode) => void;
    onOpenFiles: (node: InfrastructureNode) => void;
    onAddHost: () => void;
    onEditHost?: (host: any) => void;
    onDeleteHost?: (e: React.MouseEvent, id: string) => void;
    activeResourceId?: string;
}

export const InfrastructureTree: React.FC<InfrastructureTreeProps> = ({
    onOpenTerminal,
    onOpenFiles,
    onAddHost,
    onEditHost,
    onDeleteHost,
    activeResourceId
}) => {
    const [treeData, setTreeData] = useState<InfrastructureNode[]>([]);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [filterProvider, setFilterProvider] = useState<'all' | 'ssh' | 'docker' | 'local'>('all');
    const [collapsedSections, setCollapsedSections] = useState<{ [id: string]: boolean }>({});
    const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

    // Inline Renaming / Alias (F2)
    const [renamingNodeId, setRenamingNodeId] = useState<string | null>(null);
    const [renameValue, setRenameValue] = useState('');

    // Context Menu
    const [contextMenu, setContextMenu] = useState<{
        visible: boolean;
        x: number;
        y: number;
        node?: InfrastructureNode;
    }>({ visible: false, x: 0, y: 0 });

    // Universal Action Logs Modal
    const [logModal, setLogModal] = useState<{
        visible: boolean;
        title: string;
        logs: string;
        node?: InfrastructureNode;
        loading: boolean;
    }>({ visible: false, title: '', logs: '', loading: false });

    // Load from cache instantly
    const loadTreeFromCache = async () => {
        try {
            const data = await GetUnifiedInfrastructureTree();
            setTreeData(data || []);
        } catch (err) {
            console.error('Failed to load infrastructure tree:', err);
        }
    };

    // Manual refresh with spinner
    const handleManualRefresh = async () => {
        setLoading(true);
        try {
            await RefreshDiscovery();
            const data = await GetUnifiedInfrastructureTree();
            setTreeData(data || []);
        } catch (err) {
            console.error('Manual discovery refresh failed:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadTreeFromCache();
        TriggerBackgroundRefresh();

        EventsOn('discovery:updated', () => {
            loadTreeFromCache();
        });

        return () => {
            EventsOff('discovery:updated');
        };
    }, []);

    // Dismiss context menu
    useEffect(() => {
        const dismiss = () => setContextMenu((prev) => (prev.visible ? { ...prev, visible: false } : prev));
        window.addEventListener('click', dismiss);
        return () => window.removeEventListener('click', dismiss);
    }, []);

    // Keyboard Shortcuts (F5, F2, Enter, Arrow keys)
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'F5') {
                e.preventDefault();
                handleManualRefresh();
            } else if (e.key === 'F2' && selectedNodeId) {
                e.preventDefault();
                const node = findNodeById(treeData, selectedNodeId);
                if (node) {
                    setRenamingNodeId(node.id);
                    setRenameValue(node.alias || node.name);
                }
            } else if (e.key === 'Enter' && selectedNodeId) {
                const node = findNodeById(treeData, selectedNodeId);
                if (node && node.capabilities.canOpenTerminal) {
                    onOpenTerminal(node);
                }
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [selectedNodeId, treeData]);

    const findNodeById = (nodes: InfrastructureNode[], id: string): InfrastructureNode | null => {
        for (const n of nodes) {
            if (n.id === id) return n;
            if (n.children && n.children.length > 0) {
                const found = findNodeById(n.children, id);
                if (found) return found;
            }
        }
        return null;
    };

    const toggleSection = (id: string) => {
        setCollapsedSections((prev) => ({ ...prev, [id]: !prev[id] }));
    };

    const collapseAll = () => {
        const collapsed: { [id: string]: boolean } = {};
        const walk = (nodes: InfrastructureNode[]) => {
            nodes.forEach((n) => {
                if (n.children && n.children.length > 0) {
                    collapsed[n.id] = true;
                    walk(n.children);
                }
            });
        };
        walk(treeData);
        setCollapsedSections(collapsed);
    };

    const handleSaveAlias = async (node: InfrastructureNode) => {
        if (!node.resourceId) {
            setRenamingNodeId(null);
            return;
        }
        await SetResourceAlias(node.resourceId, renameValue.trim());
        setRenamingNodeId(null);
        loadTreeFromCache();
    };

    const handleToggleFavorite = async (node: InfrastructureNode) => {
        if (node.resourceId) {
            await ToggleResourceFavorite(node.resourceId);
            loadTreeFromCache();
        }
    };

    const handleExecuteAction = async (actionId: string, node: InfrastructureNode, params?: { [key: string]: string }) => {
        try {
            const res = await ExecuteResourceAction({
                actionId,
                providerId: node.providerId || 'provider-docker',
                resourceId: node.resourceId || node.id,
                hostId: node.hostId || '',
                params: params || {}
            } as any);

            if (!res.success) {
                console.error(`Action ${actionId} failed:`, res.error);
            }
            loadTreeFromCache();
            return res;
        } catch (err) {
            console.error(`Failed to execute action ${actionId}:`, err);
            return { success: false, error: String(err) };
        }
    };

    const handleOpenLogs = async (node: InfrastructureNode) => {
        setLogModal({
            visible: true,
            title: `Logs: ${node.name}`,
            logs: 'Fetching logs...',
            node,
            loading: true
        });

        const res = await handleExecuteAction('resource.logs', node, { tail: '150' });
        setLogModal((prev) => ({
            ...prev,
            logs: res?.output || res?.error || 'No logs available.',
            loading: false
        }));
    };

    // XPipe-like Flattened / Categorized Organization
    const categorized = useMemo(() => {
        const localNodes: InfrastructureNode[] = [];
        const serverNodes: InfrastructureNode[] = [];
        const groupNodes: { [group: string]: InfrastructureNode[] } = {};
        const dockerNodes: InfrastructureNode[] = [];

        const q = searchQuery.toLowerCase().trim();

        const matchesQuery = (n: InfrastructureNode): boolean => {
            if (!q) return true;
            return Boolean(
                n.name.toLowerCase().includes(q) ||
                (n.alias && n.alias.toLowerCase().includes(q)) ||
                (n.metadata?.hostname && n.metadata.hostname.toLowerCase().includes(q)) ||
                (n.metadata?.username && n.metadata.username.toLowerCase().includes(q)) ||
                (n.metadata?.image && n.metadata.image.toLowerCase().includes(q)) ||
                (n.metadata?.environment && n.metadata.environment.toLowerCase().includes(q)) ||
                (n.metadata?.folder && n.metadata.folder.toLowerCase().includes(q))
            );
        };

        const processNode = (n: InfrastructureNode) => {
            if (n.nodeType === 'provider') {
                if (n.children) {
                    n.children.forEach(processNode);
                }
                return;
            }

            if (n.nodeType === 'group') {
                const groupName = n.name;
                if (!groupNodes[groupName]) {
                    groupNodes[groupName] = [];
                }
                if (n.children) {
                    n.children.forEach((child) => {
                        if (matchesQuery(child)) {
                            groupNodes[groupName].push(child);
                        }
                    });
                }
                return;
            }

            if (n.nodeType === 'resource' || n.nodeType === 'device') {
                if (!matchesQuery(n)) return;

                if (n.providerId?.includes('local') || n.id.includes('local')) {
                    localNodes.push(n);
                } else if (n.providerId?.includes('docker')) {
                    dockerNodes.push(n);
                } else if (n.metadata?.folder) {
                    const group = n.metadata.folder;
                    if (!groupNodes[group]) groupNodes[group] = [];
                    groupNodes[group].push(n);
                } else {
                    serverNodes.push(n);
                }
            }
        };

        treeData.forEach(processNode);

        // Filter out empty groups if search query exists
        const cleanGroups: { [group: string]: InfrastructureNode[] } = {};
        Object.keys(groupNodes).forEach((grp) => {
            if (groupNodes[grp].length > 0) {
                cleanGroups[grp] = groupNodes[grp];
            }
        });

        return {
            localNodes,
            serverNodes,
            groupNodes: cleanGroups,
            dockerNodes,
            totalCount: localNodes.length + serverNodes.length + dockerNodes.length + Object.values(cleanGroups).reduce((a, b) => a + b.length, 0)
        };
    }, [treeData, searchQuery]);

    // Status Dot Helper
    const renderStatusIndicator = (status: string) => {
        const s = (status || '').toLowerCase();
        if (s === 'online' || s === 'running' || s === 'ready') {
            return <div className="w-2 h-2 rounded-full bg-emerald-400 shrink-0 ring-2 ring-emerald-500/20" />;
        }
        if (s === 'degraded' || s === 'connecting') {
            return <div className="w-2 h-2 rounded-full bg-amber-400 shrink-0 animate-pulse" />;
        }
        if (s === 'failed' || s === 'unavailable') {
            return <div className="w-2 h-2 rounded-full bg-rose-500 shrink-0" />;
        }
        // Offline / stopped
        return <div className="w-2 h-2 rounded-full border border-zinc-500 bg-transparent shrink-0" />;
    };

    // XPipe Host Row Component
    const renderHostRow = (node: InfrastructureNode, depth: number = 0) => {
        const isSelected = selectedNodeId === node.id || (activeResourceId && (node.resourceId === activeResourceId || node.hostId === activeResourceId));
        const isRenaming = renamingNodeId === node.id;
        const isDocker = node.providerId?.includes('docker');
        const isLocal = node.providerId?.includes('local') || node.id.includes('local');

        // Secondary metadata string (IP / container image / user@host)
        let subText = '';
        if (isLocal) {
            subText = 'Native PTY • local';
        } else if (isDocker) {
            subText = node.metadata?.image ? `${node.metadata.image.split(':')[0]}` : (node.metadata?.ports || 'docker container');
        } else {
            const host = node.metadata?.hostname || '';
            const user = node.metadata?.username || '';
            const port = node.metadata?.port || '22';
            if (host) {
                subText = user ? `${user}@${host}${port !== '22' ? `:${port}` : ''}` : host;
            }
        }

        return (
            <div
                key={node.id}
                onClick={() => {
                    setSelectedNodeId(node.id);
                    if (node.capabilities.canOpenTerminal) {
                        onOpenTerminal(node);
                    }
                }}
                onDoubleClick={() => {
                    if (node.capabilities.canOpenTerminal) {
                        onOpenTerminal(node);
                    }
                }}
                onContextMenu={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setSelectedNodeId(node.id);
                    setContextMenu({
                        visible: true,
                        x: e.clientX,
                        y: e.clientY,
                        node
                    });
                }}
                className={`group px-2.5 py-1.5 flex items-center justify-between text-xs cursor-pointer select-none transition-colors border-l-2 ${
                    isSelected
                        ? 'bg-zinc-800/70 border-zinc-300 text-textMain font-medium'
                        : 'border-transparent text-textMuted hover:bg-bgPanel hover:text-textMain'
                }`}
                style={{ paddingLeft: `${Math.max(10, depth * 12 + 10)}px` }}
            >
                {/* Left Side: Status Dot + Host Details */}
                <div className="flex items-center gap-2.5 min-w-0 flex-1">
                    {renderStatusIndicator(node.status)}

                    <div className="flex flex-col min-w-0 flex-1 leading-tight">
                        {isRenaming ? (
                            <input
                                type="text"
                                value={renameValue}
                                onChange={(e) => setRenameValue(e.target.value)}
                                autoFocus
                                onClick={(e) => e.stopPropagation()}
                                onBlur={() => handleSaveAlias(node)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter') handleSaveAlias(node);
                                    if (e.key === 'Escape') setRenamingNodeId(null);
                                }}
                                className="bg-bgMain border border-borderActive rounded px-1.5 py-0.5 text-xs text-textMain outline-none w-36 font-mono"
                            />
                        ) : (
                            <div className="flex items-center gap-1.5 min-w-0">
                                <span className="truncate text-textMain font-medium text-xs">
                                    {node.alias || node.name}
                                </span>
                                {node.alias && (
                                    <span className="text-[9px] text-textFaint font-mono truncate">({node.name})</span>
                                )}
                            </div>
                        )}

                        {subText && (
                            <span className="text-[10px] text-zinc-500 font-mono truncate mt-0.5">
                                {subText}
                            </span>
                        )}
                    </div>
                </div>

                {/* Right Side: Quick Action Icons (Revealed on Hover) */}
                <div className="hidden group-hover:flex items-center gap-1 shrink-0 ml-2">
                    {node.capabilities.canOpenTerminal && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onOpenTerminal(node);
                            }}
                            className="p-1 text-textFaint hover:text-textMain rounded hover:bg-bgMain transition-colors"
                            title="Connect"
                        >
                            <Terminal size={12} />
                        </button>
                    )}

                    {node.capabilities.canBrowseFiles && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onOpenFiles(node);
                            }}
                            className="p-1 text-textFaint hover:text-textMain rounded hover:bg-bgMain transition-colors"
                            title="SFTP Files"
                        >
                            <Folder size={12} />
                        </button>
                    )}

                    {node.capabilities.canOpenLogs && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                handleOpenLogs(node);
                            }}
                            className="p-1 text-textFaint hover:text-sky-400 rounded hover:bg-bgMain transition-colors"
                            title="Logs"
                        >
                            <FileText size={12} />
                        </button>
                    )}

                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNodeId(node.id);
                            const rect = e.currentTarget.getBoundingClientRect();
                            setContextMenu({
                                visible: true,
                                x: rect.left,
                                y: rect.bottom + 4,
                                node
                            });
                        }}
                        className="p-1 text-textFaint hover:text-textMain rounded hover:bg-bgMain transition-colors"
                        title="Actions"
                    >
                        <MoreVertical size={12} />
                    </button>
                </div>
            </div>
        );
    };

    return (
        <div className="flex-1 flex flex-col h-full bg-bgMain border-r border-borderDark select-none overflow-hidden font-sans">
            {/* Top Minimalist Hub Bar */}
            <div className="px-3 py-2 border-b border-borderDark/70 bg-bgPanel flex flex-col gap-2 shrink-0">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1.5">
                        <Layers size={13} className="text-zinc-400" />
                        <span className="text-[11px] font-bold text-textMain tracking-wider uppercase">Infrastructure</span>
                        <span className="text-[9px] font-mono text-zinc-500 bg-bgCard px-1.5 py-0.2 rounded border border-borderDark">
                            {categorized.totalCount}
                        </span>
                    </div>

                    <div className="flex items-center gap-0.5 text-textFaint">
                        <button
                            onClick={handleManualRefresh}
                            disabled={loading}
                            className={`p-1 hover:text-textMain rounded hover:bg-bgCard transition-colors ${
                                loading ? 'animate-spin text-zinc-300' : ''
                            }`}
                            title="Refresh Infrastructure (F5)"
                        >
                            <RefreshCw size={12} />
                        </button>
                        <button
                            onClick={collapseAll}
                            className="p-1 hover:text-textMain rounded hover:bg-bgCard transition-colors"
                            title="Collapse All Groups"
                        >
                            <FoldHorizontal size={12} />
                        </button>
                        <button
                            onClick={onAddHost}
                            className="p-1 hover:text-textMain rounded hover:bg-bgCard transition-colors"
                            title="Add SSH Host"
                        >
                            <Plus size={13} />
                        </button>
                    </div>
                </div>

                {/* Compact Search Bar */}
                <div className="relative flex items-center">
                    <Search size={11} className="absolute left-2 text-zinc-500" />
                    <input
                        type="text"
                        placeholder="Search hosts, IP, tags..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full bg-bgMain border border-borderDark rounded pl-6 pr-6 py-1 text-[11px] text-textMain placeholder-zinc-500 outline-none focus:border-zinc-500 font-mono"
                    />
                    {searchQuery && (
                        <button
                            onClick={() => setSearchQuery('')}
                            className="absolute right-1.5 text-zinc-500 hover:text-textMain p-0.5"
                        >
                            <X size={11} />
                        </button>
                    )}
                </div>

                {/* Compact Filter Buttons */}
                <div className="flex items-center gap-1">
                    {(['all', 'ssh', 'docker', 'local'] as const).map((f) => (
                        <button
                            key={f}
                            onClick={() => setFilterProvider(f)}
                            className={`flex-1 py-0.5 text-[9px] font-mono font-semibold rounded text-center transition-colors ${
                                filterProvider === f
                                    ? 'bg-zinc-800 text-zinc-100 border border-zinc-600 shadow-sm'
                                    : 'text-zinc-500 hover:text-zinc-300 bg-bgMain border border-borderDark/40'
                            }`}
                        >
                            {f.toUpperCase()}
                        </button>
                    ))}
                </div>
            </div>

            {/* List / Tree Container */}
            <div className="flex-1 overflow-y-auto divide-y divide-borderDark/20">
                {/* 1. LOCAL SECTION */}
                {(filterProvider === 'all' || filterProvider === 'local') && categorized.localNodes.length > 0 && (
                    <div className="py-1">
                        <div
                            onClick={() => toggleSection('local')}
                            className="px-3 py-1 flex items-center justify-between text-[10px] font-semibold text-zinc-500 uppercase tracking-wider cursor-pointer hover:text-zinc-300"
                        >
                            <div className="flex items-center gap-1">
                                {collapsedSections['local'] ? <ChevronRight size={10} /> : <ChevronDown size={10} />}
                                <span>Local</span>
                            </div>
                        </div>
                        {!collapsedSections['local'] && categorized.localNodes.map((n) => renderHostRow(n, 0))}
                    </div>
                )}

                {/* 2. SERVERS (Ungrouped SSH Hosts) */}
                {(filterProvider === 'all' || filterProvider === 'ssh') && categorized.serverNodes.length > 0 && (
                    <div className="py-1">
                        <div
                            onClick={() => toggleSection('servers')}
                            className="px-3 py-1 flex items-center justify-between text-[10px] font-semibold text-zinc-500 uppercase tracking-wider cursor-pointer hover:text-zinc-300"
                        >
                            <div className="flex items-center gap-1">
                                {collapsedSections['servers'] ? <ChevronRight size={10} /> : <ChevronDown size={10} />}
                                <span>Servers</span>
                            </div>
                            <span className="font-mono text-[9px]">({categorized.serverNodes.length})</span>
                        </div>
                        {!collapsedSections['servers'] && categorized.serverNodes.map((n) => renderHostRow(n, 0))}
                    </div>
                )}

                {/* 3. GROUPS (Folders) */}
                {(filterProvider === 'all' || filterProvider === 'ssh') && Object.keys(categorized.groupNodes).length > 0 && (
                    <div className="py-1">
                        <div className="px-3 py-1 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider">
                            <span>Groups</span>
                        </div>

                        {Object.entries(categorized.groupNodes).map(([groupName, hosts]) => {
                            const isGroupCollapsed = !!collapsedSections[`group-${groupName}`];
                            return (
                                <div key={groupName} className="flex flex-col">
                                    <div
                                        onClick={() => toggleSection(`group-${groupName}`)}
                                        className="px-3 py-1 flex items-center justify-between text-xs text-zinc-300 cursor-pointer hover:bg-bgPanel transition-colors font-medium"
                                    >
                                        <div className="flex items-center gap-1.5 min-w-0">
                                            <span className="text-zinc-500">
                                                {isGroupCollapsed ? <ChevronRight size={11} /> : <ChevronDown size={11} />}
                                            </span>
                                            {isGroupCollapsed ? (
                                                <Folder size={13} className="text-zinc-400 shrink-0" />
                                            ) : (
                                                <FolderOpen size={13} className="text-zinc-300 shrink-0" />
                                            )}
                                            <span className="truncate text-xs">{groupName}</span>
                                        </div>
                                        <span className="text-[10px] font-mono text-zinc-500">({hosts.length})</span>
                                    </div>

                                    {!isGroupCollapsed && (
                                        <div className="flex flex-col border-l border-borderDark/30 ml-4">
                                            {hosts.map((h) => renderHostRow(h, 1))}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                )}

                {/* 4. DOCKER CONTAINERS */}
                {(filterProvider === 'all' || filterProvider === 'docker') && categorized.dockerNodes.length > 0 && (
                    <div className="py-1">
                        <div
                            onClick={() => toggleSection('docker')}
                            className="px-3 py-1 flex items-center justify-between text-[10px] font-semibold text-zinc-500 uppercase tracking-wider cursor-pointer hover:text-zinc-300"
                        >
                            <div className="flex items-center gap-1">
                                {collapsedSections['docker'] ? <ChevronRight size={10} /> : <ChevronDown size={10} />}
                                <span>Docker Containers</span>
                            </div>
                            <span className="font-mono text-[9px]">({categorized.dockerNodes.length})</span>
                        </div>
                        {!collapsedSections['docker'] && categorized.dockerNodes.map((n) => renderHostRow(n, 0))}
                    </div>
                )}

                {/* Empty State */}
                {categorized.totalCount === 0 && (
                    <div className="p-6 text-center text-xs text-zinc-500 flex flex-col items-center justify-center gap-2">
                        <Server size={24} className="text-zinc-600" />
                        <span>No endpoints found</span>
                        <button
                            onClick={onAddHost}
                            className="text-xs text-zinc-300 hover:text-white underline mt-1"
                        >
                            + Add SSH Host
                        </button>
                    </div>
                )}
            </div>

            {/* XPipe Clean Context Menu */}
            {contextMenu.visible && contextMenu.node && (
                <div
                    className="fixed z-50 w-48 bg-bgCard border border-borderDark rounded-lg shadow-2xl py-1 text-xs text-textMain overflow-hidden select-none"
                    style={{ top: `${contextMenu.y}px`, left: `${contextMenu.x}px` }}
                    onClick={(e) => e.stopPropagation()}
                >
                    {contextMenu.node.capabilities.canOpenTerminal && (
                        <button
                            onClick={() => {
                                onOpenTerminal(contextMenu.node!);
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                        >
                            <Terminal size={12} className="text-zinc-400" />
                            <span>Connect (New Tab)</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canBrowseFiles && (
                        <button
                            onClick={() => {
                                onOpenFiles(contextMenu.node!);
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                        >
                            <Folder size={12} className="text-zinc-400" />
                            <span>Open SFTP Files</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canOpenLogs && (
                        <button
                            onClick={() => {
                                handleOpenLogs(contextMenu.node!);
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-sky-400 transition-colors"
                        >
                            <FileText size={12} className="text-zinc-400" />
                            <span>View Container Logs</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canStart && (
                        <button
                            onClick={async () => {
                                await handleExecuteAction('resource.start', contextMenu.node!);
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-emerald-400 transition-colors"
                        >
                            <Play size={12} className="text-zinc-400" />
                            <span>Start Container</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canStop && (
                        <button
                            onClick={async () => {
                                await handleExecuteAction('resource.stop', contextMenu.node!);
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-amber-400 transition-colors"
                        >
                            <Square size={12} className="text-zinc-400" />
                            <span>Stop Container</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canRestart && (
                        <button
                            onClick={async () => {
                                await handleExecuteAction('resource.restart', contextMenu.node!);
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-sky-400 transition-colors"
                        >
                            <RotateCcw size={12} className="text-zinc-400" />
                            <span>Restart Container</span>
                        </button>
                    )}

                    <div className="h-[1px] bg-borderDark my-1"></div>

                    {onEditHost && contextMenu.node.hostId && (
                        <button
                            onClick={() => {
                                onEditHost({ id: contextMenu.node!.hostId, name: contextMenu.node!.name });
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                        >
                            <Edit3 size={12} className="text-zinc-400" />
                            <span>Edit Host...</span>
                        </button>
                    )}

                    <button
                        onClick={() => {
                            setRenamingNodeId(contextMenu.node!.id);
                            setRenameValue(contextMenu.node!.alias || contextMenu.node!.name);
                            setContextMenu({ ...contextMenu, visible: false });
                        }}
                        className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                    >
                        <div className="flex items-center gap-2">
                            <Edit3 size={12} className="text-zinc-400" />
                            <span>Set Custom Alias</span>
                        </div>
                        <span className="text-[9px] text-zinc-500 font-mono">F2</span>
                    </button>

                    <button
                        onClick={() => {
                            handleToggleFavorite(contextMenu.node!);
                            setContextMenu({ ...contextMenu, visible: false });
                        }}
                        className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-amber-400 transition-colors"
                    >
                        <Star size={12} className="text-zinc-400" />
                        <span>Toggle Favorite</span>
                    </button>

                    {onDeleteHost && contextMenu.node.hostId && (
                        <>
                            <div className="h-[1px] bg-borderDark my-1"></div>
                            <button
                                onClick={(e) => {
                                    onDeleteHost(e, contextMenu.node!.hostId!);
                                    setContextMenu({ ...contextMenu, visible: false });
                                }}
                                className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-rose-400 transition-colors"
                            >
                                <Trash2 size={12} className="text-zinc-400" />
                                <span>Delete Host</span>
                            </button>
                        </>
                    )}
                </div>
            )}

            {/* Docker / Container Logs Modal */}
            {logModal.visible && (
                <div className="fixed inset-0 z-50 bg-black/75 flex items-center justify-center p-4">
                    <div className="bg-bgCard border border-borderDark rounded-lg w-[750px] max-w-full h-[500px] flex flex-col shadow-2xl overflow-hidden">
                        <div className="px-4 py-2 bg-bgPanel border-b border-borderDark flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <FileText size={13} className="text-sky-400" />
                                <span className="text-xs font-mono font-semibold text-textMain">{logModal.title}</span>
                            </div>
                            <div className="flex items-center gap-2">
                                <button
                                    onClick={() => logModal.node && handleOpenLogs(logModal.node)}
                                    className="p-1 hover:text-textMain text-zinc-400 rounded transition-colors"
                                    title="Refresh Logs"
                                >
                                    <RefreshCw size={12} />
                                </button>
                                <button
                                    onClick={() => setLogModal({ ...logModal, visible: false })}
                                    className="text-zinc-400 hover:text-white text-xs px-2 py-0.5 rounded bg-bgMain hover:bg-bgHover border border-borderDark"
                                >
                                    Close
                                </button>
                            </div>
                        </div>

                        <div className="flex-1 p-3 bg-black overflow-y-auto font-mono text-xs text-zinc-300 leading-relaxed whitespace-pre-wrap select-text">
                            {logModal.logs}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
