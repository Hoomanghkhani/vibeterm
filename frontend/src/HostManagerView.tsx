import React, { useState } from 'react';
import { Server, Plus, Search, Terminal, Edit3, Trash2, Shield, ArrowUpRight } from 'lucide-react';

interface HostManagerViewProps {
    hosts: any[];
    onConnect: (host: any) => void;
    onEdit: (host: any) => void;
    onDelete: (id: string) => void;
    onNew: () => void;
}

export const HostManagerView: React.FC<HostManagerViewProps> = ({ hosts = [], onConnect, onEdit, onDelete, onNew }) => {
    const [search, setSearch] = useState('');
    const [selectedEnv, setSelectedEnv] = useState<string>('all');

    const getHostName = (h: any) => h?.name || h?.Name || 'Unnamed Server';
    const getHostIp = (h: any) => h?.hostname || h?.Hostname || '127.0.0.1';
    const getHostUser = (h: any) => h?.username || h?.Username || 'root';
    const getHostPort = (h: any) => h?.port || h?.Port || 22;
    const getHostEnv = (h: any) => h?.environment || h?.Environment || 'production';
    const getHostTags = (h: any) => h?.tags || h?.Tags || [];
    const getHostNotes = (h: any) => h?.notes || h?.Notes || '';
    const getHostId = (h: any) => h?.id || h?.ID || '';

    const filteredHosts = (hosts || []).filter((h) => {
        if (!h) return false;
        const name = getHostName(h).toLowerCase();
        const hostname = getHostIp(h).toLowerCase();
        const username = getHostUser(h).toLowerCase();
        const env = getHostEnv(h);
        const q = search.toLowerCase();

        const matchesSearch = name.includes(q) || hostname.includes(q) || username.includes(q);
        const matchesEnv = selectedEnv === 'all' || env === selectedEnv;
        return matchesSearch && matchesEnv;
    });

    return (
        <div className="flex-1 flex flex-col bg-bgMain overflow-hidden select-none">
            {/* Action Bar */}
            <div className="h-12 px-6 border-b border-borderDark flex items-center justify-between shrink-0 bg-bgCard">
                <div className="flex items-center gap-3">
                    <div className="w-72 h-8 bg-bgMain border border-borderDark rounded-lg flex items-center px-3 text-xs text-textMuted gap-2">
                        <Search size={13} className="text-textFaint" />
                        <input 
                            type="text" 
                            placeholder="Filter servers by name, IP, user..." 
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="bg-transparent border-none outline-none w-full text-textMain placeholder:text-textFaint"
                        />
                    </div>

                    {/* Environment Pill Filter */}
                    <div className="flex items-center gap-1 bg-bgMain p-1 rounded-lg border border-borderDark text-xs">
                        {['all', 'production', 'staging', 'dev'].map((env) => (
                            <button
                                key={env}
                                onClick={() => setSelectedEnv(env)}
                                className={`px-2.5 py-1 rounded-md capitalize font-medium transition-colors ${selectedEnv === env ? 'bg-bgPanel text-textMain shadow-sm' : 'text-textFaint hover:text-textMuted'}`}
                            >
                                {env}
                            </button>
                        ))}
                    </div>
                </div>

                <button 
                    onClick={onNew}
                    className="px-3.5 py-1.5 rounded-lg bg-textMain text-bgMain font-medium text-xs flex items-center gap-1.5 hover:opacity-90 transition-opacity"
                >
                    <Plus size={14} strokeWidth={2.5} />
                    <span>New Endpoint</span>
                </button>
            </div>

            {/* Server Grid */}
            <div className="flex-1 overflow-y-auto p-6">
                {filteredHosts.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-center p-8 text-textMuted">
                        <Server size={36} className="text-textFaint mb-3 opacity-60" />
                        <h3 className="text-sm font-semibold text-textMain">No Endpoints Found</h3>
                        <p className="text-xs text-textFaint max-w-sm mt-1">No servers match your current search or environment filter.</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {filteredHosts.map((host, idx) => {
                            const id = getHostId(host) || `host-${idx}`;
                            const name = getHostName(host);
                            const ip = getHostIp(host);
                            const user = getHostUser(host);
                            const port = getHostPort(host);
                            const env = getHostEnv(host);
                            const tags = getHostTags(host);
                            const notes = getHostNotes(host);

                            return (
                                <div 
                                    key={id}
                                    className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col justify-between hover:border-borderActive hover:bg-bgPanel/50 transition-all duration-200 shadow-card"
                                >
                                    <div>
                                        <div className="flex items-start justify-between mb-2">
                                            <div className="flex items-center gap-2.5 min-w-0">
                                                <div className="w-8 h-8 rounded-lg bg-bgMain border border-borderDark flex items-center justify-center text-textMain shrink-0">
                                                    <Server size={16} strokeWidth={1.5} />
                                                </div>
                                                <div className="min-w-0">
                                                    <h3 className="text-xs font-semibold text-textMain truncate">{name}</h3>
                                                    <span className="text-[11px] text-textFaint font-mono block truncate">{user}@{ip}:{port}</span>
                                                </div>
                                            </div>

                                            <span className={`text-[10px] px-2 py-0.5 rounded font-mono uppercase font-semibold tracking-wider ${
                                                env === 'production' ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20' :
                                                env === 'staging' ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20' :
                                                'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                                            }`}>
                                                {env}
                                            </span>
                                        </div>

                                        {Array.isArray(tags) && tags.length > 0 && (
                                            <div className="flex items-center gap-1.5 flex-wrap my-3">
                                                {tags.map((tag: string, i: number) => (
                                                    <span key={i} className="text-[10px] px-2 py-0.5 rounded bg-bgMain border border-borderDark text-textMuted font-mono">
                                                        #{tag}
                                                    </span>
                                                ))}
                                            </div>
                                        )}

                                        {notes && (
                                            <p className="text-[11px] text-textFaint line-clamp-2 my-2 italic">
                                                "{notes}"
                                            </p>
                                        )}
                                    </div>

                                    {/* Card Footer Actions */}
                                    <div className="pt-3 border-t border-borderDark flex items-center justify-between mt-2">
                                        <div className="flex items-center gap-1.5 text-[11px] text-textFaint">
                                            <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                                            <span>SSH Protocol</span>
                                        </div>

                                        <div className="flex items-center gap-1">
                                            <button 
                                                onClick={() => onEdit(host)} 
                                                className="p-1.5 rounded hover:bg-bgMain text-textFaint hover:text-textMain transition-colors"
                                                title="Edit Server"
                                            >
                                                <Edit3 size={13} />
                                            </button>
                                            <button 
                                                onClick={() => onDelete(id)} 
                                                className="p-1.5 rounded hover:bg-bgMain text-textFaint hover:text-rose-400 transition-colors"
                                                title="Delete Server"
                                            >
                                                <Trash2 size={13} />
                                            </button>
                                            <button 
                                                onClick={() => onConnect(host)}
                                                className="px-2.5 py-1 rounded bg-bgMain border border-borderDark hover:border-borderActive text-textMain text-xs font-medium flex items-center gap-1 ml-1 transition-colors"
                                            >
                                                <Terminal size={12} />
                                                <span>Connect</span>
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
};
