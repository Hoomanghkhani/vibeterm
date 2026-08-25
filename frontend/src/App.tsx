import { useState, useEffect } from 'react';
import { GetHosts, SaveHost, DeleteHost } from '../wailsjs/go/main/App';
import { Terminal, Server, FolderGit2, Play, Settings, Search, Plus, X, ChevronDown, ChevronRight, MoreHorizontal, Split, LayoutGrid, Edit3, Trash2 } from 'lucide-react';
import { TerminalComponent } from './TerminalComponent';
import { HostModal } from './HostModal';
import { HostManagerView } from './HostManagerView';
import './index.css';

function App() {
    const [hosts, setHosts] = useState<any[]>([]);
    const [activeTab, setActiveTab] = useState('local');
    const [activeHostID, setActiveHostID] = useState<string | undefined>(undefined);
    const [activeView, setActiveView] = useState<'terminal' | 'manager' | 'git' | 'automation'>('terminal');
    const [hostsExpanded, setHostsExpanded] = useState(true);

    // Host Edit / Create Modal State
    const [showHostModal, setShowHostModal] = useState(false);
    const [editingHost, setEditingHost] = useState<any>(null);

    const loadHosts = () => {
        GetHosts().then((data) => setHosts(data)).catch(console.error);
    };

    useEffect(() => {
        loadHosts();
    }, []);

    const handleSaveHost = async (hostData: any) => {
        await SaveHost(hostData);
        loadHosts();
    };

    const handleDeleteHost = async (e: React.MouseEvent, id: string) => {
        e.stopPropagation();
        await DeleteHost(id);
        if (activeHostID === id) {
            setActiveTab('local');
            setActiveHostID(undefined);
        }
        loadHosts();
    };

    const handleConnectHost = (host: any) => {
        setActiveTab(host.Name);
        setActiveHostID(host.ID);
        setActiveView('terminal');
    };

    const handleEditHost = (host: any) => {
        setEditingHost(host);
        setShowHostModal(true);
    };

    const handleNewHost = () => {
        setEditingHost(null);
        setShowHostModal(true);
    };

    const handleSelectLocal = () => {
        setActiveTab('local');
        setActiveHostID(undefined);
        setActiveView('terminal');
    };

    return (
        <div className="flex flex-col h-screen bg-bgMain text-textMain font-sans overflow-hidden select-none">
            {/* Top Unified Header Bar */}
            <div className="h-10 bg-bgSidebar border-b border-borderDark flex items-center justify-between px-3 shrink-0" style={{ '--wails-draggable': 'drag' } as any}>
                <div className="flex items-center gap-2">
                    <Terminal size={16} className="text-textMain" />
                </div>

                {/* Command Palette */}
                <div className="w-80 h-7 bg-bgMain border border-borderDark rounded-md flex items-center px-3 text-xs text-textMuted justify-between hover:border-borderActive transition-colors cursor-pointer" style={{ '--wails-draggable': 'no-drag' } as any}>
                    <div className="flex items-center gap-2 w-full">
                        <Search size={13} className="text-textFaint" />
                        <span className="text-textMuted text-xs font-normal">Type a command or search...</span>
                    </div>
                    <kbd className="text-[10px] font-mono text-textFaint px-1.5 py-0.5 rounded bg-bgPanel border border-borderDark">Ctrl+K</kbd>
                </div>

                <div className="flex items-center gap-2 text-textMuted" style={{ '--wails-draggable': 'no-drag' } as any}>
                    <button className="p-1.5 hover:text-textMain hover:bg-bgPanel rounded transition-colors">
                        <Split size={14} />
                    </button>
                    <button className="p-1.5 hover:text-textMain hover:bg-bgPanel rounded transition-colors">
                        <Settings size={14} />
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
                        onClick={() => setActiveView('terminal')} 
                        tooltip="Terminal"
                    />
                    <ActivityButton 
                        icon={<LayoutGrid size={19} strokeWidth={1.7} />} 
                        active={activeView === 'manager'} 
                        onClick={() => setActiveView('manager')} 
                        tooltip="Server Manager"
                    />
                    <ActivityButton 
                        icon={<FolderGit2 size={19} strokeWidth={1.7} />} 
                        active={activeView === 'git'} 
                        onClick={() => setActiveView('git')} 
                        tooltip="GitOps"
                    />
                    <ActivityButton 
                        icon={<Play size={19} strokeWidth={1.7} />} 
                        active={activeView === 'automation'} 
                        onClick={() => setActiveView('automation')} 
                        tooltip="Tunnels & Forwarding"
                    />
                    <div className="mt-auto mb-1">
                        <ActivityButton 
                            icon={<Settings size={19} strokeWidth={1.7} />} 
                            active={false} 
                            onClick={() => {}} 
                            tooltip="Settings"
                        />
                    </div>
                </div>

                {/* Primary Sidebar (Hosts tree explorer for quick access) */}
                {activeView === 'terminal' && (
                    <div className="w-60 bg-bgSidebar border-r border-borderDark flex flex-col shrink-0 overflow-hidden">
                        <div className="h-9 px-3.5 flex items-center justify-between text-xs font-semibold text-textMuted uppercase tracking-wider border-b border-borderDark/40">
                            <span>Endpoints</span>
                            <button onClick={handleNewHost} className="p-1 hover:text-textMain rounded hover:bg-bgPanel text-textFaint transition-colors" title="New Endpoint">
                                <Plus size={14} />
                            </button>
                        </div>

                        <div className="flex-1 overflow-y-auto py-1">
                            <div 
                                className="px-3 py-1.5 flex items-center gap-1.5 text-xs font-semibold text-textFaint uppercase tracking-wider cursor-pointer hover:text-textMain select-none"
                                onClick={() => setHostsExpanded(!hostsExpanded)}
                            >
                                {hostsExpanded ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
                                <span>Configured Hosts</span>
                            </div>

                            {hostsExpanded && (
                                <div className="flex flex-col mt-0.5">
                                    <TreeItem 
                                        name="Local Shell" 
                                        active={activeTab === 'local'} 
                                        onClick={handleSelectLocal}
                                        icon={<Terminal size={14} className="text-textMuted" />}
                                        badge="native"
                                    />
                                    {hosts && hosts.map((host: any, idx: number) => (
                                        <TreeItem 
                                            key={idx}
                                            name={host.Name} 
                                            active={activeTab === host.Name} 
                                            onClick={() => handleConnectHost(host)}
                                            status={host.Health === 'online' || host.Health === 1 ? 'online' : 'offline'}
                                            subtext={host.Hostname}
                                            onEdit={() => handleEditHost(host)}
                                            onDelete={(e) => handleDeleteHost(e, host.ID)}
                                        />
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                )}

                {/* Main Workspace Area */}
                {activeView === 'terminal' && (
                    <div className="flex-1 bg-bgMain flex flex-col overflow-hidden min-w-0">
                        {/* Tab Header Bar */}
                        <div className="h-9 bg-bgSidebar border-b border-borderDark flex items-center overflow-x-auto no-scrollbar shrink-0">
                            <div className="h-full px-4 flex items-center gap-2 bg-bgMain border-r border-borderDark border-t-2 border-t-textMain text-xs font-medium text-textMain cursor-pointer">
                                <Terminal size={13} className="text-textMuted" />
                                <span>{activeTab}</span>
                                <button className="p-0.5 hover:bg-bgPanel rounded text-textFaint hover:text-textMain ml-1">
                                    <X size={12} />
                                </button>
                            </div>
                        </div>

                        {/* Terminal Live Canvas */}
                        <div className="flex-1 overflow-hidden bg-bgMain">
                            <TerminalComponent 
                                key={activeTab} 
                                sessionType={activeTab === 'local' ? 'local' : 'ssh'} 
                                hostID={activeHostID} 
                            />
                        </div>

                        {/* Status Bar */}
                        <div className="h-6 bg-bgSidebar border-t border-borderDark flex items-center px-3.5 text-[11px] text-textMuted justify-between shrink-0 select-none">
                            <div className="flex items-center gap-4">
                                <span className="flex items-center gap-1.5 hover:text-textMain cursor-pointer font-medium">
                                    <span className="w-2 h-2 rounded-full bg-emerald-400"></span> {activeTab}
                                </span>
                                <span className="hover:text-textMain cursor-pointer">UTF-8</span>
                                <span className="hover:text-textMain cursor-pointer">Interactive PTY</span>
                            </div>
                            <div className="flex items-center gap-3">
                                <span className="hover:text-textMain cursor-pointer">xterm-256color</span>
                                <span className="hover:text-textMain cursor-pointer">AES-256-GCM</span>
                            </div>
                        </div>
                    </div>
                )}

                {/* Full Server Manager Dashboard */}
                {activeView === 'manager' && (
                    <HostManagerView 
                        hosts={hosts} 
                        onConnect={handleConnectHost} 
                        onEdit={handleEditHost} 
                        onDelete={(id) => DeleteHost(id).then(loadHosts)}
                        onNew={handleNewHost}
                    />
                )}

                {activeView === 'git' && (
                    <div className="flex-1 bg-bgMain flex flex-col items-center justify-center p-8 text-center text-textMuted select-none">
                        <FolderGit2 size={36} className="text-textFaint mb-3" />
                        <h2 className="text-sm font-semibold text-textMain">GitOps Vault Sync</h2>
                        <p className="text-xs text-textFaint max-w-sm mt-1">Automatic two-way synchronization of your SSH server configs and port forwards with a private Git repository.</p>
                    </div>
                )}

                {activeView === 'automation' && (
                    <div className="flex-1 bg-bgMain flex flex-col items-center justify-center p-8 text-center text-textMuted select-none">
                        <Play size={36} className="text-textFaint mb-3" />
                        <h2 className="text-sm font-semibold text-textMain">Port Forwarding & SOCKS5 Proxy</h2>
                        <p className="text-xs text-textFaint max-w-sm mt-1">Manage dynamic SSH tunnels, local port forwards, and reverse bastion tunnels.</p>
                    </div>
                )}
            </div>

            {/* Host Edit / Create Modal */}
            <HostModal 
                isOpen={showHostModal} 
                onClose={() => setShowHostModal(false)} 
                onSave={handleSaveHost}
                initialHost={editingHost}
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

function TreeItem({ name, active, onClick, icon, status, subtext, badge, onEdit, onDelete }: { name: string, active: boolean, onClick: () => void, icon?: any, status?: 'online' | 'offline', subtext?: string, badge?: string, onEdit?: () => void, onDelete?: (e: React.MouseEvent) => void }) {
    return (
        <div 
            onClick={onClick}
            className={`group px-3 py-1.5 cursor-pointer flex items-center justify-between text-xs transition-colors ${active ? 'bg-bgPanel text-textMain font-medium' : 'text-textMuted hover:bg-bgHover hover:text-textMain'}`}
        >
            <div className="flex items-center gap-2.5 min-w-0">
                {icon ? icon : (
                    <div className={`w-2 h-2 rounded-full shrink-0 ${status === 'online' ? 'bg-emerald-400' : 'bg-zinc-600'}`} />
                )}
                <span className="truncate">{name}</span>
            </div>
            <div className="flex items-center gap-1">
                {badge && <span className="text-[10px] text-textFaint font-mono bg-bgPanel px-1.5 py-0.5 rounded border border-borderDark">{badge}</span>}
                {subtext && <span className="text-xs text-textFaint truncate ml-2 group-hover:hidden">{subtext}</span>}
                {onEdit && (
                    <button 
                        onClick={(e) => { e.stopPropagation(); onEdit(); }} 
                        className="hidden group-hover:block p-0.5 text-textFaint hover:text-textMain rounded transition-colors"
                        title="Edit Host"
                    >
                        <Edit3 size={12} />
                    </button>
                )}
                {onDelete && (
                    <button 
                        onClick={onDelete} 
                        className="hidden group-hover:block p-0.5 text-textFaint hover:text-rose-400 rounded transition-colors"
                        title="Delete Host"
                    >
                        <Trash2 size={12} />
                    </button>
                )}
            </div>
        </div>
    );
}

export default App;
