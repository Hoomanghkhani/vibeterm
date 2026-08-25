import React, { useState, useEffect } from 'react';
import { X, Key, Shield, Folder, Tag, Terminal, Globe, Server, Check } from 'lucide-react';

interface HostModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSave: (host: any) => void;
    initialHost?: any;
}

export const HostModal: React.FC<HostModalProps> = ({ isOpen, onClose, onSave, initialHost }) => {
    const [tab, setTab] = useState<'general' | 'auth' | 'advanced'>('general');
    
    const [formData, setFormData] = useState({
        id: '',
        name: '',
        hostname: '',
        port: 22,
        protocol: 'ssh',
        username: 'root',
        authMethod: 'password',
        password: '',
        privateKeyPath: '',
        privateKeyData: '',
        keyPassphrase: '',
        environment: 'production',
        folder: 'Servers',
        tags: '',
        notes: '',
    });

    useEffect(() => {
        if (initialHost) {
            setFormData({
                id: initialHost.ID || initialHost.id || '',
                name: initialHost.Name || initialHost.name || '',
                hostname: initialHost.Hostname || initialHost.hostname || '',
                port: initialHost.Port || initialHost.port || 22,
                protocol: initialHost.Protocol || initialHost.protocol || 'ssh',
                username: initialHost.Username || initialHost.username || 'root',
                authMethod: initialHost.AuthMethod || initialHost.authMethod || 'password',
                password: initialHost.Password || initialHost.password || '',
                privateKeyPath: initialHost.PrivateKeyPath || initialHost.privateKeyPath || '',
                privateKeyData: initialHost.PrivateKeyData || initialHost.privateKeyData || '',
                keyPassphrase: initialHost.KeyPassphrase || initialHost.keyPassphrase || '',
                environment: initialHost.Environment || initialHost.environment || 'production',
                folder: initialHost.Folder || initialHost.folder || 'Servers',
                tags: Array.isArray(initialHost.Tags || initialHost.tags) ? (initialHost.Tags || initialHost.tags).join(', ') : '',
                notes: initialHost.Notes || initialHost.notes || '',
            });
        } else {
            setFormData({
                id: `host-${Date.now()}`,
                name: '',
                hostname: '',
                port: 22,
                protocol: 'ssh',
                username: 'root',
                authMethod: 'password',
                password: '',
                privateKeyPath: '',
                privateKeyData: '',
                keyPassphrase: '',
                environment: 'production',
                folder: 'Servers',
                tags: 'linux, cloud',
                notes: '',
            });
        }
    }, [initialHost, isOpen]);

    if (!isOpen) return null;

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const tagsArray = formData.tags
            .split(',')
            .map((t) => t.trim())
            .filter(Boolean);

        onSave({
            ID: formData.id || `host-${Date.now()}`,
            Name: formData.name,
            Hostname: formData.hostname,
            Port: Number(formData.port),
            Protocol: formData.protocol,
            Username: formData.username,
            AuthMethod: formData.authMethod,
            Password: formData.password,
            PrivateKeyPath: formData.privateKeyPath,
            PrivateKeyData: formData.privateKeyData,
            KeyPassphrase: formData.keyPassphrase,
            Environment: formData.environment,
            Folder: formData.folder,
            Tags: tagsArray,
            Notes: formData.notes,
            Health: 'online',
            LatencyMs: 0,
            CreatedAt: new Date().toISOString(),
            UpdatedAt: new Date().toISOString(),
        });
        onClose();
    };

    return (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4">
            <div className="w-[520px] bg-bgCard border border-borderDark rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
                {/* Modal Header */}
                <div className="h-11 px-4 border-b border-borderDark flex items-center justify-between shrink-0 bg-bgPanel">
                    <div className="flex items-center gap-2">
                        <Server size={15} className="text-textMain" />
                        <span className="text-xs font-semibold text-textMain uppercase tracking-wider">
                            {initialHost ? 'Edit Endpoint' : 'New SSH Endpoint'}
                        </span>
                    </div>
                    <button onClick={onClose} className="text-textFaint hover:text-textMain p-1 rounded transition-colors">
                        <X size={15} />
                    </button>
                </div>

                {/* Sub-tab Switcher */}
                <div className="flex border-b border-borderDark bg-bgMain px-4 text-xs">
                    <button 
                        type="button"
                        onClick={() => setTab('general')} 
                        className={`py-2 px-3 border-b-2 font-medium transition-colors ${tab === 'general' ? 'border-textMain text-textMain' : 'border-transparent text-textFaint hover:text-textMuted'}`}
                    >
                        General
                    </button>
                    <button 
                        type="button"
                        onClick={() => setTab('auth')} 
                        className={`py-2 px-3 border-b-2 font-medium transition-colors ${tab === 'auth' ? 'border-textMain text-textMain' : 'border-transparent text-textFaint hover:text-textMuted'}`}
                    >
                        Authentication & Key
                    </button>
                    <button 
                        type="button"
                        onClick={() => setTab('advanced')} 
                        className={`py-2 px-3 border-b-2 font-medium transition-colors ${tab === 'advanced' ? 'border-textMain text-textMain' : 'border-transparent text-textFaint hover:text-textMuted'}`}
                    >
                        Metadata & Grouping
                    </button>
                </div>

                {/* Form Body */}
                <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-5 space-y-4 text-xs">
                    {tab === 'general' && (
                        <div className="space-y-3.5">
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Display Label / Server Name *</label>
                                <input 
                                    type="text" 
                                    placeholder="e.g. AWS Bastion / Hetzner Node 01" 
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    required
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                />
                            </div>

                            <div className="grid grid-cols-3 gap-3">
                                <div className="col-span-2">
                                    <label className="block text-textMuted font-medium mb-1">Hostname or IP Address *</label>
                                    <input 
                                        type="text" 
                                        placeholder="192.168.1.50 or server.domain.com" 
                                        value={formData.hostname}
                                        onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                                        required
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                    />
                                </div>
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Port</label>
                                    <input 
                                        type="number" 
                                        value={formData.port}
                                        onChange={(e) => setFormData({ ...formData, port: Number(e.target.value) })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                    />
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-3">
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Default Username</label>
                                    <input 
                                        type="text" 
                                        value={formData.username}
                                        onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                    />
                                </div>
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Protocol</label>
                                    <select 
                                        value={formData.protocol}
                                        onChange={(e) => setFormData({ ...formData, protocol: e.target.value })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                    >
                                        <option value="ssh">SSH (Port 22)</option>
                                        <option value="telnet">Telnet</option>
                                        <option value="docker">Docker Exec</option>
                                        <option value="k8s">Kubernetes Pod</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    )}

                    {tab === 'auth' && (
                        <div className="space-y-3.5">
                            <div>
                                <label className="block text-textMuted font-medium mb-1">Authentication Method</label>
                                <div className="grid grid-cols-3 gap-2">
                                    {['password', 'private_key', 'ssh_agent'].map((method) => (
                                        <button
                                            key={method}
                                            type="button"
                                            onClick={() => setFormData({ ...formData, authMethod: method })}
                                            className={`py-2 px-2.5 rounded-lg border text-center font-medium capitalize transition-all ${formData.authMethod === method ? 'bg-bgPanel border-borderActive text-textMain shadow-sm' : 'bg-bgMain border-borderDark text-textFaint hover:text-textMuted'}`}
                                        >
                                            {method.replace('_', ' ')}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            {formData.authMethod === 'password' && (
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Password</label>
                                    <input 
                                        type="password" 
                                        placeholder="••••••••••••" 
                                        value={formData.password}
                                        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                    />
                                    <span className="text-[10px] text-textFaint mt-1 block">Encrypted locally with AES-256-GCM</span>
                                </div>
                            )}

                            {formData.authMethod === 'private_key' && (
                                <div className="space-y-3">
                                    <div>
                                        <label className="block text-textMuted font-medium mb-1">Private Key Path</label>
                                        <input 
                                            type="text" 
                                            placeholder="~/.ssh/id_ed25519 or /home/user/.ssh/id_rsa" 
                                            value={formData.privateKeyPath}
                                            onChange={(e) => setFormData({ ...formData, privateKeyPath: e.target.value })}
                                            className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-textMuted font-medium mb-1">Passphrase (if key is protected)</label>
                                        <input 
                                            type="password" 
                                            placeholder="Optional key passphrase" 
                                            value={formData.keyPassphrase}
                                            onChange={(e) => setFormData({ ...formData, keyPassphrase: e.target.value })}
                                            className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                        />
                                    </div>
                                </div>
                            )}

                            {formData.authMethod === 'ssh_agent' && (
                                <div className="p-3 bg-bgMain border border-borderDark rounded-lg text-textMuted text-[11px] leading-relaxed">
                                    Connects seamlessly through the system <span className="text-textMain font-mono">SSH_AUTH_SOCK</span> agent without storing any keys locally.
                                </div>
                            )}
                        </div>
                    )}

                    {tab === 'advanced' && (
                        <div className="space-y-3.5">
                            <div className="grid grid-cols-2 gap-3">
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Folder / Hierarchy</label>
                                    <input 
                                        type="text" 
                                        placeholder="e.g. AWS / Kubernetes" 
                                        value={formData.folder}
                                        onChange={(e) => setFormData({ ...formData, folder: e.target.value })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                    />
                                </div>
                                <div>
                                    <label className="block text-textMuted font-medium mb-1">Environment Label</label>
                                    <select 
                                        value={formData.environment}
                                        onChange={(e) => setFormData({ ...formData, environment: e.target.value })}
                                        className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive capitalize"
                                    >
                                        <option value="production">Production</option>
                                        <option value="staging">Staging</option>
                                        <option value="dev">Development</option>
                                        <option value="edge">Edge / IoT</option>
                                    </select>
                                </div>
                            </div>

                            <div>
                                <label className="block text-textMuted font-medium mb-1">Tags (Comma-separated)</label>
                                <input 
                                    type="text" 
                                    placeholder="nginx, database, backup" 
                                    value={formData.tags}
                                    onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive"
                                />
                            </div>

                            <div>
                                <label className="block text-textMuted font-medium mb-1">Operator Notes</label>
                                <textarea 
                                    rows={3}
                                    placeholder="Add private infrastructure notes or runbook instructions..."
                                    value={formData.notes}
                                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                                    className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive resize-none"
                                />
                            </div>
                        </div>
                    )}

                    {/* Modal Footer */}
                    <div className="pt-3 border-t border-borderDark flex items-center justify-end gap-2.5">
                        <button 
                            type="button" 
                            onClick={onClose}
                            className="px-3.5 py-2 rounded-md hover:bg-bgHover text-textMuted hover:text-textMain font-medium transition-colors"
                        >
                            Cancel
                        </button>
                        <button 
                            type="submit" 
                            className="px-4 py-2 rounded-md bg-textMain text-bgMain font-semibold hover:opacity-90 transition-opacity flex items-center gap-1.5"
                        >
                            <Check size={14} strokeWidth={2.5} />
                            <span>Save Endpoint</span>
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
