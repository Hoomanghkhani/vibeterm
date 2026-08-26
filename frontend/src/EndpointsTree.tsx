import React, { useState, useRef, useEffect } from 'react';
import { 
    Folder, 
    FolderOpen, 
    ChevronRight, 
    ChevronDown, 
    Terminal, 
    Edit3, 
    Trash2, 
    Plus, 
    FolderPlus, 
    RefreshCw, 
    FoldHorizontal, 
    Check, 
    X,
    Server,
    Copy,
    ExternalLink
} from 'lucide-react';
import { RenameFolder, DeleteHost } from '../wailsjs/go/main/App';

interface EndpointsTreeProps {
    hosts: any[];
    activeHostId?: string;
    onConnectHost: (host: any) => void;
    onEditHost: (host: any) => void;
    onDeleteHost: (e: React.MouseEvent, id: string) => void;
    onNewHostInFolder: (folderPath: string) => void;
    onReloadHosts: () => void;
}

interface FolderNode {
    name: string;
    fullPath: string;
    subFolders: { [key: string]: FolderNode };
    hosts: any[];
}

export const EndpointsTree: React.FC<EndpointsTreeProps> = ({
    hosts = [],
    activeHostId,
    onConnectHost,
    onEditHost,
    onDeleteHost,
    onNewHostInFolder,
    onReloadHosts,
}) => {
    // Collapsed folders set
    const [collapsedFolders, setCollapsedFolders] = useState<{ [path: string]: boolean }>({});
    
    // Inline creation state (VS Code style)
    const [creatingInFolder, setCreatingInFolder] = useState<string | null>(null); // null = not creating, '' = root, 'path' = inside folder
    const [newFolderName, setNewFolderName] = useState('');

    // Inline renaming state (F2 style)
    const [renamingPath, setRenamingPath] = useState<string | null>(null);
    const [renameValue, setRenameValue] = useState<string>('');

    // Right-click context menu
    const [contextMenu, setContextMenu] = useState<{
        visible: boolean;
        x: number;
        y: number;
        type: 'folder' | 'host' | 'root';
        targetPath?: string;
        targetHost?: any;
    }>({ visible: false, x: 0, y: 0, type: 'root' });

    const getHostName = (h: any) => h?.name || h?.Name || 'Server';
    const getHostId = (h: any) => h?.id || h?.ID || '';
    const getHostIp = (h: any) => h?.hostname || h?.Hostname || '';
    const getHostFolder = (h: any) => h?.folder || h?.Folder || '';

    // Build recursive folder tree
    const rootTree: FolderNode = {
        name: 'root',
        fullPath: '',
        subFolders: {},
        hosts: [],
    };

    hosts.forEach((host) => {
        const folderPath = getHostFolder(host).trim();
        if (!folderPath) {
            rootTree.hosts.push(host);
        } else {
            const segments = folderPath.split('/').map((s: string) => s.trim()).filter(Boolean);
            let current = rootTree;
            let currentPathAcc = '';

            segments.forEach((seg: string) => {
                currentPathAcc = currentPathAcc ? `${currentPathAcc}/${seg}` : seg;
                if (!current.subFolders[seg]) {
                    current.subFolders[seg] = {
                        name: seg,
                        fullPath: currentPathAcc,
                        subFolders: {},
                        hosts: [],
                    };
                }
                current = current.subFolders[seg];
            });

            current.hosts.push(host);
        }
    });

    const toggleFolder = (path: string) => {
        setCollapsedFolders((prev) => ({
            ...prev,
            [path]: !prev[path],
        }));
    };

    const collapseAll = () => {
        const allPaths: { [path: string]: boolean } = {};
        const collect = (node: FolderNode) => {
            if (node.fullPath) allPaths[node.fullPath] = true;
            Object.values(node.subFolders).forEach(collect);
        };
        collect(rootTree);
        setCollapsedFolders(allPaths);
    };

    const handleCommitNewFolder = () => {
        const folderToCreate = newFolderName.trim();
        if (!folderToCreate) {
            setCreatingInFolder(null);
            return;
        }

        const parentPath = creatingInFolder || '';
        const fullNewPath = parentPath 
            ? `${parentPath}/${folderToCreate}`
            : folderToCreate;

        onNewHostInFolder(fullNewPath);
        setCreatingInFolder(null);
        setNewFolderName('');
    };

    const handleStartRename = (path: string, currentName: string) => {
        setRenamingPath(path);
        setRenameValue(currentName);
        setContextMenu({ ...contextMenu, visible: false });
    };

    const handleSaveRename = async (oldFullPath: string) => {
        if (!renameValue.trim()) {
            setRenamingPath(null);
            return;
        }

        const segments = oldFullPath.split('/');
        segments[segments.length - 1] = renameValue.trim();
        const newFullPath = segments.join('/');

        if (oldFullPath !== newFullPath) {
            await RenameFolder(oldFullPath, newFullPath);
            onReloadHosts();
        }
        setRenamingPath(null);
    };

    // Global click listener to dismiss context menu
    useEffect(() => {
        const dismiss = () => setContextMenu((prev) => (prev.visible ? { ...prev, visible: false } : prev));
        window.addEventListener('click', dismiss);
        return () => window.removeEventListener('click', dismiss);
    }, []);

    const handleFolderContextMenu = (e: React.MouseEvent, node: FolderNode) => {
        e.preventDefault();
        e.stopPropagation();
        setContextMenu({
            visible: true,
            x: e.clientX,
            y: e.clientY,
            type: 'folder',
            targetPath: node.fullPath,
        });
    };

    const handleHostContextMenu = (e: React.MouseEvent, host: any) => {
        e.preventDefault();
        e.stopPropagation();
        setContextMenu({
            visible: true,
            x: e.clientX,
            y: e.clientY,
            type: 'host',
            targetHost: host,
        });
    };

    // Render a single folder node and its recursive children
    const renderFolder = (node: FolderNode, depth: number = 0) => {
        const isCollapsed = !!collapsedFolders[node.fullPath];
        const isRenaming = renamingPath === node.fullPath;
        const isCreatingHere = creatingInFolder === node.fullPath;
        const totalItems = countTotalHosts(node);

        return (
            <div key={node.fullPath} className="flex flex-col select-none">
                {/* Folder Row */}
                <div 
                    className="group px-2 py-1 flex items-center justify-between text-xs hover:bg-bgPanel cursor-pointer transition-colors"
                    style={{ paddingLeft: `${Math.max(8, depth * 12 + 8)}px` }}
                    onClick={() => toggleFolder(node.fullPath)}
                    onContextMenu={(e) => handleFolderContextMenu(e, node)}
                >
                    <div className="flex items-center gap-1.5 min-w-0 flex-1">
                        <span className="text-textFaint group-hover:text-textMain transition-transform">
                            {isCollapsed ? <ChevronRight size={12} /> : <ChevronDown size={12} />}
                        </span>
                        {isCollapsed ? (
                            <Folder size={14} className="text-zinc-400 fill-zinc-400/20 shrink-0" />
                        ) : (
                            <FolderOpen size={14} className="text-zinc-300 fill-zinc-300/30 shrink-0" />
                        )}

                        {isRenaming ? (
                            <div className="flex items-center gap-1 min-w-0" onClick={(e) => e.stopPropagation()}>
                                <input 
                                    type="text"
                                    value={renameValue}
                                    onChange={(e) => setRenameValue(e.target.value)}
                                    autoFocus
                                    onBlur={() => handleSaveRename(node.fullPath)}
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter') handleSaveRename(node.fullPath);
                                        if (e.key === 'Escape') setRenamingPath(null);
                                    }}
                                    className="bg-bgMain border border-borderActive rounded px-1 py-0.5 text-xs text-textMain outline-none w-28 font-mono"
                                />
                            </div>
                        ) : (
                            <span className="truncate text-textMain font-medium text-xs">
                                {node.name}
                            </span>
                        )}
                        <span className="text-[10px] text-textFaint font-mono">({totalItems})</span>
                    </div>

                    {/* VS Code Hover Actions */}
                    <div className="hidden group-hover:flex items-center gap-0.5 shrink-0 ml-1">
                        <button 
                            onClick={(e) => { e.stopPropagation(); onNewHostInFolder(node.fullPath); }}
                            className="p-1 text-textFaint hover:text-textMain rounded hover:bg-bgMain transition-colors"
                            title="New Endpoint in folder"
                        >
                            <Plus size={12} />
                        </button>
                        <button 
                            onClick={(e) => { 
                                e.stopPropagation(); 
                                setCollapsedFolders((prev) => ({ ...prev, [node.fullPath]: false }));
                                setCreatingInFolder(node.fullPath); 
                                setNewFolderName('');
                            }}
                            className="p-1 text-textFaint hover:text-textMain rounded hover:bg-bgMain transition-colors"
                            title="New Subfolder"
                        >
                            <FolderPlus size={12} />
                        </button>
                    </div>
                </div>

                {/* Sub-items with VS Code Indentation Guide */}
                {!isCollapsed && (
                    <div className="flex flex-col border-l border-borderDark/30 ml-3.5">
                        {/* Inline folder creation row inside this folder */}
                        {isCreatingHere && (
                            <div 
                                className="px-2 py-1 flex items-center gap-1.5 text-xs bg-bgPanel"
                                style={{ paddingLeft: `${Math.max(6, depth * 12 + 6)}px` }}
                            >
                                <Folder size={14} className="text-zinc-400 fill-zinc-400/20 shrink-0" />
                                <input 
                                    type="text" 
                                    placeholder="folder-name" 
                                    value={newFolderName}
                                    onChange={(e) => setNewFolderName(e.target.value)}
                                    autoFocus
                                    onBlur={handleCommitNewFolder}
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter') handleCommitNewFolder();
                                        if (e.key === 'Escape') setCreatingInFolder(null);
                                    }}
                                    className="flex-1 bg-bgMain border border-borderActive rounded px-1 py-0.5 text-xs text-textMain outline-none font-mono"
                                />
                            </div>
                        )}

                        {/* Nested Sub-Folders */}
                        {Object.values(node.subFolders).map((sub) => renderFolder(sub, depth + 1))}

                        {/* Hosts directly in this folder */}
                        {node.hosts.map((h, idx) => renderHostItem(h, depth + 1, idx))}
                    </div>
                )}
            </div>
        );
    };

    const renderHostItem = (host: any, depth: number, idx: number) => {
        const id = getHostId(host) || `host-${idx}`;
        const name = getHostName(host);
        const ip = getHostIp(host);
        const isOnline = host.health === 'online' || host.Health === 'online' || host.Health === 1;
        const isActive = activeHostId === id;

        return (
            <div
                key={id}
                onClick={() => onConnectHost(host)}
                onContextMenu={(e) => handleHostContextMenu(e, host)}
                className={`group py-1 pr-2 flex items-center justify-between text-xs cursor-pointer transition-colors ${
                    isActive 
                        ? 'bg-bgPanel text-textMain font-medium' 
                        : 'text-textMuted hover:bg-bgHover hover:text-textMain'
                }`}
                style={{ paddingLeft: `${Math.max(10, depth * 12 + 10)}px` }}
            >
                <div className="flex items-center gap-2 min-w-0">
                    <div className={`w-2 h-2 rounded-full shrink-0 ${isOnline ? 'bg-emerald-400' : 'bg-zinc-600'}`} />
                    <span className="truncate text-textMain font-sans text-xs">{name}</span>
                </div>

                <div className="flex items-center gap-1 shrink-0 ml-2">
                    <span className="text-[10px] text-textFaint font-mono truncate group-hover:hidden">{ip}</span>
                    <button 
                        onClick={(e) => { e.stopPropagation(); onEditHost(host); }}
                        className="hidden group-hover:block p-0.5 text-textFaint hover:text-textMain rounded transition-colors"
                        title="Edit Host"
                    >
                        <Edit3 size={12} />
                    </button>
                    <button 
                        onClick={(e) => onDeleteHost(e, id)}
                        className="hidden group-hover:block p-0.5 text-textFaint hover:text-rose-400 rounded transition-colors"
                        title="Delete Host"
                    >
                        <Trash2 size={12} />
                    </button>
                </div>
            </div>
        );
    };

    const countTotalHosts = (node: FolderNode): number => {
        let count = node.hosts.length;
        Object.values(node.subFolders).forEach((sub) => {
            count += countTotalHosts(sub);
        });
        return count;
    };

    return (
        <div className="flex-1 flex flex-col overflow-y-auto py-1 select-none relative">
            {/* Top Local Shell Quick Item */}
            <div 
                onClick={() => onConnectHost({ ID: 'local', Name: 'Local Shell', type: 'local' })}
                className={`group px-3 py-1.5 cursor-pointer flex items-center justify-between text-xs transition-colors ${
                    !activeHostId ? 'bg-bgPanel text-textMain font-medium' : 'text-textMuted hover:bg-bgHover hover:text-textMain'
                }`}
            >
                <div className="flex items-center gap-2.5 min-w-0">
                    <Terminal size={14} className="text-textMuted shrink-0" />
                    <span className="truncate text-textMain">Local Shell</span>
                </div>
                <span className="text-[9px] text-textFaint font-mono bg-bgPanel px-1.5 py-0.5 rounded border border-borderDark">native</span>
            </div>

            <div className="h-[1px] bg-borderDark/40 my-1 mx-2"></div>

            {/* VS Code Section Header Toolbar */}
            <div className="px-3 py-1 flex items-center justify-between text-[11px] font-semibold text-textFaint uppercase tracking-wider group">
                <span>Servers</span>
                
                {/* Standard VS Code Explorer Action Icons */}
                <div className="flex items-center gap-0.5 text-textFaint">
                    <button 
                        onClick={() => onNewHostInFolder('')}
                        className="p-1 hover:text-textMain rounded hover:bg-bgPanel transition-colors"
                        title="New Endpoint"
                    >
                        <Plus size={13} />
                    </button>
                    <button 
                        onClick={() => { setCreatingInFolder(''); setNewFolderName(''); }}
                        className="p-1 hover:text-textMain rounded hover:bg-bgPanel transition-colors"
                        title="New Folder"
                    >
                        <FolderPlus size={13} />
                    </button>
                    <button 
                        onClick={onReloadHosts}
                        className="p-1 hover:text-textMain rounded hover:bg-bgPanel transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw size={12} />
                    </button>
                    <button 
                        onClick={collapseAll}
                        className="p-1 hover:text-textMain rounded hover:bg-bgPanel transition-colors"
                        title="Collapse Folders"
                    >
                        <FoldHorizontal size={13} />
                    </button>
                </div>
            </div>

            {/* Root Inline Folder Input (VS Code style) */}
            {creatingInFolder === '' && (
                <div className="px-3 py-1 flex items-center gap-1.5 text-xs bg-bgPanel my-0.5">
                    <Folder size={14} className="text-zinc-400 fill-zinc-400/20 shrink-0" />
                    <input 
                        type="text" 
                        placeholder="folder-name" 
                        value={newFolderName}
                        onChange={(e) => setNewFolderName(e.target.value)}
                        autoFocus
                        onBlur={handleCommitNewFolder}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') handleCommitNewFolder();
                            if (e.key === 'Escape') setCreatingInFolder(null);
                        }}
                        className="flex-1 bg-bgMain border border-borderActive rounded px-1 py-0.5 text-xs text-textMain outline-none font-mono"
                    />
                </div>
            )}

            {/* Render Folders Tree */}
            {Object.values(rootTree.subFolders).map((sub) => renderFolder(sub, 0))}

            {/* Render Root Hosts */}
            {rootTree.hosts.map((h, idx) => renderHostItem(h, 0, idx))}

            {/* VS Code Right Click Context Menu */}
            {contextMenu.visible && (
                <div 
                    className="fixed z-50 w-44 bg-bgCard border border-borderDark rounded-lg shadow-2xl py-1 text-xs text-textMain overflow-hidden select-none"
                    style={{ top: `${contextMenu.y}px`, left: `${contextMenu.x}px` }}
                    onClick={(e) => e.stopPropagation()}
                >
                    {contextMenu.type === 'folder' && (
                        <>
                            <button
                                onClick={() => {
                                    onNewHostInFolder(contextMenu.targetPath || '');
                                    setContextMenu({ ...contextMenu, visible: false });
                                }}
                                className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                            >
                                <Plus size={13} className="text-textFaint" />
                                <span>New Endpoint...</span>
                            </button>
                            <button
                                onClick={() => {
                                    setCreatingInFolder(contextMenu.targetPath || '');
                                    setContextMenu({ ...contextMenu, visible: false });
                                }}
                                className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                            >
                                <FolderPlus size={13} className="text-textFaint" />
                                <span>New Subfolder...</span>
                            </button>
                            <button
                                onClick={() => {
                                    const path = contextMenu.targetPath || '';
                                    const name = path.split('/').pop() || '';
                                    handleStartRename(path, name);
                                }}
                                className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                            >
                                <div className="flex items-center gap-2">
                                    <Edit3 size={13} className="text-textFaint" />
                                    <span>Rename</span>
                                </div>
                                <span className="text-[10px] text-textFaint font-mono">F2</span>
                            </button>
                        </>
                    )}

                    {contextMenu.type === 'host' && (
                        <>
                            <button
                                onClick={() => {
                                    onConnectHost(contextMenu.targetHost);
                                    setContextMenu({ ...contextMenu, visible: false });
                                }}
                                className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                            >
                                <Terminal size={13} className="text-textFaint" />
                                <span>Connect (New Tab)</span>
                            </button>
                            <button
                                onClick={() => {
                                    onEditHost(contextMenu.targetHost);
                                    setContextMenu({ ...contextMenu, visible: false });
                                }}
                                className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-white transition-colors"
                            >
                                <Edit3 size={13} className="text-textFaint" />
                                <span>Edit Endpoint...</span>
                            </button>
                            <div className="h-[1px] bg-borderDark my-1"></div>
                            <button
                                onClick={(e) => {
                                    onDeleteHost(e, getHostId(contextMenu.targetHost));
                                    setContextMenu({ ...contextMenu, visible: false });
                                }}
                                className="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-bgHover hover:text-rose-400 transition-colors"
                            >
                                <Trash2 size={13} className="text-textFaint" />
                                <span>Delete</span>
                            </button>
                        </>
                    )}
                </div>
            )}
        </div>
    );
};
