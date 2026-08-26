import React, { useState } from 'react';
import { Radio, Search, Plus, Play, CheckCircle2, Shield, Globe, Terminal, Loader2, ArrowUpRight } from 'lucide-react';
import { ScanSubnet } from '../wailsjs/go/main/App';

interface ScannerViewProps {
    onImportHost: (device: any) => void;
}

export const ScannerView: React.FC<ScannerViewProps> = ({ onImportHost }) => {
    const [cidr, setCidr] = useState('192.168.1.0/24');
    const [scanning, setScanning] = useState(false);
    const [devices, setDevices] = useState<any[]>([]);
    const [error, setError] = useState<string | null>(null);

    const handleScan = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!cidr) return;
        setScanning(true);
        setError(null);
        setDevices([]);
        try {
            const results = await ScanSubnet(cidr, [22, 3389, 80, 443, 8080, 5900]);
            setDevices(results || []);
        } catch (err: any) {
            setError(String(err));
        } finally {
            setScanning(false);
        }
    };

    return (
        <div className="flex-1 flex flex-col bg-bgMain overflow-hidden select-none">
            {/* Scan Toolbar */}
            <div className="h-14 px-6 border-b border-borderDark flex items-center justify-between shrink-0 bg-bgCard">
                <form onSubmit={handleScan} className="flex items-center gap-3 w-full max-w-xl">
                    <div className="flex-1 h-9 bg-bgMain border border-borderDark rounded-lg flex items-center px-3 text-xs text-textMuted gap-2">
                        <Radio size={14} className="text-textFaint" />
                        <input 
                            type="text" 
                            placeholder="Target Subnet CIDR (e.g. 192.168.1.0/24 or 10.0.0.1)" 
                            value={cidr}
                            onChange={(e) => setCidr(e.target.value)}
                            className="bg-transparent border-none outline-none w-full text-textMain font-mono placeholder:text-textFaint"
                        />
                    </div>

                    <button 
                        type="submit"
                        disabled={scanning}
                        className="px-4 py-2 rounded-lg bg-textMain text-bgMain font-medium text-xs flex items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-40 shrink-0"
                    >
                        {scanning ? <Loader2 size={13} className="animate-spin" /> : <Play size={13} />}
                        <span>{scanning ? 'Scanning...' : 'Scan Subnet'}</span>
                    </button>
                </form>

                <div className="text-xs text-textFaint font-mono">
                    {devices.length > 0 && `${devices.length} hosts discovered`}
                </div>
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-y-auto p-6">
                {error && (
                    <div className="p-4 mb-4 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs">
                        Scan error: {error}
                    </div>
                )}

                {devices.length === 0 && !scanning ? (
                    <div className="h-full flex flex-col items-center justify-center text-center p-8 text-textMuted">
                        <Radio size={36} className="text-textFaint mb-3 opacity-60" />
                        <h3 className="text-sm font-semibold text-textMain">Network Discovery</h3>
                        <p className="text-xs text-textFaint max-w-sm mt-1">Enter a CIDR subnet or single IP to scan for SSH, RDP, and HTTP servers.</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {devices.map((d, i) => (
                            <div key={i} className="bg-bgCard border border-borderDark rounded-xl p-4 flex flex-col justify-between shadow-card hover:border-borderActive transition-all">
                                <div>
                                    <div className="flex items-start justify-between mb-2">
                                        <div className="flex items-center gap-2">
                                            <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                                            <h4 className="text-xs font-semibold text-textMain font-mono">{d.ip}</h4>
                                        </div>
                                        <span className="text-[10px] text-emerald-400 font-mono">
                                            {d.latencyMs ? `${d.latencyMs.toFixed(1)}ms` : '<1ms'}
                                        </span>
                                    </div>

                                    {/* Services list */}
                                    <div className="my-2 space-y-1">
                                        {d.services && d.services.map((srv: string, idx: number) => (
                                            <div key={idx} className="text-[11px] font-mono text-textMuted bg-bgMain px-2 py-1 rounded border border-borderDark truncate">
                                                {srv}
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                <div className="pt-3 border-t border-borderDark flex items-center justify-between mt-3">
                                    <span className="text-[10px] font-mono uppercase text-textFaint">{d.matchedProto || 'Unknown'}</span>
                                    <button 
                                        onClick={() => onImportHost(d)}
                                        className="px-3 py-1 rounded bg-bgMain border border-borderDark hover:border-borderActive text-textMain text-xs font-medium flex items-center gap-1.5 transition-colors"
                                    >
                                        <Plus size={12} />
                                        <span>Add as Host</span>
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};
