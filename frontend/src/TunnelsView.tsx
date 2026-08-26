import React, { useState, useEffect } from 'react';
import { Play, Square, Plus, Radio, ArrowRight, ArrowLeft, RefreshCw, X, Shield, Globe, HardDrive } from 'lucide-react';
import { GetActiveTunnels, StartPortForward, StopPortForward } from '../wailsjs/go/main/App';

interface TunnelsViewProps {
    hosts: any[];
}

export const TunnelsView: React.FC<TunnelsViewProps> = ({ hosts }) => {
    const [tunnels, setTunnels] = useState<any[]>([]);
    const [showModal, setShowModal] = useState(false);
    const [loading, setLoading] = useState(false);
    const [selectedHostId, setSelectedHostId] = useState(hosts[0]?.ID || '');

    const [form, setForm] = useState({
        name: '',
        type: 'local',
        bindAddress: '127.0.0.1',
        bindPort: 8080,
        targetAddress: '127.0.0.1',
        targetPort: 80,
    });

    const refreshTunnels = () => {
        GetActiveTunnels().then((data) => setTunnels(data || [])).catch(console.error);
    };

    useEffect(() => {
        refreshTunnels();
        const timer = setInterval(refreshTunnels, 2000);
        return () => clearInterval(timer);
    }, []);

    const handleCreateTunnel = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!selectedHostId) return;
        setLoading(true);
        try {
            await StartPortForward(selectedHostId, {
                id: `tunnel-${Date.now()}`,
                hostId: selectedHostId,
                name: form.name || `${form.type.toUpperCase()}:${form.bindPort}`,
                type: form.type,
                bindAddress: form.bindAddress,
                bindPort: Number(form.bindPort),
                targetAddress: form.targetAddress,
                targetPort: Number(form.targetPort),
                autoStart: true,
                active: true,
                rxBytes: 0,
                txBytes: 0,
                activeConns: 0,
            } as any);
            setShowModal(false);
            refreshTunnels();
        } catch (err: any) {
            alert(`Error creating tunnel: ${err}`);
        } finally {
            setLoading(false);
        }
    };

    const handleStop = async (id: string) => {
        await StopPortForward(id);
        refreshTunnels();
    };

    const formatBytes = (bytes: number) => {
        if (!bytes) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain overflow-hidden select-none">
            {/* Header Toolbar */}
            <div className="h-12 px-6 border-b border-borderDark flex items-center justify-between shrink-0 bg-bgCard">
                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></div>
                        <span className="text-xs font-semibold text-textMain uppercase tracking-wider">Port Forwarding & SOCKS5 Tunnels</span>
                    </div>
                    <span className="text-[11px] text-textFaint font-mono">({tunnels.length} active)</span>
                </div>

                <div className="flex items-center gap-2">
                    <button 
                        onClick={refreshTunnels}
                        className="p-1.5 rounded-lg bg-bgMain border border-borderDark hover:text-textMain text-textFaint transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw size={13} />
                    </button>
                    <button 
                        onClick={() => setShowModal(true)}
                        className="px-3.5 py-1.5 rounded-lg bg-textMain text-bgMain font-medium text-xs flex items-center gap-1.5 hover:opacity-90 transition-opacity"
                    >
                        <Plus size={14} strokeWidth={2.5} />
                        <span>New Tunnel</span>
                    </button>
                </div>
            </div>

            {/* Tunnel List */}
            <div className="flex-1 overflow-y-auto p-6">
                {tunnels.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-center p-8 text-textMuted">
                        <Radio size={36} className="text-textFaint mb-3 opacity-60" />
                        <h3 className="text-sm font-semibold text-textMain">No Active Tunnels</h3>
                        <p className="text-xs text-textFaint max-w-sm mt-1">Create a local (-L), remote (-R), or dynamic SOCKS5 (-D) tunnel over encrypted SSH.</p>
                        <button 
                            onClick={() => setShowModal(true)}
                            className="mt-4 px-4 py-2 rounded-lg bg-bgCard border border-borderDark text-textMain text-xs font-medium hover:border-borderActive transition-colors flex items-center gap-1.5"
                        >
                            <Plus size={13} />
                            <span>Create Tunnel</span>
                        </button>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {tunnels.map((t) => (
                            <div key={t.id} className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col justify-between shadow-card hover:border-borderActive transition-all">
                                <div>
                                    <div className="flex items-start justify-between mb-3">
                                        <div className="flex items-center gap-2 min-w-0">
                                            <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                                            <h4 className="text-xs font-semibold text-textMain truncate">{t.name}</h4>
                                        </div>
                                        <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-bgMain border border-borderDark text-textMuted font-bold">
                                            {t.type}
                                        </span>
                                    </div>

                                    {/* Route diagram */}
                                    <div className="bg-bgMain p-2.5 rounded-lg border border-borderDark font-mono text-[11px] space-y-1.5 my-2">
                                        <div className="flex items-center justify-between text-textMuted">
                                            <span>Bind:</span>
                                            <span className="text-textMain">{t.bindAddress}:{t.bindPort}</span>
                                        </div>
                                        {t.type !== 'dynamic' && (
                                            <div className="flex items-center justify-between text-textMuted">
                                                <span>Target:</span>
                                                <span className="text-textMain">{t.targetAddress}:{t.targetPort}</span>
                                            </div>
                                        )}
                                    </div>

                                    {/* Stats */}
                                    <div className="grid grid-cols-2 gap-2 text-[10px] text-textFaint font-mono mt-3">
                                        <div>Rx: <span className="text-textMuted">{formatBytes(t.rxBytes)}</span></div>
                                        <div>Tx: <span className="text-textMuted">{formatBytes(t.txBytes)}</span></div>
                                    </div>
                                </div>

                                <div className="pt-3 border-t border-borderDark flex items-center justify-between mt-3">
                                    <span className="text-[10px] text-emerald-400 font-mono flex items-center gap-1">
                                        <Shield size={11} /> {t.activeConns || 0} active conns
                                    </span>
                                    <button 
                                        onClick={() => handleStop(t.id)}
                                        className="px-2.5 py-1 rounded bg-bgMain hover:bg-rose-500/10 text-rose-400 border border-borderDark hover:border-rose-500/30 text-xs font-medium flex items-center gap-1 transition-colors"
                                    >
                                        <Square size={11} />
                                        <span>Stop</span>
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Create Tunnel Modal */}
            {showModal && (
                <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4">
                    <div className="w-[460px] bg-bgCard border border-borderDark rounded-xl shadow-2xl overflow-hidden">
                        <div className="h-11 px-4 border-b border-borderDark flex items-center justify-between bg-bgPanel">
                            <span className="text-xs font-semibold text-textMain uppercase tracking-wider">New Forwarding Tunnel</span>
                            <button onClick={() => setShowModal(false)} className="text-textFaint hover:text-textMain"><X size={15} /></button>
                        </div>

                        <form onSubmit={handleCreateTunnel} className="p-5 space-y-3.5 text-xs">
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Target Server (SSH Host)</label>
                                <select 
                                    value={selectedHostId}
                                    onChange={(e) => setSelectedHostId(e.target.value)}
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                >
                                    {hosts.map((h) => (
                                        <option key={h.ID} value={h.ID}>{h.Name} ({h.Hostname})</option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="block text-textMuted font-medium mb-1">Tunnel Type</label>
                                <div className="grid grid-cols-3 gap-2">
                                    {[
                                        { id: 'local', label: 'Local (-L)' },
                                        { id: 'remote', label: 'Remote (-R)' },
                                        { id: 'dynamic', label: 'SOCKS5 (-D)' }
                                    ].map((t) => (
                                        <button
                                            key={t.id}
                                            type="button"
                                            onClick={() => setForm({ ...form, type: t.id })}
                                            className={`py-2 px-2 rounded-lg border font-medium text-center transition-all ${form.type === t.id ? 'bg-bgPanel border-borderActive text-textMain shadow-sm' : 'bg-bgMain border-borderDark text-textFaint hover:text-textMuted'}`}
                                        >
                                            {t.label}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div>
                                <label className="block text-textMuted font-medium mb-1">Tunnel Name / Description</label>
                                <input 
                                    type="text" 
                                    placeholder="e.g. MySQL Remote DB Proxy" 
                                    value={form.name}
                                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                />
                            </div>

                            <div className="grid grid-cols-2 gap-3">
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Local Bind Address</label>
                                    <input 
                                        type="text" 
                                        value={form.bindAddress}
                                        onChange={(e) => setForm({ ...form, bindAddress: e.target.value })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                    />
                                </div>
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Local Port</label>
                                    <input 
                                        type="number" 
                                        value={form.bindPort}
                                        onChange={(e) => setForm({ ...form, bindPort: Number(e.target.value) })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                    />
                                </div>
                            </div>

                            {form.type !== 'dynamic' && (
                                <div className="grid grid-cols-2 gap-3">
                                    <div>
                                        <label className="block text-textMuted font-medium mb-1">Remote Target Host</label>
                                        <input 
                                            type="text" 
                                            value={form.targetAddress}
                                            onChange={(e) => setForm({ ...form, targetAddress: e.target.value })}
                                            className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-textMuted font-medium mb-1">Remote Port</label>
                                        <input 
                                            type="number" 
                                            value={form.targetPort}
                                            onChange={(e) => setForm({ ...form, targetPort: Number(e.target.value) })}
                                            className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                        />
                                    </div>
                                </div>
                            )}

                            <div className="pt-3 border-t border-borderDark flex items-center justify-end gap-2.5">
                                <button 
                                    type="button" 
                                    onClick={() => setShowModal(false)}
                                    className="px-3.5 py-2 rounded-md hover:bg-bgHover text-textMuted font-medium transition-colors"
                                >
                                    Cancel
                                </button>
                                <button 
                                    type="submit" 
                                    disabled={loading}
                                    className="px-4 py-2 rounded-md bg-textMain text-bgMain font-semibold hover:opacity-90 transition-opacity flex items-center gap-1.5"
                                >
                                    <Play size={13} />
                                    <span>Start Tunnel</span>
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};
