import React, { useState } from 'react';
import { TerminalComponent } from './TerminalComponent';
import { Split, X, Terminal, Plus, Columns, Rows, Grid } from 'lucide-react';

export interface TerminalPane {
    id: string;
    title: string;
    type: 'local' | 'ssh' | 'docker';
    hostID?: string;
}

export type SplitMode = 'none' | 'vertical' | 'horizontal' | 'quad';

interface TerminalSplitLayoutProps {
    tabId: string;
    panes: TerminalPane[];
    splitMode: SplitMode;
    activePaneId: string;
    onSelectPane: (paneId: string) => void;
    onClosePane: (paneId: string) => void;
    onSplitPane: (mode: SplitMode) => void;
}

export const TerminalSplitLayout: React.FC<TerminalSplitLayoutProps> = ({
    tabId,
    panes,
    splitMode,
    activePaneId,
    onSelectPane,
    onClosePane,
    onSplitPane,
}) => {
    // If only 1 pane or splitMode === 'none'
    if (panes.length <= 1 || splitMode === 'none') {
        const singlePane = panes[0] || { id: `pane-${tabId}-0`, title: 'Terminal', type: 'local' };
        return (
            <div className="w-full h-full relative">
                <TerminalComponent 
                    key={singlePane.id}
                    sessionType={singlePane.type}
                    hostID={singlePane.hostID}
                />
            </div>
        );
    }

    // Grid classes based on split mode
    let containerClass = 'grid w-full h-full gap-[1px] bg-borderDark/80';
    if (splitMode === 'vertical') {
        containerClass += ' grid-cols-2 grid-rows-1';
    } else if (splitMode === 'horizontal') {
        containerClass += ' grid-cols-1 grid-rows-2';
    } else if (splitMode === 'quad') {
        containerClass += ' grid-cols-2 grid-rows-2';
    }

    return (
        <div className={containerClass}>
            {panes.map((pane, idx) => {
                const isActive = pane.id === activePaneId;
                return (
                    <div 
                        key={pane.id}
                        onClick={() => onSelectPane(pane.id)}
                        className={`flex flex-col bg-bgMain overflow-hidden relative border transition-colors ${
                            isActive ? 'border-borderActive' : 'border-borderDark/40'
                        }`}
                    >
                        {/* Split Pane Mini Header Bar */}
                        <div className={`h-6 px-2 flex items-center justify-between text-[11px] border-b shrink-0 select-none ${
                            isActive ? 'bg-bgPanel text-textMain border-borderDark' : 'bg-bgCard text-textFaint border-borderDark/50'
                        }`}>
                            <div className="flex items-center gap-1.5 min-w-0">
                                <Terminal size={11} className={isActive ? 'text-textMain' : 'text-textFaint'} />
                                <span className="truncate font-medium">{pane.title}</span>
                            </div>

                            <div className="flex items-center gap-1">
                                <button 
                                    onClick={(e) => { e.stopPropagation(); onSplitPane(splitMode === 'vertical' ? 'horizontal' : 'vertical'); }}
                                    className="p-0.5 hover:text-textMain rounded hover:bg-bgHover transition-colors text-textFaint"
                                    title="Toggle Split Direction"
                                >
                                    <Split size={10} className={splitMode === 'horizontal' ? 'rotate-90' : ''} />
                                </button>
                                {panes.length > 1 && (
                                    <button 
                                        onClick={(e) => { e.stopPropagation(); onClosePane(pane.id); }}
                                        className="p-0.5 hover:text-rose-400 rounded hover:bg-bgHover transition-colors text-textFaint"
                                        title="Close Pane"
                                    >
                                        <X size={10} />
                                    </button>
                                )}
                            </div>
                        </div>

                        {/* Terminal Canvas */}
                        <div className="flex-1 overflow-hidden relative">
                            <TerminalComponent 
                                sessionType={pane.type}
                                hostID={pane.hostID}
                            />
                        </div>
                    </div>
                );
            })}
        </div>
    );
};
