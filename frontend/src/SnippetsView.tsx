import React, { useState, useEffect } from 'react';
import { Play, Plus, Search, Copy, Check, Trash2, Edit3, X, Terminal, Code2, Tag } from 'lucide-react';
import { GetSnippets, SaveSnippet, DeleteSnippet, SendTerminalInput } from '../wailsjs/go/main/App';

interface SnippetsViewProps {
    onRunInTerminal?: (command: string) => void;
}

export const SnippetsView: React.FC<SnippetsViewProps> = ({ onRunInTerminal }) => {
    const [snippets, setSnippets] = useState<any[]>([]);
    const [search, setSearch] = useState('');
    const [showModal, setShowModal] = useState(false);
    const [editingSnippet, setEditingSnippet] = useState<any>(null);
    const [copiedId, setCopiedId] = useState<string | null>(null);

    const [form, setForm] = useState({
        id: '',
        title: '',
        description: '',
        command: '',
        tags: '',
    });

    const loadSnippets = () => {
        GetSnippets().then((data) => setSnippets(data || [])).catch(console.error);
    };

    useEffect(() => {
        loadSnippets();
    }, []);

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault();
        const tagsArray = form.tags.split(',').map((t) => t.trim()).filter(Boolean);
        await SaveSnippet({
            id: form.id || `snippet-${Date.now()}`,
            title: form.title,
            description: form.description,
            command: form.command,
            tags: tagsArray,
            variables: {},
        } as any);
        setShowModal(false);
        loadSnippets();
    };

    const handleDelete = async (id: string) => {
        await DeleteSnippet(id);
        loadSnippets();
    };

    const handleCopy = (id: string, cmd: string) => {
        navigator.clipboard.writeText(cmd);
        setCopiedId(id);
        setTimeout(() => setCopiedId(null), 2000);
    };

    const handleEdit = (s: any) => {
        setForm({
            id: s.id,
            title: s.title,
            description: s.description,
            command: s.command,
            tags: (s.tags || []).join(', '),
        });
        setEditingSnippet(s);
        setShowModal(true);
    };

    const handleNew = () => {
        setForm({
            id: '',
            title: '',
            description: '',
            command: '',
            tags: 'docker, k8s',
        });
        setEditingSnippet(null);
        setShowModal(true);
    };

    const filteredSnippets = snippets.filter((s) => 
        s.title.toLowerCase().includes(search.toLowerCase()) ||
        s.description.toLowerCase().includes(search.toLowerCase()) ||
        s.command.toLowerCase().includes(search.toLowerCase())
    );

    return (
        <div className="flex-1 flex flex-col bg-bgMain overflow-hidden select-none">
            {/* Action Bar */}
            <div className="h-12 px-6 border-b border-borderDark flex items-center justify-between shrink-0 bg-bgCard">
                <div className="flex items-center gap-3">
                    <div className="w-80 h-8 bg-bgMain border border-borderDark rounded-lg flex items-center px-3 text-xs text-textMuted gap-2">
                        <Search size={13} className="text-textFaint" />
                        <input 
                            type="text" 
                            placeholder="Search scripts, runbooks, docker commands..." 
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="bg-transparent border-none outline-none w-full text-textMain placeholder:text-textFaint"
                        />
                    </div>
                </div>

                <button 
                    onClick={handleNew}
                    className="px-3.5 py-1.5 rounded-lg bg-textMain text-bgMain font-medium text-xs flex items-center gap-1.5 hover:opacity-90 transition-opacity"
                >
                    <Plus size={14} strokeWidth={2.5} />
                    <span>New Snippet</span>
                </button>
            </div>

            {/* Grid */}
            <div className="flex-1 overflow-y-auto p-6">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filteredSnippets.map((s) => (
                        <div key={s.id} className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col justify-between shadow-card hover:border-borderActive transition-all">
                            <div>
                                <div className="flex items-start justify-between mb-1.5">
                                    <div className="flex items-center gap-2">
                                        <Code2 size={15} className="text-textMuted" />
                                        <h3 className="text-xs font-semibold text-textMain truncate">{s.title}</h3>
                                    </div>
                                    <div className="flex items-center gap-1">
                                        <button onClick={() => handleEdit(s)} className="p-1 rounded hover:bg-bgMain text-textFaint hover:text-textMain"><Edit3 size={12} /></button>
                                        <button onClick={() => handleDelete(s.id)} className="p-1 rounded hover:bg-bgMain text-textFaint hover:text-rose-400"><Trash2 size={12} /></button>
                                    </div>
                                </div>

                                <p className="text-[11px] text-textFaint line-clamp-2 mb-3">{s.description}</p>

                                {/* Command Preview Box */}
                                <div className="bg-bgMain p-2.5 rounded-lg border border-borderDark font-mono text-[11px] text-emerald-400 overflow-x-auto select-text">
                                    {s.command}
                                </div>

                                {s.tags && s.tags.length > 0 && (
                                    <div className="flex items-center gap-1 flex-wrap mt-3">
                                        {s.tags.map((t: string, i: number) => (
                                            <span key={i} className="text-[9px] px-1.5 py-0.5 rounded bg-bgMain border border-borderDark text-textMuted font-mono">
                                                #{t}
                                            </span>
                                        ))}
                                    </div>
                                )}
                            </div>

                            <div className="pt-3 border-t border-borderDark flex items-center justify-end gap-2 mt-3">
                                <button 
                                    onClick={() => handleCopy(s.id, s.command)}
                                    className="px-2.5 py-1 rounded bg-bgMain border border-borderDark hover:border-borderActive text-textMain text-xs font-medium flex items-center gap-1 transition-colors"
                                >
                                    {copiedId === s.id ? <Check size={12} className="text-emerald-400" /> : <Copy size={12} />}
                                    <span>{copiedId === s.id ? 'Copied' : 'Copy'}</span>
                                </button>
                                {onRunInTerminal && (
                                    <button 
                                        onClick={() => onRunInTerminal(s.command)}
                                        className="px-2.5 py-1 rounded bg-textMain text-bgMain text-xs font-medium flex items-center gap-1 transition-opacity hover:opacity-90"
                                    >
                                        <Play size={11} />
                                        <span>Run</span>
                                    </button>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Modal */}
            {showModal && (
                <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4">
                    <div className="w-[480px] bg-bgCard border border-borderDark rounded-xl shadow-2xl overflow-hidden">
                        <div className="h-11 px-4 border-b border-borderDark flex items-center justify-between bg-bgPanel">
                            <span className="text-xs font-semibold text-textMain uppercase tracking-wider">{editingSnippet ? 'Edit Snippet' : 'New Automation Snippet'}</span>
                            <button onClick={() => setShowModal(false)} className="text-textFaint hover:text-textMain"><X size={15} /></button>
                        </div>
                        <form onSubmit={handleSave} className="p-5 space-y-3.5 text-xs">
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Snippet Title *</label>
                                <input 
                                    type="text" 
                                    placeholder="e.g. Docker Prune & Flush Cache"
                                    value={form.title}
                                    onChange={(e) => setForm({ ...form, title: e.target.value })}
                                    required
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                />
                            </div>
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Description</label>
                                <input 
                                    type="text" 
                                    placeholder="What this automation runbook does..."
                                    value={form.description}
                                    onChange={(e) => setForm({ ...form, description: e.target.value })}
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                />
                            </div>
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Bash / Shell Command(s) *</label>
                                <textarea 
                                    rows={4}
                                    placeholder="docker system prune -af && systemctl restart nginx"
                                    value={form.command}
                                    onChange={(e) => setForm({ ...form, command: e.target.value })}
                                    required
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono resize-none"
                                />
                            </div>
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Tags (Comma-separated)</label>
                                <input 
                                    type="text" 
                                    placeholder="docker, cleanup, deploy"
                                    value={form.tags}
                                    onChange={(e) => setForm({ ...form, tags: e.target.value })}
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                />
                            </div>
                            <div className="pt-3 border-t border-borderDark flex items-center justify-end gap-2.5">
                                <button type="button" onClick={() => setShowModal(false)} className="px-3.5 py-2 rounded-md hover:bg-bgHover text-textMuted font-medium">Cancel</button>
                                <button type="submit" className="px-4 py-2 rounded-md bg-textMain text-bgMain font-semibold hover:opacity-90">Save Snippet</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};
