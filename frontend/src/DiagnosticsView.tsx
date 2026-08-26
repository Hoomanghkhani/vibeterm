import React, { useState } from 'react';
import { Activity, Search, CheckCircle2, XCircle, Globe, Terminal, ShieldAlert } from 'lucide-react';
import { TestDiagnosticsTCP, TestDiagnosticsDNS } from '../wailsjs/go/main/App';

export const DiagnosticsView: React.FC = () => {
    const [targetHost, setTargetHost] = useState('');
    const [targetPort, setTargetPort] = useState(22);
    const [isRunning, setIsRunning] = useState(false);
    const [results, setResults] = useState<any[]>([]);

    const handleRunProbe = async () => {
        if (!targetHost.trim()) return;
        setIsRunning(true);

        try {
            // Run DNS
            const dnsRes = await TestDiagnosticsDNS(targetHost);
            // Run TCP
            const tcpRes = await TestDiagnosticsTCP(targetHost, targetPort);

            setResults([
                {
                    id: `diag-${Date.now()}-tcp`,
                    type: 'TCP Handshake',
                    target: `${targetHost}:${targetPort}`,
                    success: tcpRes.success,
                    latency: tcpRes.latencyMs,
                    message: tcpRes.message,
                    time: new Date().toLocaleTimeString(),
                },
                {
                    id: `diag-${Date.now()}-dns`,
                    type: 'DNS Lookup',
                    target: targetHost,
                    success: dnsRes.success,
                    latency: dnsRes.latencyMs,
                    message: dnsRes.message,
                    ips: dnsRes.ips,
                    time: new Date().toLocaleTimeString(),
                },
                ...results,
            ]);
        } catch (err: any) {
            console.error('Diagnostics failed:', err);
        } finally {
            setIsRunning(false);
        }
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain text-textMain overflow-hidden font-sans p-6">
            {/* Header */}
            <div className="flex items-center justify-between pb-4 border-b border-borderDark shrink-0">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-bgPanel rounded-lg border border-borderDark">
                        <Activity size={18} className="text-textMain" />
                    </div>
                    <div>
                        <h1 className="text-sm font-semibold tracking-wide">Infrastructure & Network Diagnostics</h1>
                        <p className="text-xs text-textFaint">Probe network reachability, TCP handshake latency, and DNS resolution</p>
                    </div>
                </div>
            </div>

            {/* Probe Controls Bar */}
            <div className="bg-bgCard border border-borderDark rounded-xl p-4 my-6 flex items-center gap-3 shrink-0">
                <div className="flex-1 flex items-center gap-2 bg-bgMain border border-borderDark focus-within:border-borderActive rounded-lg px-3 py-2">
                    <Globe size={14} className="text-textFaint shrink-0" />
                    <input 
                        type="text"
                        placeholder="Target Hostname or IP (e.g. '1.1.1.1', 'api.github.com', '192.168.1.1')..."
                        value={targetHost}
                        onChange={(e) => setTargetHost(e.target.value)}
                        onKeyDown={(e) => { if (e.key === 'Enter') handleRunProbe(); }}
                        className="flex-1 bg-transparent text-xs text-textMain placeholder:text-textFaint outline-none font-mono"
                    />
                </div>
                <div className="w-28 flex items-center gap-2 bg-bgMain border border-borderDark focus-within:border-borderActive rounded-lg px-3 py-2">
                    <span className="text-xs text-textFaint font-mono">Port:</span>
                    <input 
                        type="number"
                        value={targetPort}
                        onChange={(e) => setTargetPort(parseInt(e.target.value) || 22)}
                        className="w-full bg-transparent text-xs text-textMain outline-none font-mono"
                    />
                </div>
                <button
                    onClick={handleRunProbe}
                    disabled={isRunning || !targetHost.trim()}
                    className="px-4 py-2 bg-white text-black hover:bg-zinc-200 disabled:opacity-40 rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-colors shadow-sm"
                >
                    <Activity size={13} className={isRunning ? 'animate-pulse' : ''} />
                    <span>Run Probe</span>
                </button>
            </div>

            {/* Probe Results List */}
            <div className="flex-1 bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col overflow-hidden">
                <h2 className="text-xs font-semibold text-textFaint uppercase tracking-wider mb-3">Diagnostic Results</h2>
                <div className="flex-1 overflow-y-auto space-y-2.5 pr-1">
                    {results.length === 0 ? (
                        <div className="text-xs text-textFaint text-center py-16">
                            Enter a target hostname and port above to run real-time TCP & DNS diagnostic probes.
                        </div>
                    ) : (
                        results.map((r) => (
                            <div key={r.id} className="p-3.5 bg-bgMain border border-borderDark rounded-lg flex items-start justify-between gap-4 text-xs">
                                <div className="space-y-1 min-w-0">
                                    <div className="flex items-center gap-2 font-mono">
                                        {r.success ? (
                                            <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />
                                        ) : (
                                            <XCircle size={14} className="text-rose-400 shrink-0" />
                                        )}
                                        <span className="font-semibold text-textMain">{r.type}: {r.target}</span>
                                        <span className={`text-[10px] font-mono px-1.5 py-0.2 rounded border ${
                                            r.success ? 'bg-emerald-950/40 text-emerald-400 border-emerald-800/40' : 'bg-rose-950/40 text-rose-400 border-rose-800/40'
                                        }`}>
                                            {r.latency?.toFixed(2)} ms
                                        </span>
                                    </div>
                                    <p className="text-textFaint font-mono text-[11px]">{r.message}</p>
                                    {r.ips && r.ips.length > 0 && (
                                        <div className="flex items-center gap-1.5 pt-1">
                                            {r.ips.map((ip: string) => (
                                                <span key={ip} className="text-[10px] text-zinc-300 font-mono bg-bgPanel px-1.5 py-0.5 rounded border border-borderDark">
                                                    {ip}
                                                </span>
                                            ))}
                                        </div>
                                    )}
                                </div>
                                <span className="text-[10px] text-textFaint font-mono shrink-0">{r.time}</span>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </div>
    );
};
