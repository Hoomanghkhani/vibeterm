import React, { useState, useEffect } from 'react';
import { Blocks, CheckCircle2, Download, Upload, RefreshCw, Server, Folder, ShieldCheck, Terminal } from 'lucide-react';
import { 
    GetInstalledPlugins, 
    TogglePlugin, 
    DiscoverSSHConfig, 
    DiscoverDockerContainers, 
    DetectSystemTools, 
    ImportTermiusJSON, 
    ImportMobaXterm,
    SaveHost
} from '../wailsjs/go/main/App';

interface PluginsViewProps {
    onReloadHosts: () => void;
}

export const PluginsView: React.FC<PluginsViewProps> = ({ onReloadHosts }) => {
    const [plugins, setPlugins] = useState<any[]>([]);
    const [systemTools, setSystemTools] = useState<any[]>([]);
    const [discoveredSSH, setDiscoveredSSH] = useState<any[]>([]);
    const [discoveredDocker, setDiscoveredDocker] = useState<any[]>([]);
    const [activeTab, setActiveTab] = useState<'installed' | 'discovery' | 'importers'>('installed');
    const [isScanning, setIsScanning] = useState(false);
    const [mobaText, setMobaText] = useState('');
    const [termiusJson, setTermiusJson] = useState('');
    const [importStatus, setImportStatus] = useState('');

    const loadData = async () => {
        try {
            const p = await GetInstalledPlugins();
            setPlugins(p || []);
            const t = await DetectSystemTools();
            setSystemTools(t || []);
        } catch (err) {
            console.error('Failed to load plugins:', err);
        }
    };

    useEffect(() => {
        loadData();
    }, []);

    const handleToggle = async (id: string, currentEnabled: boolean) => {
        await TogglePlugin(id, !currentEnabled);
        loadData();
    };

    const handleScanDiscovery = async () => {
        setIsScanning(true);
        try {
            const sshHosts = await DiscoverSSHConfig();
            setDiscoveredSSH(sshHosts || []);
            const containers = await DiscoverDockerContainers();
            setDiscoveredDocker(containers || []);
        } catch (err) {
            console.error('Discovery error:', err);
        } finally {
            setIsScanning(false);
        }
    };

    const handleImportSSHHost = async (host: any) => {
        await SaveHost(host);
        setDiscoveredSSH(prev => prev.filter(h => h.id !== host.id));
        onReloadHosts();
    };

    const handleImportDockerContainer = async (container: any) => {
        const newHost: any = {
            id: `docker-${container.id.substring(0, 12)}`,
            name: container.names || container.image,
            hostname: '127.0.0.1',
            port: 22,
            protocol: 'docker',
            username: 'root',
            environment: 'dev',
            folder: 'Docker Containers',
            dockerEndpoint: container.id,
            tags: ['docker', container.image],
        };
        await SaveHost(newHost);
        setDiscoveredDocker(prev => prev.filter(c => c.id !== container.id));
        onReloadHosts();
    };

    const handleMobaImport = async () => {
        if (!mobaText.trim()) return;
        const res = await ImportMobaXterm(mobaText);
        for (const h of res.hosts || []) {
            await SaveHost(h);
        }
        setImportStatus(`Successfully imported ${res.importedCount} hosts from MobaXterm!`);
        setMobaText('');
        onReloadHosts();
    };

    const handleTermiusImport = async () => {
        if (!termiusJson.trim()) return;
        try {
            const res = await ImportTermiusJSON(termiusJson);
            for (const h of res.hosts || []) {
                await SaveHost(h);
            }
            setImportStatus(`Successfully imported ${res.importedCount} hosts from Termius!`);
            setTermiusJson('');
            onReloadHosts();
        } catch (err: any) {
            setImportStatus(`Import failed: ${err}`);
        }
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain text-textMain overflow-hidden font-sans p-6">
            {/* Top Header Bar */}
            <div className="flex items-center justify-between pb-4 border-b border-borderDark shrink-0">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-bgPanel rounded-lg border border-borderDark">
                        <Blocks size={18} className="text-textMain" />
                    </div>
                    <div>
                        <h1 className="text-sm font-semibold tracking-wide">Extensions, Discovery & Tool Providers</h1>
                        <p className="text-xs text-textFaint">Manage extensible providers, auto-discover infrastructure, and import from other tools</p>
                    </div>
                </div>

                {/* Tabs Switcher */}
                <div className="flex items-center gap-1 bg-bgCard p-0.5 rounded-lg border border-borderDark text-xs">
                    <button
                        onClick={() => setActiveTab('installed')}
                        className={`px-3 py-1.5 rounded font-medium transition-colors ${
                            activeTab === 'installed' ? 'bg-bgPanel text-textMain shadow-sm font-semibold' : 'text-textFaint hover:text-textMuted'
                        }`}
                    >
                        Installed Plugins
                    </button>
                    <button
                        onClick={() => { setActiveTab('discovery'); handleScanDiscovery(); }}
                        className={`px-3 py-1.5 rounded font-medium transition-colors ${
                            activeTab === 'discovery' ? 'bg-bgPanel text-textMain shadow-sm font-semibold' : 'text-textFaint hover:text-textMuted'
                        }`}
                    >
                        Auto-Discovery
                    </button>
                    <button
                        onClick={() => setActiveTab('importers')}
                        className={`px-3 py-1.5 rounded font-medium transition-colors ${
                            activeTab === 'importers' ? 'bg-bgPanel text-textMain shadow-sm font-semibold' : 'text-textFaint hover:text-textMuted'
                        }`}
                    >
                        Session Importers
                    </button>
                </div>
            </div>

            {/* Tab 1: Installed Plugins & System Tool Detection */}
            {activeTab === 'installed' && (
                <div className="flex-1 grid grid-cols-3 gap-6 pt-6 overflow-hidden min-h-0">
                    {/* Plugins List */}
                    <div className="col-span-2 flex flex-col gap-3 overflow-y-auto pr-1">
                        <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider">Active Providers ({plugins.length})</h2>
                        {plugins.map((p) => (
                            <div key={p.id} className="p-4 bg-bgCard border border-borderDark rounded-xl flex items-start justify-between gap-4">
                                <div className="space-y-1 min-w-0">
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs font-semibold text-textMain">{p.displayName}</span>
                                        <span className="text-[10px] text-textFaint font-mono bg-bgPanel px-1.5 py-0.2 rounded border border-borderDark">v{p.version}</span>
                                        <span className="text-[10px] text-emerald-400 font-mono flex items-center gap-1">
                                            <ShieldCheck size={11} /> Verified
                                        </span>
                                    </div>
                                    <p className="text-xs text-textFaint leading-relaxed">{p.description}</p>
                                    <div className="flex items-center gap-1.5 pt-1">
                                        {p.permissions?.map((perm: string) => (
                                            <span key={perm} className="text-[9px] text-textFaint font-mono uppercase bg-bgMain px-1.5 py-0.5 rounded border border-borderDark">
                                                {perm}
                                            </span>
                                        ))}
                                    </div>
                                </div>

                                <button
                                    onClick={() => handleToggle(p.id, p.enabled)}
                                    className={`px-3 py-1 rounded-lg text-xs font-medium border transition-colors shrink-0 ${
                                        p.enabled 
                                            ? 'bg-bgPanel border-borderActive text-textMain hover:bg-bgHover' 
                                            : 'bg-bgMain border-borderDark text-textFaint hover:text-textMain'
                                    }`}
                                >
                                    {p.enabled ? 'Enabled' : 'Disabled'}
                                </button>
                            </div>
                        ))}
                    </div>

                    {/* Local CLI Tools Status */}
                    <div className="col-span-1 bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col overflow-hidden">
                        <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider mb-3">System CLI Tools</h2>
                        <div className="flex-1 overflow-y-auto space-y-2 pr-1">
                            {systemTools.map((t) => (
                                <div key={t.name} className="p-2.5 bg-bgMain border border-borderDark rounded-lg flex items-center justify-between text-xs">
                                    <div className="flex items-center gap-2 font-mono">
                                        <div className={`w-2 h-2 rounded-full ${t.installed ? 'bg-emerald-400' : 'bg-zinc-600'}`} />
                                        <span className="text-textMain font-semibold">{t.name}</span>
                                    </div>
                                    <span className="text-[10px] text-textFaint font-mono truncate max-w-[120px]">
                                        {t.installed ? t.version : 'Not found'}
                                    </span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            {/* Tab 2: Auto-Discovery (~/.ssh/config & Docker) */}
            {activeTab === 'discovery' && (
                <div className="flex-1 flex flex-col gap-6 pt-6 overflow-hidden min-h-0">
                    <div className="flex items-center justify-between">
                        <span className="text-xs text-textFaint">Auto-discover configured SSH hosts and local Docker containers without manual typing.</span>
                        <button
                            onClick={handleScanDiscovery}
                            disabled={isScanning}
                            className="px-3 py-1.5 bg-bgPanel hover:bg-bgHover border border-borderDark rounded-lg text-xs text-textMain flex items-center gap-1.5 transition-colors"
                        >
                            <RefreshCw size={12} className={isScanning ? 'animate-spin' : ''} />
                            <span>Rescan System</span>
                        </button>
                    </div>

                    <div className="flex-1 grid grid-cols-2 gap-6 overflow-hidden">
                        {/* SSH Config Discovery */}
                        <div className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col overflow-hidden">
                            <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider mb-3 flex items-center justify-between">
                                <span>~/.ssh/config Profiles ({discoveredSSH.length})</span>
                            </h2>
                            <div className="flex-1 overflow-y-auto space-y-2 pr-1">
                                {discoveredSSH.length === 0 ? (
                                    <div className="text-xs text-textFaint text-center py-12">No new unimported SSH config hosts found.</div>
                                ) : (
                                    discoveredSSH.map((h) => (
                                        <div key={h.id} className="p-3 bg-bgMain border border-borderDark rounded-lg flex items-center justify-between text-xs">
                                            <div className="min-w-0">
                                                <div className="font-semibold text-textMain truncate">{h.name}</div>
                                                <div className="text-[11px] text-textFaint font-mono truncate">{h.username}@{h.hostname}:{h.port}</div>
                                            </div>
                                            <button
                                                onClick={() => handleImportSSHHost(h)}
                                                className="px-2.5 py-1 bg-white text-black hover:bg-zinc-200 rounded text-xs font-semibold shrink-0 ml-2 transition-colors"
                                            >
                                                Import Host
                                            </button>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>

                        {/* Docker Discovery */}
                        <div className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col overflow-hidden">
                            <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider mb-3 flex items-center justify-between">
                                <span>Docker Containers ({discoveredDocker.length})</span>
                            </h2>
                            <div className="flex-1 overflow-y-auto space-y-2 pr-1">
                                {discoveredDocker.length === 0 ? (
                                    <div className="text-xs text-textFaint text-center py-12">No active local Docker containers detected.</div>
                                ) : (
                                    discoveredDocker.map((c) => (
                                        <div key={c.id} className="p-3 bg-bgMain border border-borderDark rounded-lg flex items-center justify-between text-xs">
                                            <div className="min-w-0">
                                                <div className="font-semibold text-textMain truncate">{c.names || c.image}</div>
                                                <div className="text-[11px] text-textFaint font-mono truncate">{c.status}</div>
                                            </div>
                                            <button
                                                onClick={() => handleImportDockerContainer(c)}
                                                className="px-2.5 py-1 bg-white text-black hover:bg-zinc-200 rounded text-xs font-semibold shrink-0 ml-2 transition-colors"
                                            >
                                                Add to Tree
                                            </button>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Tab 3: Session Importers */}
            {activeTab === 'importers' && (
                <div className="flex-1 grid grid-cols-2 gap-6 pt-6 overflow-hidden min-h-0">
                    {/* MobaXterm Importer */}
                    <div className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col gap-3">
                        <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider">Import from MobaXterm</h2>
                        <p className="text-xs text-textFaint">Paste exported MobaXterm session entries or .mxtsessions text format below:</p>
                        <textarea
                            rows={8}
                            placeholder="Paste session definitions (e.g. MyServer = #109#1%192.168.1.50%22%root%...)"
                            value={mobaText}
                            onChange={(e) => setMobaText(e.target.value)}
                            className="flex-1 bg-bgMain border border-borderDark rounded-lg p-3 text-xs text-textMain font-mono outline-none focus:border-borderActive resize-none"
                        />
                        <button
                            onClick={handleMobaImport}
                            className="px-4 py-2 bg-white text-black hover:bg-zinc-200 rounded-lg text-xs font-semibold transition-colors shrink-0"
                        >
                            Import MobaXterm Sessions
                        </button>
                    </div>

                    {/* Termius Importer */}
                    <div className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col gap-3">
                        <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider">Import from Termius</h2>
                        <p className="text-xs text-textFaint">Paste exported Termius JSON host configuration below:</p>
                        <textarea
                            rows={8}
                            placeholder='Paste Termius JSON (e.g. { "hosts": [ { "label": "Web", "address": "10.0.0.1", "username": "admin" } ] })'
                            value={termiusJson}
                            onChange={(e) => setTermiusJson(e.target.value)}
                            className="flex-1 bg-bgMain border border-borderDark rounded-lg p-3 text-xs text-textMain font-mono outline-none focus:border-borderActive resize-none"
                        />
                        <button
                            onClick={handleTermiusImport}
                            className="px-4 py-2 bg-white text-black hover:bg-zinc-200 rounded-lg text-xs font-semibold transition-colors shrink-0"
                        >
                            Import Termius JSON
                        </button>
                    </div>
                </div>
            )}

            {/* Import Status Alert */}
            {importStatus && (
                <div className="mt-4 p-3 bg-bgPanel border border-borderActive rounded-lg text-xs text-emerald-400 flex items-center justify-between">
                    <span>{importStatus}</span>
                    <button onClick={() => setImportStatus('')} className="text-textFaint hover:text-white">✕</button>
                </div>
            )}
        </div>
    );
};
