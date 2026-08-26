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
    SlidersHorizontal,
    Plus,
    FoldHorizontal,
    Globe,
    Cpu
} from 'lucide-react';
import {
    GetUnifiedInfrastructureTree,
    RefreshDiscovery,
    SetResourceAlias,
    ToggleResourceFavorite,
    DockerStartContainer,
    DockerStopContainer,
    DockerRestartContainer,
    DockerRemoveContainer,
    DockerGetLogs,
    LaunchRemoteService
} from '../wailsjs/go/main/App';

interface InfrastructureNode {
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
    activeResourceId?: string;
}

export const InfrastructureTree: React.FC<InfrastructureTreeProps> = ({
    onOpenTerminal,
    onOpenFiles,
    onAddHost,
    activeResourceId
}) => {
    const [treeData, setTreeData] = useState<InfrastructureNode[]>([]);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [filterProvider, setFilterProvider] = useState<'all' | 'ssh' | 'docker' | 'local'>('all');
    const [collapsedNodes, setCollapsedNodes] = useState<{ [id: string]: boolean }>({});
    const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

    // Renaming / Alias
    const [renamingNodeId, setRenamingNodeId] = useState<string | null>(null);
    const [renameValue, setRenameValue] = useState('');

    // Context Menu
    const [contextMenu, setContextMenu] = useState<{
        visible: boolean;
        x: number;
        y: number;
        node?: InfrastructureNode;
    }>({ visible: false, x: 0, y: 0 });

    // Docker Logs Modal
    const [logModal, setLogModal] = useState<{
        visible: boolean;
        title: string;
        logs: string;
        containerId?: string;
        loading: boolean;
    }>({ visible: false, title: '', logs: '', loading: false });

    // Service launch feedback
    const [serviceLaunching, setServiceLaunching] = useState<string | null>(null);

    const fetchTree = async (isRefresh = false) => {
        setLoading(true);
        try {
            if (isRefresh) {
                await RefreshDiscovery();
            }
            const data = await GetUnifiedInfrastructureTree();
            setTreeData(data || []);
        } catch (err) {
            console.error('Failed to load infrastructure tree:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTree(false);
    }, []);

    // Dismiss context menu on outside click
    useEffect(() => {
        const dismiss = () => setContextMenu((prev) => (prev.visible ? { ...prev, visible: false } : prev));
        window.addEventListener('click', dismiss);
        return () => window.removeEventListener('click', dismiss);
    }, []);

    // Keyboard navigation and F5 refresh
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'F5') {
                e.preventDefault();
                fetchTree(true);
            }
            if (e.key === 'F2' && selectedNodeId) {
                e.preventDefault();
                const node = findNodeById(treeData, selectedNodeId);
                if (node) {
                    setRenamingNodeId(node.id);
                    setRenameValue(node.alias || node.name);
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

    const toggleCollapse = (id: string) => {
        setCollapsedNodes((prev) => ({ ...prev, [id]: !prev[id] }));
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
        setCollapsedNodes(collapsed);
    };

    const handleSaveAlias = async (node: InfrastructureNode) => {
        if (!node.resourceId) {
            setRenamingNodeId(null);
            return;
        }
        await SetResourceAlias(node.resourceId, renameValue.trim());
        setRenamingNodeId(null);
        fetchTree(false);
    };

    const handleToggleFavorite = async (node: InfrastructureNode) => {
        if (node.resourceId) {
            await ToggleResourceFavorite(node.resourceId);
            fetchTree(false);
        }
    };

    const handleOpenLogs = async (node: InfrastructureNode) => {
        if (!node.resourceId) return;
        setLogModal({
            visible: true,
            title: `Logs: ${node.name}`,
            logs: 'Fetching logs...',
            containerId: node.resourceId,
            loading: true
        });
        try {
            const logs = await DockerGetLogs(node.resourceId, 150);
            setLogModal((prev) => ({ ...prev, logs: logs || 'No logs available.', loading: false }));
        } catch (err: any) {
            setLogModal((prev) => ({ ...prev, logs: `Error reading logs: ${err}`, loading: false }));
        }
    };

    const handleLaunchService = async (node: InfrastructureNode) => {
        if (!node.hostId || !node.serviceId) return;
        setServiceLaunching(node.id);
        try {
            await LaunchRemoteService(node.hostId, {
                id: node.serviceId,
                hostId: node.hostId,
                name: node.name,
                type: 'http',
                remoteHost: node.metadata?.remoteHost || '127.0.0.1',
                remotePort: parseInt(node.metadata?.remotePort || '80', 10),
                autoTunnel: true,
                path: node.metadata?.path || ''
            } as any);
        } catch (err) {
            console.error('Failed to launch service:', err);
        } finally {
            setServiceLaunching(null);
        }
    };

    // Filter and search tree
    const filteredTree = useMemo(() => {
        if (!searchQuery && filterProvider === 'all') return treeData;

        const matchesQuery = (n: InfrastructureNode): boolean => {
            const q = searchQuery.toLowerCase();
            return Boolean(
                n.name.toLowerCase().includes(q) ||
                (n.alias && n.alias.toLowerCase().includes(q)) ||
                (n.metadata?.hostname && n.metadata.hostname.toLowerCase().includes(q)) ||
                (n.metadata?.image && n.metadata.image.toLowerCase().includes(q))
            );
        };

        const filterBranch = (n: InfrastructureNode): InfrastructureNode | null => {
            if (filterProvider !== 'all') {
                if (n.nodeType === 'provider') {
                    if (filterProvider === 'ssh' && !n.id.includes('ssh')) return null;
                    if (filterProvider === 'docker' && !n.id.includes('docker')) return null;
                    if (filterProvider === 'local' && !n.id.includes('local')) return null;
                }
            }

            const children: InfrastructureNode[] = [];
            if (n.children) {
                for (const c of n.children) {
                    const filtered = filterBranch(c);
                    if (filtered) children.push(filtered);
                }
            }

            if (matchesQuery(n) || children.length > 0) {
                return { ...n, children };
            }
            return null;
        };

        const result: InfrastructureNode[] = [];
        for (const root of treeData) {
            const f = filterBranch(root);
            if (f) result.push(f);
        }
        return result;
    }, [treeData, searchQuery, filterProvider]);

    const renderNodeIcon = (node: InfrastructureNode, isCollapsed: boolean) => {
        switch (node.nodeType) {
            case 'provider':
                if (node.providerId?.includes('docker')) return <Box size={14} className="text-sky-400 shrink-0" />;
                if (node.providerId?.includes('ssh')) return <Server size={14} className="text-zinc-300 shrink-0" />;
                return <Cpu size={14} className="text-emerald-400 shrink-0" />;
            case 'group':
                return isCollapsed ? (
                    <Folder size={14} className="text-zinc-400 fill-zinc-400/20 shrink-0" />
                ) : (
                    <FolderOpen size={14} className="text-zinc-300 fill-zinc-300/30 shrink-0" />
                );
            case 'resource':
                if (node.providerId?.includes('docker')) return <Box size={13} className="text-sky-400/80 shrink-0" />;
                return <Server size={13} className="text-zinc-400 shrink-0" />;
            case 'connection':
                return <Terminal size={12} className="text-textMuted shrink-0" />;
            case 'service':
                return <Globe size={12} className="text-amber-400 shrink-0" />;
            default:
                return <Activity size={13} className="text-textFaint shrink-0" />;
        }
    };

    const renderStatusBadge = (status: string) => {
        const s = (status || '').toLowerCase();
        if (s === 'online' || s === 'running' || s === 'ready') {
            return <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0 shadow-sm shadow-emerald-500/50" />;
        }
        if (s === 'degraded') {
            return <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0" />;
        }
        if (s === 'failed' || s === 'unavailable') {
            return <div className="w-1.5 h-1.5 rounded-full bg-rose-400 shrink-0" />;
        }
        return <div className="w-1.5 h-1.5 rounded-full bg-zinc-600 shrink-0" />;
    };

    const renderNode = (node: InfrastructureNode, depth: number = 0) => {
        const hasChildren = node.children && node.children.length > 0;
        const isCollapsed = !!collapsedNodes[node.id];
        const isSelected = selectedNodeId === node.id;
        const isRenaming = renamingNodeId === node.id;

        return (
            <div key={node.id} className="flex flex-col select-none">
                {/* Node Row */}
                <div
                    onClick={() => {
                        setSelectedNodeId(node.id);
                        if (hasChildren) {
                            toggleCollapse(node.id);
                        } else if (node.capabilities.canOpenTerminal) {
                            onOpenTerminal(node);
                        } else if (node.capabilities.canOpenService) {
                            handleLaunchService(node);
                        }
                    }}
                    onDoubleClick={() => {
                        if (node.capabilities.canOpenTerminal) {
                            onOpenTerminal(node);
                        } else if (node.capabilities.canOpenService) {
                            handleLaunchService(node);
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
                    className={`group px-2 py-1 flex items-center justify-between text-xs cursor-pointer transition-colors ${
                        isSelected
                            ? 'bg-bgPanel text-textMain font-medium border-l-2 border-zinc-400'
                            : 'text-textMuted hover:bg-bgHover hover:text-textMain'
                    }`}
                    style={{ paddingLeft: `${Math.max(8, depth * 12 + 8)}px` }}
                >
                    <div className="flex items-center gap-1.5 min-w-0 flex-1">
                        {hasChildren ? (
                            <span className="text-textFaint group-hover:text-textMain transition-transform">
                                {isCollapsed ? <ChevronRight size={12} /> : <ChevronDown size={12} />}
                            </span>
                        ) : (
                            <span className="w-3" />
                        )}

                        {renderNodeIcon(node, isCollapsed)}
                        {renderStatusBadge(node.status)}

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
                                className="bg-bgMain border border-borderActive rounded px-1 py-0.5 text-xs text-textMain outline-none w-32 font-mono"
                            />
                        ) : (
                            <div className="flex items-center gap-1.5 min-w-0 truncate">
                                <span className="truncate text-textMain text-xs">
                                    {node.alias || node.name}
                                </span>
                                {node.alias && (
                                    <span className="text-[10px] text-textFaint font-mono">({node.name})</span>
                                )}
                            </div>
                        )}
                    </div>

                    {/* Quick Action Icons on Hover */}
                    <div className="hidden group-hover:flex items-center gap-1 shrink-0 ml-1">
                        {node.capabilities.canOpenTerminal && (
                            <button
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onOpenTerminal(node);
                                }}
                                className="p-0.5 text-textFaint hover:text-textMain rounded transition-colors"
                                title="Open Terminal"
                            >
                                <Terminal size={12} />
                            </button>
                        )}
                        {node.capabilities.canOpenLogs && (
                            <button
                                onClick={(e) => {
                                    e.stopPropagation();
                                    handleOpenLogs(node);
                                }}
                                className="p-0.5 text-textFaint hover:text-sky-400 rounded transition-colors"
                                title="View Logs"
                            >
                                <FileText size={12} />
                            </button>
                        )}
                        {node.capabilities.canOpenService && (
                            <button
                                onClick={(e) => {
                                    e.stopPropagation();
                                    handleLaunchService(node);
                                }}
                                className="p-0.5 text-textFaint hover:text-amber-400 rounded transition-colors"
                                title="Open Remote Service"
                            >
                                <ExternalLink size={12} />
                            </button>
                        )}
                    </div>
                </div>

                {/* Recursive Children with Indentation Guide */}
                {!isCollapsed && hasChildren && (
                    <div className="flex flex-col border-l border-borderDark/25 ml-3">
                        {node.children!.map((child) => renderNode(child, depth + 1))}
                    </div>
                )}
            </div>
        );
    };

    return (
        <div className="flex-1 flex flex-col h-full bg-bgMain border-r border-borderDark select-none overflow-hidden">
            {/* Top Hub Bar */}
            <div className="p-2 border-b border-borderDark bg-bgPanel flex flex-col gap-2">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1.5">
                        <Activity size={14} className="text-zinc-300" />
                        <span className="text-xs font-semibold text-textMain tracking-wide">INFRASTRUCTURE</span>
                    </div>

                    <div className="flex items-center gap-1 text-textFaint">
                        <button
                            onClick={() => fetchTree(true)}
                            disabled={loading}
                            className={`p-1 hover:text-textMain rounded hover:bg-bgCard transition-colors ${
                                loading ? 'animate-spin text-zinc-300' : ''
                            }`}
                            title="Refresh Infrastructure (F5)"
                        >
                            <RefreshCw size={13} />
                        </button>
                        <button
                            onClick={collapseAll}
                            className="p-1 hover:text-textMain rounded hover:bg-bgCard transition-colors"
                            title="Collapse All"
                        >
                            <FoldHorizontal size={13} />
                        </button>
                        <button
                            onClick={onAddHost}
                            className="p-1 hover:text-textMain rounded hover:bg-bgCard transition-colors"
                            title="Add SSH Host"
                        >
                            <Plus size={14} />
                        </button>
                    </div>
                </div>

                {/* Search Bar */}
                <div className="relative flex items-center">
                    <Search size={12} className="absolute left-2 text-textFaint" />
                    <input
                        type="text"
                        placeholder="Search resources, tags, IP..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full bg-bgMain border border-borderDark rounded pl-7 pr-2 py-1 text-xs text-textMain placeholder-textFaint outline-none focus:border-zinc-500 font-mono"
                    />
                </div>

                {/* Filter Chips */}
                <div className="flex items-center gap-1">
                    {(['all', 'ssh', 'docker', 'local'] as const).map((f) => (
                        <button
                            key={f}
                            onClick={() => setFilterProvider(f)}
                            className={`px-2 py-0.5 rounded text-[10px] font-mono transition-colors ${
                                filterProvider === f
                                    ? 'bg-zinc-800 text-textMain border border-zinc-600 font-medium'
                                    : 'text-textFaint hover:text-textMuted bg-bgMain border border-borderDark/40'
                            }`}
                        >
                            {f.toUpperCase()}
                        </button>
                    ))}
                </div>
            </div>

            {/* Tree View Canvas */}
            <div className="flex-1 overflow-y-auto py-1">
                {filteredTree.length === 0 ? (
                    <div className="p-4 text-center text-xs text-textFaint">
                        {loading ? 'Discovering infrastructure...' : 'No resources discovered.'}
                    </div>
                ) : (
                    filteredTree.map((root) => renderNode(root, 0))
                )}
            </div>

            {/* Universal Context Menu */}
            {contextMenu.visible && contextMenu.node && (
                <div
                    className="fixed z-50 w-52 bg-bgCard border border-borderDark rounded-lg shadow-2xl py-1 text-xs text-textMain overflow-hidden select-none"
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
                            <Terminal size={13} className="text-textFaint" />
                            <span>Open Terminal</span>
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
                            <Folder size={13} className="text-textFaint" />
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
                            <FileText size={13} className="text-textFaint" />
                            <span>View Container Logs</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canStart && (
                        <button
                            onClick={async () => {
                                if (contextMenu.node?.resourceId) {
                                    await DockerStartContainer(contextMenu.node.resourceId);
                                    fetchTree(false);
                                }
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-emerald-400 transition-colors"
                        >
                            <Play size={13} className="text-textFaint" />
                            <span>Start Container</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canStop && (
                        <button
                            onClick={async () => {
                                if (contextMenu.node?.resourceId) {
                                    await DockerStopContainer(contextMenu.node.resourceId);
                                    fetchTree(false);
                                }
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-amber-400 transition-colors"
                        >
                            <Square size={13} className="text-textFaint" />
                            <span>Stop Container</span>
                        </button>
                    )}

                    {contextMenu.node.capabilities.canRestart && (
                        <button
                            onClick={async () => {
                                if (contextMenu.node?.resourceId) {
                                    await DockerRestartContainer(contextMenu.node.resourceId);
                                    fetchTree(false);
                                }
                                setContextMenu({ ...contextMenu, visible: false });
                            }}
                            className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-sky-400 transition-colors"
                        >
                            <RotateCcw size={13} className="text-textFaint" />
                            <span>Restart Container</span>
                        </button>
                    )}

                    <div className="h-[1px] bg-borderDark my-1"></div>

                    <button
                        onClick={() => {
                            setRenamingNodeId(contextMenu.node!.id);
                            setRenameValue(contextMenu.node!.alias || contextMenu.node!.name);
                            setContextMenu({ ...contextMenu, visible: false });
                        }}
                        className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                    >
                        <div className="flex items-center gap-2">
                            <Edit3 size={13} className="text-textFaint" />
                            <span>Set Custom Alias...</span>
                        </div>
                        <span className="text-[10px] text-textFaint font-mono">F2</span>
                    </button>

                    <button
                        onClick={() => {
                            handleToggleFavorite(contextMenu.node!);
                            setContextMenu({ ...contextMenu, visible: false });
                        }}
                        className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-amber-400 transition-colors"
                    >
                        <Star size={13} className="text-textFaint" />
                        <span>Toggle Favorite</span>
                    </button>
                </div>
            )}

            {/* Docker Live Logs Modal */}
            {logModal.visible && (
                <div className="fixed inset-0 z-50 bg-black/75 flex items-center justify-center p-4">
                    <div className="bg-bgCard border border-borderDark rounded-lg w-[750px] max-w-full h-[500px] flex flex-col shadow-2xl overflow-hidden">
                        <div className="px-4 py-2.5 bg-bgPanel border-b border-borderDark flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <FileText size={14} className="text-sky-400" />
                                <span className="text-xs font-mono font-semibold text-textMain">{logModal.title}</span>
                            </div>
                            <div className="flex items-center gap-2">
                                <button
                                    onClick={() => handleOpenLogs({ resourceId: logModal.containerId, name: logModal.title } as any)}
                                    className="p-1 hover:text-textMain text-textFaint rounded transition-colors"
                                    title="Refresh Logs"
                                >
                                    <RefreshCw size={13} />
                                </button>
                                <button
                                    onClick={() => setLogModal({ ...logModal, visible: false })}
                                    className="text-textFaint hover:text-white text-xs px-2 py-1 rounded bg-bgMain hover:bg-bgHover border border-borderDark"
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
