import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { StartLocalTerminal, StartSSHTerminal, StartDockerTerminal, SendTerminalInput, ResizeTerminal, CloseTerminal } from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import { Copy, Clipboard, CheckSquare, Trash2 } from 'lucide-react';

interface TerminalProps {
    sessionType: 'local' | 'ssh' | 'docker';
    hostID?: string;
}

export const TerminalComponent: React.FC<TerminalProps> = ({ sessionType, hostID }) => {
    const terminalRef = useRef<HTMLDivElement>(null);
    const sessionIDRef = useRef<string | null>(null);
    const xtermRef = useRef<Terminal | null>(null);

    // Context Menu State
    const [menu, setMenu] = useState<{ visible: boolean; x: number; y: number }>({
        visible: false,
        x: 0,
        y: 0,
    });

    useEffect(() => {
        if (!terminalRef.current) return;

        // Initialize Xterm
        const term = new Terminal({
            theme: {
                background: '#09090b',
                foreground: '#ffffff',
                cursor: '#ffffff',
                selectionBackground: 'rgba(255, 255, 255, 0.25)',
                black: '#09090b',
                red: '#f87171',
                green: '#4ade80',
                yellow: '#facc15',
                blue: '#60a5fa',
                magenta: '#c084fc',
                cyan: '#38bdf8',
                white: '#f4f4f5',
                brightBlack: '#71717a',
                brightRed: '#ef4444',
                brightGreen: '#22c55e',
                brightYellow: '#eab308',
                brightBlue: '#3b82f6',
                brightMagenta: '#a855f7',
                brightCyan: '#06b6d4',
                brightWhite: '#ffffff',
            },
            fontFamily: '"JetBrains Mono", "Fira Code", monospace',
            fontSize: 13,
            lineHeight: 1.25,
            cursorBlink: true,
            cursorStyle: 'block',
            allowTransparency: true,
        });

        const fitAddon = new FitAddon();
        term.loadAddon(fitAddon);
        term.open(terminalRef.current);
        fitAddon.fit();

        xtermRef.current = term;

        // Start backend session
        const initSession = async () => {
            try {
                let sid = '';
                const cols = term.cols || 80;
                const rows = term.rows || 24;

                if (sessionType === 'local') {
                    sid = await StartLocalTerminal(cols, rows);
                } else if (sessionType === 'ssh' && hostID) {
                    sid = await StartSSHTerminal(hostID, cols, rows);
                } else if (sessionType === 'docker' && hostID) {
                    sid = await StartDockerTerminal(hostID, cols, rows);
                }

                sessionIDRef.current = sid;

                if (sid) {
                    EventsOn(`terminal:output:${sid}`, (data: string) => {
                        term.write(data);
                    });
                }

                // Handle keyboard input
                term.onData((data) => {
                    if (sessionIDRef.current) {
                        SendTerminalInput(sessionIDRef.current, data);
                    }
                });

                // Handle resize
                const handleResize = () => {
                    try {
                        fitAddon.fit();
                        if (sessionIDRef.current) {
                            ResizeTerminal(sessionIDRef.current, term.cols, term.rows);
                        }
                    } catch (e) {
                        // ignore if element is not rendered
                    }
                };

                window.addEventListener('resize', handleResize);

                // ResizeObserver on terminal DOM element
                let resizeObserver: ResizeObserver | null = null;
                if (terminalRef.current && window.ResizeObserver) {
                    resizeObserver = new ResizeObserver(() => {
                        handleResize();
                    });
                    resizeObserver.observe(terminalRef.current);
                }

                return () => {
                    window.removeEventListener('resize', handleResize);
                    if (resizeObserver) {
                        resizeObserver.disconnect();
                    }
                };
            } catch (err: any) {
                term.writeln(`\r\n\x1b[31m[Error starting session]: ${err}\x1b[0m\r\n`);
            }
        };

        const cleanupPromise = initSession();

        // Right-click Context Menu Listener
        const handleContextMenu = (e: MouseEvent) => {
            e.preventDefault();
            setMenu({
                visible: true,
                x: e.clientX,
                y: e.clientY,
            });
        };

        const currentTermElement = terminalRef.current;
        currentTermElement.addEventListener('contextmenu', handleContextMenu);

        // Click outside closes menu
        const handleGlobalClick = () => {
            setMenu((prev) => (prev.visible ? { ...prev, visible: false } : prev));
        };
        window.addEventListener('click', handleGlobalClick);

        return () => {
            cleanupPromise.then((cleanupResize) => {
                if (cleanupResize) {
                    cleanupResize();
                }
            });
            window.removeEventListener('click', handleGlobalClick);
            if (currentTermElement) {
                currentTermElement.removeEventListener('contextmenu', handleContextMenu);
            }
            if (sessionIDRef.current) {
                EventsOff(`terminal:output:${sessionIDRef.current}`);
                CloseTerminal(sessionIDRef.current);
            }
            term.dispose();
        };
    }, [sessionType, hostID]);

    // Context Menu Actions
    const handleCopy = () => {
        if (xtermRef.current) {
            const selection = xtermRef.current.getSelection();
            if (selection) {
                navigator.clipboard.writeText(selection);
            }
        }
        setMenu({ ...menu, visible: false });
    };

    const handlePaste = async () => {
        try {
            const text = await navigator.clipboard.readText();
            if (text && sessionIDRef.current) {
                SendTerminalInput(sessionIDRef.current, text);
            }
        } catch (err) {
            console.error('Failed to read clipboard', err);
        }
        setMenu({ ...menu, visible: false });
    };

    const handleSelectAll = () => {
        if (xtermRef.current) {
            xtermRef.current.selectAll();
        }
        setMenu({ ...menu, visible: false });
    };

    const handleClear = () => {
        if (xtermRef.current) {
            xtermRef.current.clear();
        }
        setMenu({ ...menu, visible: false });
    };

    return (
        <div className="relative w-full h-full bg-bgMain">
            <div ref={terminalRef} className="w-full h-full" />

            {/* Custom VS Code Style Right-Click Menu */}
            {menu.visible && (
                <div
                    className="fixed z-50 w-44 bg-bgCard border border-borderDark rounded-lg shadow-2xl py-1 text-xs text-textMain overflow-hidden select-none"
                    style={{ top: `${menu.y}px`, left: `${menu.x}px` }}
                    onClick={(e) => e.stopPropagation()}
                >
                    <button
                        onClick={handleCopy}
                        className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                    >
                        <div className="flex items-center gap-2">
                            <Copy size={13} className="text-textFaint" />
                            <span>Copy</span>
                        </div>
                        <span className="text-[10px] text-textFaint font-mono">Ctrl+C</span>
                    </button>

                    <button
                        onClick={handlePaste}
                        className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                    >
                        <div className="flex items-center gap-2">
                            <Clipboard size={13} className="text-textFaint" />
                            <span>Paste</span>
                        </div>
                        <span className="text-[10px] text-textFaint font-mono">Ctrl+V</span>
                    </button>

                    <button
                        onClick={handleSelectAll}
                        className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                    >
                        <div className="flex items-center gap-2">
                            <CheckSquare size={13} className="text-textFaint" />
                            <span>Select All</span>
                        </div>
                        <span className="text-[10px] text-textFaint font-mono">Ctrl+A</span>
                    </button>

                    <div className="h-[1px] bg-borderDark my-1"></div>

                    <button
                        onClick={handleClear}
                        className="w-full px-3 py-1.5 flex items-center justify-between hover:bg-bgHover hover:text-white transition-colors"
                    >
                        <div className="flex items-center gap-2">
                            <Trash2 size={13} className="text-textFaint" />
                            <span>Clear Terminal</span>
                        </div>
                    </button>
                </div>
            )}
        </div>
    );
};
