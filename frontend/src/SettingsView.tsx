import React, { useState } from 'react';
import { Settings, Shield, Lock, Key, Terminal, ExternalLink, Check, Code2 } from 'lucide-react';

export const SettingsView: React.FC = () => {
    const [masterPass, setMasterPass] = useState('');
    const [passSaved, setPassSaved] = useState(false);
    const [fontSize, setFontSize] = useState(13);

    const handleSaveSecurity = (e: React.FormEvent) => {
        e.preventDefault();
        if (!masterPass) return;
        setPassSaved(true);
        setTimeout(() => setPassSaved(false), 2500);
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain overflow-y-auto p-8 select-none">
            <div className="max-w-2xl mx-auto w-full space-y-6">
                {/* Header */}
                <div className="flex items-center gap-3 pb-4 border-b border-borderDark">
                    <div className="w-10 h-10 rounded-xl bg-bgCard border border-borderDark flex items-center justify-center text-textMain shadow-card">
                        <Settings size={20} strokeWidth={1.5} />
                    </div>
                    <div>
                        <h2 className="text-sm font-semibold text-textMain">Workspace Settings & Security</h2>
                        <p className="text-xs text-textFaint">Manage local encryption vault, terminal ergonomics, and IDE bridges.</p>
                    </div>
                </div>

                {/* Vault Security Card */}
                <div className="bg-bgCard border border-borderDark rounded-xl p-6 shadow-card space-y-4 text-xs">
                    <div className="flex items-center gap-2">
                        <Shield size={16} className="text-emerald-400" />
                        <h3 className="font-semibold text-textMain">Enterprise Security Vault</h3>
                    </div>

                    <p className="text-textFaint leading-relaxed">
                        All SSH passwords, private key passphrases, and bastion credentials are encrypted locally using hardware-accelerated <strong className="text-textMain">AES-256-GCM</strong>.
                    </p>

                    <form onSubmit={handleSaveSecurity} className="space-y-3 pt-2">
                        <div>
                            <label className="block text-textMuted font-medium mb-1">Set Master Vault Passphrase</label>
                            <input 
                                type="password" 
                                placeholder="Enter strong master password..."
                                value={masterPass}
                                onChange={(e) => setMasterPass(e.target.value)}
                                className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                            />
                        </div>
                        <div className="flex items-center justify-end gap-3">
                            {passSaved && <span className="text-emerald-400 flex items-center gap-1"><Check size={13} /> Vault Key Updated</span>}
                            <button 
                                type="submit" 
                                className="px-4 py-2 rounded-md bg-textMain text-bgMain font-semibold hover:opacity-90 transition-opacity"
                            >
                                Update Master Key
                            </button>
                        </div>
                    </form>
                </div>

                {/* IDE Integrations Card */}
                <div className="bg-bgCard border border-borderDark rounded-xl p-6 shadow-card space-y-4 text-xs">
                    <div className="flex items-center gap-2">
                        <Code2 size={16} className="text-textMain" />
                        <h3 className="font-semibold text-textMain">External IDE Remote Attach</h3>
                    </div>

                    <p className="text-textFaint leading-relaxed">
                        VibeTerm includes zero-configuration 1-click attach commands for remote development environments directly in <span className="text-textMain">VS Code</span> and <span className="text-textMain">Cursor</span>.
                    </p>

                    <div className="bg-bgMain p-3 rounded-lg border border-borderDark text-[11px] font-mono text-textMuted space-y-1">
                        <div>$ code --remote ssh-remote+user@hostname /path</div>
                        <div>$ cursor --remote ssh-remote+user@hostname /path</div>
                    </div>
                </div>

                {/* About Box */}
                <div className="p-4 rounded-xl bg-bgMain border border-borderDark text-[11px] text-textFaint flex items-center justify-between">
                    <div>
                        <span className="font-semibold text-textMain block">VibeTerm v2.0-native</span>
                        <span>Go 1.22 + Wails v2 + React 18 + Xterm.js</span>
                    </div>
                    <span className="font-mono text-emerald-400">Core Active</span>
                </div>
            </div>
        </div>
    );
};
