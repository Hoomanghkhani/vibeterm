import React, { useState, useEffect, useRef } from 'react';
import { 
    Terminal, 
    Server, 
    Folder, 
    Network, 
    Code2, 
    Radio, 
    FolderGit2, 
    Settings, 
    Split, 
    Plus, 
    RefreshCw, 
    X, 
    Command,
    Search,
    ChevronRight,
    FoldHorizontal,
    Maximize2
} from 'lucide-react';

export interface PaletteCommand {
    id: string;
    title: string;
    category: 'Terminal' | 'Navigation' | 'Endpoints' | 'Tools';
    shortcut?: string;
    icon: React.ReactNode;
    action: () => void;
}

interface CommandPaletteProps {
    isOpen: boolean;
    onClose: () => void;
    hosts: any[];
    onConnectHost: (host: any) => void;
    onNewLocalTab: () => void;
    onSplitHorizontal: () => void;
    onSplitVertical: () => void;
    onCloseActiveTab: () => void;
    onNavigate: (view: string) => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({
    isOpen,
    onClose,
    hosts = [],
    onConnectHost,
    onNewLocalTab,
    onSplitHorizontal,
    onSplitVertical,
    onCloseActiveTab,
    onNavigate,
}) => {
    const [query, setQuery] = useState('');
    const [selectedIndex, setSelectedIndex] = useState(0);
    const inputRef = useRef<HTMLInputElement>(null);

    // Build complete command list
    const baseCommands: PaletteCommand[] = [
        // Terminal commands
        {
            id: 'term-new-local',
            title: 'Terminal: New Local Shell',
            category: 'Terminal',
            shortcut: 'Ctrl+T',
            icon: <Terminal size={14} className="text-zinc-400" />,
            action: () => { onNewLocalTab(); onClose(); }
        },
        {
            id: 'term-split-vert',
            title: 'Terminal: Split Pane Vertically (Right)',
            category: 'Terminal',
            shortcut: 'Ctrl+Shift+O',
            icon: <Split size={14} className="text-zinc-400" />,
            action: () => { onSplitVertical(); onClose(); }
        },
        {
            id: 'term-split-horiz',
            title: 'Terminal: Split Pane Horizontally (Down)',
            category: 'Terminal',
            shortcut: 'Ctrl+Shift+E',
            icon: <Split size={14} className="rotate-90 text-zinc-400" />,
            action: () => { onSplitHorizontal(); onClose(); }
        },
        {
            id: 'term-close-tab',
            title: 'Terminal: Close Active Tab',
            category: 'Terminal',
            shortcut: 'Ctrl+W',
            icon: <X size={14} className="text-rose-400" />,
            action: () => { onCloseActiveTab(); onClose(); }
        },
        // Navigation commands
        {
            id: 'nav-term',
            title: 'View: Open Interactive Terminal Workspace',
            category: 'Navigation',
            shortcut: 'Ctrl+1',
            icon: <Terminal size={14} className="text-zinc-400" />,
            action: () => { onNavigate('terminal'); onClose(); }
        },
        {
            id: 'nav-manager',
            title: 'View: Open Endpoints Manager',
            category: 'Navigation',
            shortcut: 'Ctrl+2',
            icon: <Server size={14} className="text-zinc-400" />,
            action: () => { onNavigate('manager'); onClose(); }
        },
        {
            id: 'nav-files',
            title: 'View: Open SFTP File Explorer',
            category: 'Navigation',
            shortcut: 'Ctrl+3',
            icon: <Folder size={14} className="text-zinc-400" />,
            action: () => { onNavigate('files'); onClose(); }
        },
        {
            id: 'nav-tunnels',
            title: 'View: Open Port Forwarding & Tunnels',
            category: 'Navigation',
            shortcut: 'Ctrl+4',
            icon: <Network size={14} className="text-zinc-400" />,
            action: () => { onNavigate('tunnels'); onClose(); }
        },
        {
            id: 'nav-snippets',
            title: 'View: Open Automation Snippets',
            category: 'Navigation',
            shortcut: 'Ctrl+5',
            icon: <Code2 size={14} className="text-zinc-400" />,
            action: () => { onNavigate('snippets'); onClose(); }
        },
        {
            id: 'nav-scanner',
            title: 'View: Open Subnet & Port Scanner',
            category: 'Navigation',
            shortcut: 'Ctrl+6',
            icon: <Radio size={14} className="text-zinc-400" />,
            action: () => { onNavigate('scanner'); onClose(); }
        },
        {
            id: 'nav-git',
            title: 'View: Open GitOps Vault Sync',
            category: 'Navigation',
            shortcut: 'Ctrl+7',
            icon: <FolderGit2 size={14} className="text-zinc-400" />,
            action: () => { onNavigate('git'); onClose(); }
        },
        {
            id: 'nav-settings',
            title: 'View: Open Settings & Security',
            category: 'Navigation',
            shortcut: 'Ctrl+,',
            icon: <Settings size={14} className="text-zinc-400" />,
            action: () => { onNavigate('settings'); onClose(); }
        },
    ];

    // Add dynamic host connect commands
    const hostCommands: PaletteCommand[] = hosts.map((h) => {
        const name = h.name || h.Name || 'Server';
        const ip = h.hostname || h.Hostname || '';
        return {
            id: `connect-host-${h.id || h.ID}`,
            title: `Connect SSH: ${name} (${ip})`,
            category: 'Endpoints',
            icon: <Server size={14} className="text-emerald-400" />,
            action: () => { onConnectHost(h); onClose(); }
        };
    });

    const allCommands = [...baseCommands, ...hostCommands];

    // Filter commands by query
    const filtered = allCommands.filter((cmd) => {
        const lowerQ = query.toLowerCase();
        return (
            cmd.title.toLowerCase().includes(lowerQ) ||
            cmd.category.toLowerCase().includes(lowerQ) ||
            (cmd.shortcut && cmd.shortcut.toLowerCase().includes(lowerQ))
        );
    });

    useEffect(() => {
        if (isOpen) {
            setQuery('');
            setSelectedIndex(0);
            setTimeout(() => inputRef.current?.focus(), 50);
        }
    }, [isOpen]);

    useEffect(() => {
        setSelectedIndex(0);
    }, [query]);

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Escape') {
            onClose();
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            setSelectedIndex((prev) => (prev < filtered.length - 1 ? prev + 1 : 0));
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            setSelectedIndex((prev) => (prev > 0 ? prev - 1 : filtered.length - 1));
        } else if (e.key === 'Enter') {
            e.preventDefault();
            if (filtered[selectedIndex]) {
                filtered[selectedIndex].action();
            }
        }
    };

    if (!isOpen) return null;

    return (
        <div 
            className="fixed inset-0 z-50 flex items-start justify-center pt-20 bg-black/60 backdrop-blur-[2px] select-none"
            onClick={onClose}
        >
            <div 
                className="w-full max-w-xl bg-bgCard border border-borderActive/60 rounded-xl shadow-2xl overflow-hidden flex flex-col font-sans animate-in fade-in zoom-in-95 duration-100"
                onClick={(e) => e.stopPropagation()}
            >
                {/* Search Bar Input */}
                <div className="h-11 px-3.5 border-b border-borderDark flex items-center gap-2.5 bg-bgPanel/50">
                    <Search size={15} className="text-textFaint shrink-0" />
                    <input 
                        ref={inputRef}
                        type="text"
                        placeholder="Type a command or search servers (e.g. 'SSH', 'Split', 'Tunnels')..."
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        onKeyDown={handleKeyDown}
                        className="flex-1 bg-transparent text-sm text-textMain placeholder:text-textFaint outline-none font-sans"
                    />
                    <div className="flex items-center gap-1 text-[10px] text-textFaint font-mono bg-bgMain px-1.5 py-0.5 rounded border border-borderDark">
                        <span>ESC</span>
                    </div>
                </div>

                {/* Command List Results */}
                <div className="max-h-80 overflow-y-auto py-1">
                    {filtered.length === 0 ? (
                        <div className="px-4 py-8 text-center text-xs text-textFaint">
                            No matching commands or hosts found for "{query}"
                        </div>
                    ) : (
                        filtered.map((cmd, idx) => {
                            const isSelected = idx === selectedIndex;
                            return (
                                <div
                                    key={cmd.id}
                                    onClick={() => cmd.action()}
                                    onMouseEnter={() => setSelectedIndex(idx)}
                                    className={`px-3.5 py-2 flex items-center justify-between text-xs cursor-pointer transition-colors ${
                                        isSelected 
                                            ? 'bg-bgPanel text-textMain font-medium border-l-2 border-l-textMain' 
                                            : 'text-textMuted hover:bg-bgHover hover:text-textMain'
                                    }`}
                                >
                                    <div className="flex items-center gap-2.5 min-w-0">
                                        {cmd.icon}
                                        <span className="truncate">{cmd.title}</span>
                                    </div>
                                    <div className="flex items-center gap-2 shrink-0">
                                        <span className="text-[10px] text-textFaint font-mono uppercase tracking-wider">{cmd.category}</span>
                                        {cmd.shortcut && (
                                            <span className="text-[10px] text-textFaint font-mono bg-bgMain px-1.5 py-0.5 rounded border border-borderDark">
                                                {cmd.shortcut}
                                            </span>
                                        )}
                                    </div>
                                </div>
                            );
                        })
                    )}
                </div>

                {/* Footer Hint Bar */}
                <div className="h-7 bg-bgMain border-t border-borderDark px-3 flex items-center justify-between text-[11px] text-textFaint font-mono">
                    <div className="flex items-center gap-3">
                        <span><kbd className="bg-bgPanel px-1 rounded border border-borderDark">↑↓</kbd> to navigate</span>
                        <span><kbd className="bg-bgPanel px-1 rounded border border-borderDark">↵</kbd> to select</span>
                    </div>
                    <span>{filtered.length} commands available</span>
                </div>
            </div>
        </div>
    );
};
