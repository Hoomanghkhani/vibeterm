import React, { useState, useEffect } from 'react';
import { GetHosts, SaveHost, DeleteHost, SendTerminalInput } from '../wailsjs/go/main/App';
import { 
    Terminal, 
    Server, 
    FolderGit2, 
    Radio, 
    Code2, 
    Network, 
    Settings, 
    Plus, 
    X, 
    Split, 
    LayoutGrid, 
    FolderTree,
    PanelLeftClose,
    PanelLeftOpen,
    Folder,
    Search,
    Command,
    Blocks,
    Activity
} from 'lucide-react';
import { TerminalComponent } from './TerminalComponent';
import { TerminalSplitLayout, TerminalPane, SplitMode } from './TerminalSplitLayout';
import { CommandPalette } from './CommandPalette';
import { HostModal } from './HostModal';
import { HostManagerView } from './HostManagerView';
import { FileExplorerView } from './FileExplorerView';
import { SidebarSFTP } from './SidebarSFTP';
import { EndpointsTree } from './EndpointsTree';
import { TunnelsView } from './TunnelsView';
import { SnippetsView } from './SnippetsView';
import { GitOpsView } from './GitOpsView';
import { ScannerView } from './ScannerView';
import { BroadcastView } from './BroadcastView';
import { PluginsView } from './PluginsView';
import { DiagnosticsView } from './DiagnosticsView';
import { SettingsView } from './SettingsView';
import './index.css';

type ActiveView = 'terminal' | 'manager' | 'files' | 'broadcast' | 'tunnels' | 'snippets' | 'scanner' | 'plugins' | 'diagnostics' | 'git' | 'settings';
type SidebarTab = 'endpoints' | 'sftp';

export interface TerminalTab {
    id: string;
    title: string;
    type: 'local' | 'ssh';
    hostID?: string;
    splitMode: SplitMode;
    panes: TerminalPane[];
    activePaneId: string;
}

function App() {
    const [hosts, setHosts] = useState<any[]>([]);
    const [activeView, setActiveView] = useState<ActiveView>('terminal');
    const [sidebarTab, setSidebarTab] = useState<SidebarTab>('endpoints');

    // Collapsible & Pin-able sidebar state
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

    // Command Palette state
    const [showPalette, setShowPalette] = useState(false);

    // Multi-tab terminal management with Split Panes support
    const [tabs, setTabs] = useState<TerminalTab[]>([
        { 
            id: 'tab-local-init', 
            title: 'Local Shell', 
            type: 'local',
            splitMode: 'none',
            panes: [{ id: 'pane-local-init-0', title: 'Local Shell', type: 'local' }],
            activePaneId: 'pane-local-init-0'
        }
    ]);
    const [activeTabId, setActiveTabId] = useState<string>('tab-local-init');

    // Host Edit / Create Modal State
    const [showHostModal, setShowHostModal] = useState(false);
    const [editingHost, setEditingHost] = useState<any>(null);

    const getHostName = (h: any) => h?.name || h?.Name || 'Unnamed Server';
    const getHostId = (h: any) => h?.id || h?.ID || '';

    const loadHosts = () => {
        GetHosts().then((data) => setHosts(data || [])).catch(console.error);
    };

    useEffect(() => {
        loadHosts();
    }, []);

    // Global Keyboard Shortcuts (Ctrl+Shift+P, Ctrl+K, Ctrl+T, Ctrl+W, Ctrl+1..9)
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if ((e.ctrlKey || e.metaKey) && e.shiftKey && (e.key === 'P' || e.key === 'p')) {
                e.preventDefault();
                setShowPalette((prev) => !prev);
            } else if ((e.ctrlKey || e.metaKey) && (e.key === 'k' || e.key === 'K')) {
                e.preventDefault();
                setShowPalette((prev) => !prev);
            } else if ((e.ctrlKey || e.metaKey) && (e.key === 't' || e.key === 'T')) {
                e.preventDefault();
                handleNewLocalTab();
            } else if ((e.ctrlKey || e.metaKey) && (e.key === 'w' || e.key === 'W') && !e.shiftKey) {
                if (!showHostModal && !showPalette) {
                    e.preventDefault();
                    handleCloseTab(null as any, activeTabId);
                }
            } else if ((e.ctrlKey || e.metaKey) && e.shiftKey && (e.key === 'O' || e.key === 'o')) {
                e.preventDefault();
                handleSplitVertical();
            } else if ((e.ctrlKey || e.metaKey) && e.shiftKey && (e.key === 'E' || e.key === 'e')) {
                e.preventDefault();
                handleSplitHorizontal();
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [activeTabId, showHostModal, showPalette]);

    const handleSaveHost = async (hostData: any) => {
        await SaveHost(hostData);
        loadHosts();
    };

    const handleDeleteHost = async (e: React.MouseEvent, id: string) => {
        if (e) e.stopPropagation();
        await DeleteHost(id);
        const remainingTabs = tabs.filter(t => t.hostID !== id);
        if (remainingTabs.length > 0) {
            setTabs(remainingTabs);
            if (activeTabId && !remainingTabs.some(t => t.id === activeTabId)) {
                setActiveTabId(remainingTabs[0].id);
            }
        }
        loadHosts();
    };

    // Open a new tab for a server
    const handleConnectHost = (host: any) => {
        const id = getHostId(host);
        if (id === 'local') {
            handleNewLocalTab();
            return;
        }

        const name = getHostName(host);
        const newTabId = `tab-ssh-${id}-${Date.now()}`;
        const paneId = `pane-ssh-${id}-${Date.now()}`;
        
        const newTab: TerminalTab = {
            id: newTabId,
            title: name,
            type: 'ssh',
            hostID: id,
            splitMode: 'none',
            panes: [{ id: paneId, title: name, type: 'ssh', hostID: id }],
            activePaneId: paneId,
        };

        setTabs((prev) => [...prev, newTab]);
        setActiveTabId(newTabId);
        setActiveView('terminal');
        setSidebarTab('sftp');
    };

    // Open a new local shell tab
    const handleNewLocalTab = () => {
        const newTabId = `tab-local-${Date.now()}`;
        const paneId = `pane-local-${Date.now()}`;
        const title = `Local Shell ${tabs.filter(t => t.type === 'local').length + 1}`;
        const newTab: TerminalTab = {
            id: newTabId,
            title: title,
            type: 'local',
            splitMode: 'none',
            panes: [{ id: paneId, title: title, type: 'local' }],
            activePaneId: paneId,
        };

        setTabs((prev) => [...prev, newTab]);
        setActiveTabId(newTabId);
        setActiveView('terminal');
    };

    // Split active tab vertically
    const handleSplitVertical = () => {
        setTabs((prev) => prev.map((tab) => {
            if (tab.id !== activeTabId) return tab;
            const newPaneId = `pane-split-v-${Date.now()}`;
            const newPane: TerminalPane = {
                id: newPaneId,
                title: `${tab.title} (Split)`,
                type: tab.type,
                hostID: tab.hostID,
            };
            return {
                ...tab,
                splitMode: 'vertical',
                panes: [...tab.panes, newPane],
                activePaneId: newPaneId,
            };
        }));
    };

    // Split active tab horizontally
    const handleSplitHorizontal = () => {
        setTabs((prev) => prev.map((tab) => {
            if (tab.id !== activeTabId) return tab;
            const newPaneId = `pane-split-h-${Date.now()}`;
            const newPane: TerminalPane = {
                id: newPaneId,
                title: `${tab.title} (Split)`,
                type: tab.type,
                hostID: tab.hostID,
            };
            return {
                ...tab,
                splitMode: 'horizontal',
                panes: [...tab.panes, newPane],
                activePaneId: newPaneId,
            };
        }));
    };

    const handleSelectPane = (paneId: string) => {
        setTabs((prev) => prev.map((tab) => {
            if (tab.id !== activeTabId) return tab;
            return { ...tab, activePaneId: paneId };
        }));
    };

    const handleClosePane = (paneId: string) => {
        setTabs((prev) => prev.map((tab) => {
            if (tab.id !== activeTabId) return tab;
            const remaining = tab.panes.filter(p => p.id !== paneId);
            if (remaining.length === 0) return tab;
            return {
                ...tab,
                panes: remaining,
                splitMode: remaining.length <= 1 ? 'none' : tab.splitMode,
                activePaneId: remaining[0].id,
            };
        }));
    };

    const handleCloseTab = (e: React.MouseEvent | null, tabIdToClose: string) => {
        if (e) e.stopPropagation();
        if (tabs.length === 1) {
            const freshLocalId = `tab-local-${Date.now()}`;
            const freshPaneId = `pane-local-${Date.now()}`;
            const freshLocal: TerminalTab = {
                id: freshLocalId,
                title: 'Local Shell',
                type: 'local',
                splitMode: 'none',
                panes: [{ id: freshPaneId, title: 'Local Shell', type: 'local' }],
                activePaneId: freshPaneId,
            };
            setTabs([freshLocal]);
            setActiveTabId(freshLocal.id);
            return;
        }

        const newTabs = tabs.filter((t) => t.id !== tabIdToClose);
        setTabs(newTabs);

        if (activeTabId === tabIdToClose) {
            setActiveTabId(newTabs[newTabs.length - 1].id);
        }
    };

    const handleEditHost = (host: any) => {
        setEditingHost(host);
        setShowHostModal(true);
    };

    const handleNewHost = () => {
        setEditingHost(null);
        setShowHostModal(true);
    };

    const handleNewHostInFolder = (folderPath: string) => {
        setEditingHost({
            folder: folderPath,
            environment: 'production',
            port: 22,
            protocol: 'ssh',
            username: 'root',
        });
        setShowHostModal(true);
    };

    const handleImportFromScanner = (device: any) => {
        setEditingHost({
            name: `Node-${device.ip.split('.').pop()}`,
            hostname: device.ip,
            port: device.openPorts?.[0] || 22,
            username: 'root',
            protocol: device.matchedProto || 'ssh',
            environment: 'dev',
            folder: 'Discovered',
            tags: ['scanner'],
        });
        setShowHostModal(true);
    };

    const handleRunSnippet = (command: string) => {
        setActiveView('terminal');
        const currentActive = tabs.find(t => t.id === activeTabId);
        const sid = currentActive?.hostID || 'local';
        setTimeout(() => {
            SendTerminalInput(sid, command + '\n');
        }, 100);
    };

    const currentTab = tabs.find((t) => t.id === activeTabId);

    return (
        <div className="flex flex-col h-screen bg-bgMain text-textMain font-sans overflow-hidden select-none">
            {/* Top Minimal Header Bar with VS Code-grade Command Launcher */}
            <div className="h-9 bg-bgSidebar border-b border-borderDark flex items-center justify-between px-3 shrink-0" style={{ '--wails-draggable': 'drag' } as any}>
                <div className="flex items-center gap-2">
                    <Terminal size={15} className="text-textMain" />
                    <span className="text-[11px] font-semibold tracking-widest text-textFaint uppercase">Workspace</span>
                </div>

                {/* Quick Command Launcher Pill */}
                <div 
                    onClick={() => setShowPalette(true)}
                    className="flex items-center gap-2 bg-bgCard hover:bg-bgPanel px-3 py-1 rounded-lg border border-borderDark text-xs text-textFaint hover:text-textMain cursor-pointer transition-colors"
                    style={{ '--wails-draggable': 'no-drag' } as any}
                    title="Open Command Palette (Ctrl+Shift+P / Ctrl+K)"
                >
                    <Search size={12} />
                    <span className="text-[11px]">Search commands, hosts, tools...</span>
                    <kbd className="text-[10px] font-mono bg-bgMain px-1 py-0.2 rounded border border-borderDark">Ctrl+K</kbd>
                </div>

                <div className="flex items-center gap-1 text-textMuted" style={{ '--wails-draggable': 'no-drag' } as any}>
                    {activeView === 'terminal' && (
                        <button 
                            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
                            className={`p-1 hover:text-textMain hover:bg-bgPanel rounded transition-colors ${sidebarCollapsed ? 'text-textMain bg-bgPanel' : 'text-textFaint'}`}
                            title={sidebarCollapsed ? "Show Sidebar" : "Hide Sidebar"}
                        >
                            {sidebarCollapsed ? <PanelLeftOpen size={14} /> : <PanelLeftClose size={14} />}
                        </button>
                    )}
                    <button 
                        onClick={handleNewLocalTab}
                        className="p-1 hover:text-textMain hover:bg-bgPanel rounded transition-colors text-textFaint" 
                        title="New Terminal Tab (Ctrl+T)"
                    >
                        <Plus size={13} />
                    </button>
                    <button 
                        onClick={handleSplitVertical}
                        className="p-1 hover:text-textMain hover:bg-bgPanel rounded transition-colors text-textFaint" 
                        title="Split Pane Vertically (Ctrl+Shift+O)"
                    >
                        <Split size={13} />
                    </button>
                    <button 
                        onClick={handleSplitHorizontal}
                        className="p-1 hover:text-textMain hover:bg-bgPanel rounded transition-colors text-textFaint" 
                        title="Split Pane Horizontally (Ctrl+Shift+E)"
                    >
                        <Split size={13} className="rotate-90" />
                    </button>
                    <button 
                        onClick={() => setActiveView('settings')}
                        className={`p-1 hover:text-textMain hover:bg-bgPanel rounded transition-colors text-textFaint ${activeView === 'settings' ? 'text-textMain bg-bgPanel' : ''}`}
                        title="Settings"
                    >
                        <Settings size={13} />
                    </button>
                </div>
            </div>

            {/* Main Application Area */}
            <div className="flex flex-1 overflow-hidden">
                {/* Left Activity Bar */}
                <div className="w-12 bg-bgSidebar border-r border-borderDark flex flex-col items-center py-2 shrink-0 z-10">
                    <ActivityButton 
                        icon={<Terminal size={19} strokeWidth={1.7} />} 
                        active={activeView === 'terminal'} 
                        onClick={() => {
                            if (activeView === 'terminal') {
                                setSidebarCollapsed(!sidebarCollapsed);
                            } else {
                                setActiveView('terminal');
                                setSidebarCollapsed(false);
                            }
                        }} 
                        tooltip="Terminal & Explorer (Ctrl+1)"
                    />
                    <ActivityButton 
                        icon={<LayoutGrid size={19} strokeWidth={1.7} />} 
                        active={activeView === 'manager'} 
                        onClick={() => setActiveView('manager')} 
                        tooltip="Endpoints Manager (Ctrl+2)"
                    />
                    <ActivityButton 
                        icon={<FolderTree size={19} strokeWidth={1.7} />} 
                        active={activeView === 'files'} 
                        onClick={() => setActiveView('files')} 
                        tooltip="Full-Screen File Explorer (Ctrl+3)"
                    />
                    <ActivityButton 
                        icon={<Radio size={19} strokeWidth={1.7} />} 
                        active={activeView === 'broadcast'} 
                        onClick={() => setActiveView('broadcast')} 
                        tooltip="Broadcast Multi-Terminal"
                    />
                    <ActivityButton 
                        icon={<Network size={19} strokeWidth={1.7} />} 
                        active={activeView === 'tunnels'} 
                        onClick={() => setActiveView('tunnels')} 
                        tooltip="Port Forwarding & Tunnels"
                    />
                    <ActivityButton 
                        icon={<Code2 size={19} strokeWidth={1.7} />} 
                        active={activeView === 'snippets'} 
                        onClick={() => setActiveView('snippets')} 
                        tooltip="Automation Snippets"
                    />
                    <ActivityButton 
                        icon={<Activity size={19} strokeWidth={1.7} />} 
                        active={activeView === 'diagnostics'} 
                        onClick={() => setActiveView('diagnostics')} 
                        tooltip="Network Diagnostics"
                    />
                    <ActivityButton 
                        icon={<Blocks size={19} strokeWidth={1.7} />} 
                        active={activeView === 'plugins'} 
                        onClick={() => setActiveView('plugins')} 
                        tooltip="Extensions & Discovery"
                    />
                    <ActivityButton 
                        icon={<FolderGit2 size={19} strokeWidth={1.7} />} 
                        active={activeView === 'git'} 
                        onClick={() => setActiveView('git')} 
                        tooltip="GitOps Vault"
                    />
                    <div className="mt-auto mb-1">
                        <ActivityButton 
                            icon={<Settings size={19} strokeWidth={1.7} />} 
                            active={activeView === 'settings'} 
                            onClick={() => setActiveView('settings')} 
                            tooltip="Settings & Vault"
                        />
                    </div>
                </div>

                {/* Primary Sidebar */}
                {activeView === 'terminal' && !sidebarCollapsed && (
                    <div className="w-64 bg-bgSidebar border-r border-borderDark flex flex-col shrink-0 overflow-hidden transition-all duration-150">
                        {/* Sidebar Top Tab Switcher */}
                        <div className="h-9 px-2 border-b border-borderDark/60 flex items-center justify-between bg-bgCard shrink-0">
                            <div className="flex items-center gap-1 bg-bgMain p-0.5 rounded-lg border border-borderDark text-[11px]">
                                <button
                                    onClick={() => setSidebarTab('endpoints')}
                                    className={`px-2.5 py-0.5 rounded font-medium transition-colors ${sidebarTab === 'endpoints' ? 'bg-bgPanel text-textMain shadow-sm font-semibold' : 'text-textFaint hover:text-textMuted'}`}
                                >
                                    Endpoints
                                </button>
                                <button
                                    onClick={() => setSidebarTab('sftp')}
                                    className={`px-2.5 py-0.5 rounded font-medium transition-colors flex items-center gap-1 ${sidebarTab === 'sftp' ? 'bg-bgPanel text-textMain shadow-sm font-semibold' : 'text-textFaint hover:text-textMuted'}`}
                                >
                                    <Folder size={11} />
                                    <span>Files (SFTP)</span>
                                </button>
                            </div>

                            <div className="flex items-center gap-0.5">
                                {sidebarTab === 'endpoints' && (
                                    <button onClick={handleNewHost} className="p-1 hover:text-textMain rounded hover:bg-bgPanel text-textFaint transition-colors" title="New Endpoint">
                                        <Plus size={13} />
                                    </button>
                                )}
                                <button 
                                    onClick={() => setSidebarCollapsed(true)} 
                                    className="p-1 hover:text-textMain rounded hover:bg-bgPanel text-textFaint transition-colors"
                                    title="Collapse Sidebar"
                                >
                                    <PanelLeftClose size={13} />
                                </button>
                            </div>
                        </div>

                        {/* Tab Content 1: Nested Folders & Endpoints Tree */}
                        {sidebarTab === 'endpoints' && (
                            <EndpointsTree 
                                hosts={hosts}
                                activeHostId={currentTab?.hostID}
                                onConnectHost={handleConnectHost}
                                onEditHost={handleEditHost}
                                onDeleteHost={handleDeleteHost}
                                onNewHostInFolder={handleNewHostInFolder}
                                onReloadHosts={loadHosts}
                            />
                        )}

                        {/* Tab Content 2: Live SFTP Files for Active Session */}
                        {sidebarTab === 'sftp' && (
                            <SidebarSFTP 
                                hosts={hosts}
                                activeHostId={currentTab?.hostID}
                            />
                        )}
                    </div>
                )}

                {/* Main View Area */}
                {activeView === 'terminal' && (
                    <div className="flex-1 bg-bgMain flex flex-col overflow-hidden min-w-0">
                        {/* Dynamic Tab Header Bar */}
                        <div className="h-9 bg-bgSidebar border-b border-borderDark flex items-center overflow-x-auto no-scrollbar shrink-0">
                            {tabs.map((tab) => {
                                const isActive = tab.id === activeTabId;
                                return (
                                    <div
                                        key={tab.id}
                                        onClick={() => setActiveTabId(tab.id)}
                                        className={`h-full px-3.5 flex items-center gap-2 border-r border-borderDark text-xs font-medium cursor-pointer transition-colors ${
                                            isActive 
                                                ? 'bg-bgMain border-t-2 border-t-textMain text-textMain' 
                                                : 'bg-bgSidebar text-textFaint hover:text-textMuted hover:bg-bgPanel/50'
                                        }`}
                                    >
                                        <Terminal size={12} className={isActive ? 'text-textMain' : 'text-textFaint'} />
                                        <span className="max-w-[120px] truncate">{tab.title}</span>
                                        {tab.panes.length > 1 && (
                                            <span className="text-[10px] text-textFaint font-mono bg-bgPanel px-1 rounded border border-borderDark">
                                                {tab.panes.length}P
                                            </span>
                                        )}
                                        <button 
                                            onClick={(e) => handleCloseTab(e, tab.id)}
                                            className="p-0.5 hover:bg-bgPanel rounded text-textFaint hover:text-textMain ml-1 transition-colors"
                                            title="Close Tab"
                                        >
                                            <X size={11} />
                                        </button>
                                    </div>
                                );
                            })}

                            {/* Add Tab Button */}
                            <button
                                onClick={handleNewLocalTab}
                                className="px-2.5 h-full flex items-center text-textFaint hover:text-textMain hover:bg-bgPanel/50 transition-colors"
                                title="New Local Tab (Ctrl+T)"
                            >
                                <Plus size={13} />
                            </button>
                        </div>

                        {/* Terminals Split Container */}
                        <div className="flex-1 overflow-hidden bg-bgMain relative">
                            {tabs.map((tab) => (
                                <div
                                    key={tab.id}
                                    className={`w-full h-full ${tab.id === activeTabId ? 'block' : 'hidden'}`}
                                >
                                    <TerminalSplitLayout 
                                        tabId={tab.id}
                                        panes={tab.panes}
                                        splitMode={tab.splitMode}
                                        activePaneId={tab.activePaneId}
                                        onSelectPane={handleSelectPane}
                                        onClosePane={handleClosePane}
                                        onSplitPane={(mode) => {
                                            setTabs(prev => prev.map(t => t.id === tab.id ? { ...t, splitMode: mode } : t));
                                        }}
                                    />
                                </div>
                            ))}
                        </div>

                        {/* Status Bar */}
                        <div className="h-6 bg-bgSidebar border-t border-borderDark flex items-center px-3.5 text-[11px] text-textMuted justify-between shrink-0 select-none">
                            <div className="flex items-center gap-4">
                                <span className="flex items-center gap-1.5 hover:text-textMain cursor-pointer font-medium">
                                    <span className="w-2 h-2 rounded-full bg-emerald-400"></span> {currentTab?.title}
                                </span>
                                <span className="hover:text-textMain cursor-pointer">UTF-8</span>
                                <span className="hover:text-textMain cursor-pointer">PTY Multiplex</span>
                                {currentTab && currentTab.panes.length > 1 && (
                                    <span className="hover:text-textMain cursor-pointer font-mono">Split: {currentTab.splitMode}</span>
                                )}
                            </div>
                            <div className="flex items-center gap-3">
                                <span className="hover:text-textMain cursor-pointer">{tabs.length} open tab{tabs.length > 1 ? 's' : ''}</span>
                                <span className="hover:text-textMain cursor-pointer">AES-256-GCM</span>
                            </div>
                        </div>
                    </div>
                )}

                {activeView === 'manager' && (
                    <HostManagerView 
                        hosts={hosts} 
                        onConnect={handleConnectHost} 
                        onEdit={handleEditHost} 
                        onDelete={(id) => DeleteHost(id).then(loadHosts)}
                        onNew={handleNewHost}
                    />
                )}

                {activeView === 'files' && (
                    <FileExplorerView 
                        hosts={hosts} 
                        activeHostId={currentTab?.hostID}
                    />
                )}

                {activeView === 'broadcast' && (
                    <BroadcastView 
                        hosts={hosts} 
                        onConnectHost={handleConnectHost}
                    />
                )}

                {activeView === 'tunnels' && (
                    <TunnelsView hosts={hosts} />
                )}

                {activeView === 'snippets' && (
                    <SnippetsView onRunInTerminal={handleRunSnippet} />
                )}

                {activeView === 'diagnostics' && (
                    <DiagnosticsView />
                )}

                {activeView === 'plugins' && (
                    <PluginsView onReloadHosts={loadHosts} />
                )}

                {activeView === 'scanner' && (
                    <ScannerView onImportHost={handleImportFromScanner} />
                )}

                {activeView === 'git' && (
                    <GitOpsView />
                )}

                {activeView === 'settings' && (
                    <SettingsView />
                )}
            </div>

            {/* Host Edit / Create Modal */}
            <HostModal 
                isOpen={showHostModal} 
                onClose={() => setShowHostModal(false)} 
                onSave={handleSaveHost}
                initialHost={editingHost}
            />

            {/* VS Code Grade Command Palette */}
            <CommandPalette 
                isOpen={showPalette}
                onClose={() => setShowPalette(false)}
                hosts={hosts}
                onConnectHost={handleConnectHost}
                onNewLocalTab={handleNewLocalTab}
                onSplitHorizontal={handleSplitHorizontal}
                onSplitVertical={handleSplitVertical}
                onCloseActiveTab={() => handleCloseTab(null, activeTabId)}
                onNavigate={(v) => setActiveView(v as ActiveView)}
            />
        </div>
    );
}

function ActivityButton({ icon, active, onClick, tooltip }: { icon: any, active: boolean, onClick: () => void, tooltip: string }) {
    return (
        <button 
            onClick={onClick}
            title={tooltip}
            className={`w-full h-11 flex items-center justify-center relative transition-colors ${active ? 'text-textMain' : 'text-textFaint hover:text-textMuted'}`}
        >
            {active && (
                <div className="absolute left-0 top-1/2 -translate-y-1/2 w-[2px] h-6 bg-textMain rounded-r"></div>
            )}
            {icon}
        </button>
    );
}

export default App;
