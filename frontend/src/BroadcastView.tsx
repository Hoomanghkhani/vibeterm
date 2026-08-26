import React, { useState } from 'react';
import { Radio, Terminal, Send, CheckCircle2, XCircle, AlertTriangle, Play, ShieldAlert } from 'lucide-react';
import { SendTerminalInput } from '../wailsjs/go/main/App';

interface BroadcastViewProps {
    hosts: any[];
    onConnectHost: (host: any) => void;
}

export const BroadcastView: React.FC<BroadcastViewProps> = ({ hosts = [], onConnectHost }) => {
    const [selectedHostIds, setSelectedHostIds] = useState<string[]>([]);
    const [command, setCommand] = useState('');
    const [history, setHistory] = useState<{ id: string; command: string; targets: number; timestamp: string }[]>([]);
    const [isExecuting, setIsExecuting] = useState(false);
    const [showSafetyModal, setShowSafetyModal] = useState(false);

    const toggleHost = (id: string) => {
        setSelectedHostIds((prev) => 
            prev.includes(id) ? prev.filter((h) => h !== id) : [...prev, id]
        );
    };

    const selectAll = () => {
        if (selectedHostIds.length === hosts.length) {
            setSelectedHostIds([]);
        } else {
            setSelectedHostIds(hosts.map((h) => h.id || h.ID));
        }
    };

    const isDangerous = (cmd: string) => {
        const lower = cmd.toLowerCase().trim();
        return (
            lower.includes('rm -rf') ||
            lower.includes('shutdown') ||
            lower.includes('reboot') ||
            lower.includes('mkfs') ||
            lower.includes('drop database') ||
            lower.includes('kubectl delete') ||
            lower.includes('dd if=')
        );
    };

    const handleExecute = () => {
        if (!command.trim() || selectedHostIds.length === 0) return;
        if (isDangerous(command)) {
            setShowSafetyModal(true);
            return;
        }
        executeBroadcast();
    };

    const executeBroadcast = async () => {
        setShowSafetyModal(false);
        setIsExecuting(true);

        for (const hostId of selectedHostIds) {
            try {
                await SendTerminalInput(hostId, command + '\n');
            } catch (err) {
                console.error(`Broadcast failed to ${hostId}:`, err);
            }
        }

        setHistory((prev) => [
            {
                id: `hist-${Date.now()}`,
                command: command,
                targets: selectedHostIds.length,
                timestamp: new Date().toLocaleTimeString(),
            },
            ...prev.slice(0, 19),
        ]);

        setCommand('');
        setIsExecuting(false);
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain text-textMain overflow-hidden font-sans p-6">
            {/* Header */}
            <div className="flex items-center justify-between pb-4 border-b border-borderDark shrink-0">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-bgPanel rounded-lg border border-borderDark">
                        <Radio size={18} className="text-textMain" />
                    </div>
                    <div>
                        <h1 className="text-sm font-semibold tracking-wide">Multi-Execution Broadcast Terminal</h1>
                        <p className="text-xs text-textFaint">Execute commands concurrently across selected servers and fleet instances</p>
                    </div>
                </div>

                <div className="flex items-center gap-2">
                    <button 
                        onClick={selectAll}
                        className="px-3 py-1.5 bg-bgPanel hover:bg-bgHover border border-borderDark rounded-lg text-xs text-textMuted hover:text-textMain transition-colors"
                    >
                        {selectedHostIds.length === hosts.length ? 'Deselect All' : `Select All (${hosts.length})`}
                    </button>
                    <span className="text-xs font-mono text-textFaint bg-bgCard px-2.5 py-1.5 rounded-lg border border-borderDark">
                        {selectedHostIds.length} target{selectedHostIds.length !== 1 ? 's' : ''} selected
                    </span>
                </div>
            </div>

            {/* Main Content Area */}
            <div className="flex-1 grid grid-cols-3 gap-6 pt-6 overflow-hidden min-h-0">
                {/* Target Server Selection Panel */}
                <div className="col-span-1 bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col overflow-hidden">
                    <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider mb-3">Target Nodes</h2>
                    <div className="flex-1 overflow-y-auto space-y-1.5 pr-1">
                        {hosts.length === 0 ? (
                            <div className="text-xs text-textFaint text-center py-8">No configured hosts found.</div>
                        ) : (
                            hosts.map((h) => {
                                const id = h.id || h.ID;
                                const isSelected = selectedHostIds.includes(id);
                                return (
                                    <div
                                        key={id}
                                        onClick={() => toggleHost(id)}
                                        className={`p-2.5 rounded-lg border cursor-pointer flex items-center justify-between text-xs transition-colors ${
                                            isSelected 
                                                ? 'bg-bgPanel border-borderActive text-textMain' 
                                                : 'bg-bgMain border-borderDark text-textMuted hover:bg-bgPanel/50'
                                        }`}
                                    >
                                        <div className="flex items-center gap-2.5 min-w-0">
                                            <input 
                                                type="checkbox" 
                                                checked={isSelected} 
                                                onChange={() => {}} 
                                                className="accent-white rounded cursor-pointer"
                                            />
                                            <div className="truncate">
                                                <div className="font-medium truncate">{h.name || h.Name}</div>
                                                <div className="text-[10px] text-textFaint font-mono">{h.hostname || h.Hostname}</div>
                                            </div>
                                        </div>
                                        <span className="text-[10px] text-textFaint font-mono uppercase bg-bgCard px-1.5 py-0.5 rounded border border-borderDark">
                                            {h.environment || 'prod'}
                                        </span>
                                    </div>
                                );
                            })
                        )}
                    </div>
                </div>

                {/* Command Input & Execution History Panel */}
                <div className="col-span-2 flex flex-col gap-4 overflow-hidden">
                    {/* Command Bar */}
                    <div className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col gap-3 shrink-0">
                        <div className="flex items-center justify-between text-xs text-textFaint">
                            <span className="font-semibold uppercase tracking-wider">Broadcast Command</span>
                            <span className="font-mono">Target: {selectedHostIds.length} hosts</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="flex-1 flex items-center gap-2 bg-bgMain border border-borderDark focus-within:border-borderActive rounded-lg px-3 py-2">
                                <Terminal size={14} className="text-textFaint shrink-0" />
                                <input 
                                    type="text"
                                    placeholder="Enter command to broadcast (e.g. 'uptime', 'docker ps', 'systemctl status')..."
                                    value={command}
                                    onChange={(e) => setCommand(e.target.value)}
                                    onKeyDown={(e) => { if (e.key === 'Enter') handleExecute(); }}
                                    className="flex-1 bg-transparent text-xs text-textMain placeholder:text-textFaint outline-none font-mono"
                                />
                            </div>
                            <button
                                onClick={handleExecute}
                                disabled={isExecuting || !command.trim() || selectedHostIds.length === 0}
                                className="px-4 py-2 bg-white text-black hover:bg-zinc-200 disabled:opacity-40 disabled:hover:bg-white rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-colors shadow-sm"
                            >
                                <Send size={13} />
                                <span>Broadcast</span>
                            </button>
                        </div>
                    </div>

                    {/* Execution Audit History */}
                    <div className="flex-1 bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col overflow-hidden">
                        <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider mb-3">Broadcast Audit Log</h2>
                        <div className="flex-1 overflow-y-auto space-y-2 pr-1">
                            {history.length === 0 ? (
                                <div className="text-xs text-textFaint text-center py-12">
                                    No broadcast commands executed yet. Select targets and send a command.
                                </div>
                            ) : (
                                history.map((item) => (
                                    <div key={item.id} className="p-3 bg-bgMain border border-borderDark rounded-lg flex items-center justify-between text-xs">
                                        <div className="flex items-center gap-2.5 font-mono">
                                            <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />
                                            <span className="text-textMain font-semibold">$ {item.command}</span>
                                        </div>
                                        <div className="flex items-center gap-3 text-textFaint font-mono text-[11px]">
                                            <span>{item.targets} nodes</span>
                                            <span>{item.timestamp}</span>
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Safety Confirmation Modal */}
            {showSafetyModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
                    <div className="w-full max-w-md bg-bgCard border border-rose-500/50 rounded-xl p-5 shadow-2xl flex flex-col gap-4">
                        <div className="flex items-center gap-3 text-rose-400">
                            <ShieldAlert size={22} />
                            <h3 className="text-sm font-semibold text-white">Potentially Destructive Command Detected</h3>
                        </div>
                        <p className="text-xs text-textMuted leading-relaxed">
                            You are about to broadcast the following command across <span className="text-white font-mono font-bold">{selectedHostIds.length}</span> servers:
                        </p>
                        <div className="p-2.5 bg-bgMain border border-borderDark rounded-lg font-mono text-xs text-rose-300">
                            {command}
                        </div>
                        <div className="flex items-center justify-end gap-2 pt-2">
                            <button 
                                onClick={() => setShowSafetyModal(false)}
                                className="px-3 py-1.5 rounded-lg border border-borderDark text-xs text-textMuted hover:text-textMain hover:bg-bgPanel transition-colors"
                            >
                                Cancel
                            </button>
                            <button 
                                onClick={executeBroadcast}
                                className="px-3 py-1.5 bg-rose-600 hover:bg-rose-500 text-white rounded-lg text-xs font-semibold transition-colors"
                            >
                                Confirm & Execute
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
