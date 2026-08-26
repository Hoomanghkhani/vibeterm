import React, { useState, useEffect } from 'react';
import { FolderGit2, RefreshCw, Shield, Check, AlertCircle, Key, Lock, GitBranch, Terminal } from 'lucide-react';
import { GetGitOpsConfig, SaveGitOpsConfig, SyncGitOps } from '../wailsjs/go/main/App';

export const GitOpsView: React.FC = () => {
    const [config, setConfig] = useState<any>({
        repoUrl: '',
        branch: 'main',
        authType: 'ssh',
        sshKeyPath: '~/.ssh/id_ed25519',
        accessToken: '',
        autoSync: false,
        encryptSecret: true,
    });

    const [syncing, setSyncing] = useState(false);
    const [syncResult, setSyncResult] = useState<any>(null);
    const [saved, setSaved] = useState(false);

    useEffect(() => {
        GetGitOpsConfig().then((data) => {
            if (data && data.repoUrl) {
                setConfig(data);
            }
        }).catch(console.error);
    }, []);

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault();
        await SaveGitOpsConfig(config);
        setSaved(true);
        setTimeout(() => setSaved(false), 2500);
    };

    const handleSyncNow = async () => {
        setSyncing(true);
        setSyncResult(null);
        try {
            const res = await SyncGitOps();
            setSyncResult(res);
        } catch (err: any) {
            setSyncResult({ success: false, message: String(err) });
        } finally {
            setSyncing(false);
        }
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain overflow-y-auto p-8 select-none">
            <div className="max-w-2xl mx-auto w-full space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between pb-4 border-b border-borderDark">
                    <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-xl bg-bgCard border border-borderDark flex items-center justify-center text-textMain shadow-card">
                            <FolderGit2 size={20} strokeWidth={1.5} />
                        </div>
                        <div>
                            <h2 className="text-sm font-semibold text-textMain">GitOps Infrastructure Sync</h2>
                            <p className="text-xs text-textFaint">Backup, version, and sync server inventories securely via Git.</p>
                        </div>
                    </div>

                    <button
                        onClick={handleSyncNow}
                        disabled={syncing || !config.repoUrl}
                        className="px-4 py-2 rounded-lg bg-textMain text-bgMain font-medium text-xs flex items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-40"
                    >
                        <RefreshCw size={13} className={syncing ? 'animate-spin' : ''} />
                        <span>{syncing ? 'Syncing...' : 'Sync Now'}</span>
                    </button>
                </div>

                {/* Status Alert if synced */}
                {syncResult && (
                    <div className={`p-4 rounded-xl border flex items-start gap-3 text-xs leading-relaxed ${syncResult.success ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-rose-500/10 border-rose-500/30 text-rose-300'}`}>
                        {syncResult.success ? <Check size={16} className="shrink-0 mt-0.5" /> : <AlertCircle size={16} className="shrink-0 mt-0.5" />}
                        <div>
                            <span className="font-semibold block">{syncResult.message}</span>
                            {syncResult.commitHash && <span className="font-mono text-[10px] opacity-80">Commit: {syncResult.commitHash}</span>}
                        </div>
                    </div>
                )}

                {/* Config Form */}
                <form onSubmit={handleSave} className="bg-bgCard border border-borderDark rounded-xl p-6 shadow-card space-y-4 text-xs">
                    <div>
                        <label className="block text-textMuted font-medium mb-1">Git Repository URL (HTTPS or SSH) *</label>
                        <input 
                            type="text" 
                            placeholder="git@github.com:your-org/infrastructure-vault.git"
                            value={config.repoUrl}
                            onChange={(e) => setConfig({ ...config, repoUrl: e.target.value })}
                            required
                            className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-textMuted font-medium mb-1">Branch</label>
                            <div className="flex items-center gap-2 bg-bgMain border border-borderDark rounded-md px-3 py-1.5">
                                <GitBranch size={13} className="text-textFaint" />
                                <input 
                                    type="text" 
                                    value={config.branch}
                                    onChange={(e) => setConfig({ ...config, branch: e.target.value })}
                                    className="bg-transparent border-none outline-none w-full text-textMain font-mono text-xs"
                                />
                            </div>
                        </div>

                        <div>
                            <label className="block text-textMuted font-medium mb-1">Authentication Method</label>
                            <select 
                                value={config.authType}
                                onChange={(e) => setConfig({ ...config, authType: e.target.value })}
                                className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive capitalize"
                            >
                                <option value="ssh">SSH Deploy Key</option>
                                <option value="token">Personal Access Token (PAT)</option>
                            </select>
                        </div>
                    </div>

                    {config.authType === 'ssh' ? (
                        <div>
                            <label className="block text-textMuted font-medium mb-1">Deploy SSH Key Path</label>
                            <input 
                                type="text" 
                                value={config.sshKeyPath}
                                onChange={(e) => setConfig({ ...config, sshKeyPath: e.target.value })}
                                className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                            />
                        </div>
                    ) : (
                        <div>
                            <label className="block text-textMuted font-medium mb-1">GitHub / GitLab Token</label>
                            <input 
                                type="password" 
                                placeholder="ghp_xxxxxxxxxxxx"
                                value={config.accessToken}
                                onChange={(e) => setConfig({ ...config, accessToken: e.target.value })}
                                className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                            />
                        </div>
                    )}

                    {/* Zero Leak Banner */}
                    <div className="p-3.5 rounded-lg bg-bgMain border border-borderDark flex items-center gap-3">
                        <Shield size={18} className="text-emerald-400 shrink-0" />
                        <div className="text-[11px] text-textFaint leading-relaxed">
                            <span className="text-textMain font-medium">Zero-Leak Guarantee: </span>
                            All plaintext passwords and private keys are stripped or encrypted locally with AES-256 before Git commits.
                        </div>
                    </div>

                    <div className="pt-3 border-t border-borderDark flex items-center justify-end gap-3">
                        {saved && <span className="text-emerald-400 flex items-center gap-1"><Check size={13} /> Saved</span>}
                        <button 
                            type="submit" 
                            className="px-4 py-2 rounded-md bg-textMain text-bgMain font-semibold hover:opacity-90 transition-opacity"
                        >
                            Save GitOps Configuration
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
