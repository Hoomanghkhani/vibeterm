import React, { useState, useEffect, useRef } from 'react';
import { 
    Folder, 
    File, 
    FileText, 
    FileCode, 
    FolderPlus, 
    Upload, 
    RefreshCw, 
    Trash2, 
    ChevronRight, 
    ArrowUp, 
    Download, 
    Eye, 
    X, 
    Check, 
    Server, 
    Copy,
    AlertCircle
} from 'lucide-react';
import { ListRemoteFiles, UploadRemoteFile, ReadRemoteFile, DeleteRemoteFile, CreateRemoteFolder } from '../wailsjs/go/main/App';

interface FileExplorerViewProps {
    hosts: any[];
    activeHostId?: string;
}

export const FileExplorerView: React.FC<FileExplorerViewProps> = ({ hosts = [], activeHostId }) => {
    const [selectedHostId, setSelectedHostId] = useState<string>(activeHostId || hosts[0]?.id || hosts[0]?.ID || '');
    const [currentPath, setCurrentPath] = useState<string>('/root');
    const [pathInput, setPathInput] = useState<string>('/root');
    const [files, setFiles] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Drag and drop state
    const [isDragging, setIsDragging] = useState(false);
    const [uploading, setUploading] = useState(false);

    // File Preview Modal
    const [previewFile, setPreviewFile] = useState<{ path: string; name: string; content: string } | null>(null);
    const [previewLoading, setPreviewLoading] = useState(false);

    // New Folder Modal
    const [showNewFolderModal, setShowNewFolderModal] = useState(false);
    const [newFolderName, setNewFolderName] = useState('');

    const fileInputRef = useRef<HTMLInputElement>(null);

    const getHostName = (h: any) => h?.name || h?.Name || 'Server';
    const getHostId = (h: any) => h?.id || h?.ID || '';

    // If activeHostId changes externally, update selected host
    useEffect(() => {
        if (activeHostId) {
            setSelectedHostId(activeHostId);
        }
    }, [activeHostId]);

    const loadFiles = async (path: string = currentPath) => {
        if (!selectedHostId) return;
        setLoading(true);
        setError(null);
        try {
            const list = await ListRemoteFiles(selectedHostId, path);
            setFiles(list || []);
            setCurrentPath(path);
            setPathInput(path);
        } catch (err: any) {
            setError(String(err));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (selectedHostId) {
            loadFiles('/root').catch(() => loadFiles('/'));
        }
    }, [selectedHostId]);

    const handleNavigate = (newPath: string) => {
        loadFiles(newPath);
    };

    const handleGoUp = () => {
        const segments = currentPath.split('/').filter(Boolean);
        if (segments.length <= 1) {
            handleNavigate('/');
        } else {
            segments.pop();
            handleNavigate('/' + segments.join('/'));
        }
    };

    const handleFileClick = (file: any) => {
        if (file.isDir) {
            handleNavigate(file.path);
        } else {
            handlePreview(file);
        }
    };

    const handlePreview = async (file: any) => {
        setPreviewLoading(true);
        setPreviewFile({ path: file.path, name: file.name, content: '' });
        try {
            const content = await ReadRemoteFile(selectedHostId, file.path);
            setPreviewFile({ path: file.path, name: file.name, content });
        } catch (err: any) {
            setPreviewFile({ path: file.path, name: file.name, content: `Error reading file: ${err}` });
        } finally {
            setPreviewLoading(false);
        }
    };

    const handleDelete = async (e: React.MouseEvent, file: any) => {
        e.stopPropagation();
        if (!confirm(`Are you sure you want to delete ${file.name}?`)) return;
        try {
            await DeleteRemoteFile(selectedHostId, file.path);
            loadFiles(currentPath);
        } catch (err: any) {
            alert(`Delete failed: ${err}`);
        }
    };

    const handleCreateFolder = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!newFolderName) return;
        const targetPath = currentPath.endsWith('/') ? `${currentPath}${newFolderName}` : `${currentPath}/${newFolderName}`;
        try {
            await CreateRemoteFolder(selectedHostId, targetPath);
            setShowNewFolderModal(false);
            setNewFolderName('');
            loadFiles(currentPath);
        } catch (err: any) {
            alert(`Folder creation failed: ${err}`);
        }
    };

    // Drag & drop upload handlers
    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(true);
    };

    const handleDragLeave = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);
    };

    const handleDrop = async (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);

        const droppedFiles = Array.from(e.dataTransfer.files);
        if (droppedFiles.length === 0) return;

        setUploading(true);
        try {
            for (const file of droppedFiles) {
                const reader = new FileReader();
                await new Promise<void>((resolve, reject) => {
                    reader.onload = async () => {
                        try {
                            const base64Content = (reader.result as string).split(',')[1] || '';
                            const targetPath = currentPath.endsWith('/') ? `${currentPath}${file.name}` : `${currentPath}/${file.name}`;
                            await UploadRemoteFile(selectedHostId, targetPath, base64Content);
                            resolve();
                        } catch (uploadErr) {
                            reject(uploadErr);
                        }
                    };
                    reader.onerror = reject;
                    reader.readAsDataURL(file);
                });
            }
            loadFiles(currentPath);
        } catch (err: any) {
            alert(`Upload failed: ${err}`);
        } finally {
            setUploading(false);
        }
    };

    const formatSize = (bytes: number) => {
        if (!bytes) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    };

    const getFileIcon = (file: any) => {
        if (file.isDir) return <Folder size={16} className="text-amber-400 fill-amber-400/20" />;
        const ext = file.name.split('.').pop()?.toLowerCase();
        if (['js', 'ts', 'tsx', 'jsx', 'py', 'go', 'rs', 'c', 'cpp', 'html', 'css', 'json', 'yaml', 'yml', 'sh'].includes(ext)) {
            return <FileCode size={16} className="text-blue-400" />;
        }
        if (['txt', 'md', 'log', 'conf', 'cfg', 'env'].includes(ext)) {
            return <FileText size={16} className="text-zinc-300" />;
        }
        return <File size={16} className="text-zinc-400" />;
    };

    const breadcrumbs = currentPath.split('/').filter(Boolean);

    return (
        <div 
            className="flex-1 flex flex-col bg-bgMain overflow-hidden select-none relative"
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
        >
            {/* Drag & Drop Overlay */}
            {isDragging && (
                <div className="absolute inset-0 bg-bgMain/90 backdrop-blur-sm z-50 flex flex-col items-center justify-center border-2 border-dashed border-textMain m-3 rounded-2xl pointer-events-none">
                    <Upload size={48} className="text-textMain mb-2 animate-bounce" />
                    <span className="text-sm font-semibold text-textMain">Drop files here to upload via SFTP</span>
                    <span className="text-xs text-textFaint mt-1">Uploading directly into {currentPath}</span>
                </div>
            )}

            {/* Top Toolbar */}
            <div className="h-12 px-6 border-b border-borderDark flex items-center justify-between shrink-0 bg-bgCard gap-4">
                {/* Host Selector */}
                <div className="flex items-center gap-2">
                    <Server size={15} className="text-textMuted" />
                    <select
                        value={selectedHostId}
                        onChange={(e) => setSelectedHostId(e.target.value)}
                        className="bg-bgMain border border-borderDark rounded-md px-2.5 py-1.5 text-xs text-textMain outline-none focus:border-borderActive"
                    >
                        {hosts.map((h, i) => (
                            <option key={i} value={getHostId(h)}>{getHostName(h)} ({h.hostname || h.Hostname})</option>
                        ))}
                    </select>
                </div>

                {/* Path Navigation Bar */}
                <div className="flex-1 max-w-xl h-8 bg-bgMain border border-borderDark rounded-lg flex items-center px-2 text-xs text-textMuted gap-1.5">
                    <button 
                        onClick={handleGoUp} 
                        className="p-1 hover:bg-bgPanel rounded text-textFaint hover:text-textMain"
                        title="Go up one directory"
                    >
                        <ArrowUp size={13} />
                    </button>
                    <div className="h-4 w-[1px] bg-borderDark mx-1"></div>
                    <form 
                        onSubmit={(e) => { e.preventDefault(); handleNavigate(pathInput); }} 
                        className="flex-1 flex items-center"
                    >
                        <input 
                            type="text" 
                            value={pathInput}
                            onChange={(e) => setPathInput(e.target.value)}
                            className="bg-transparent border-none outline-none w-full text-textMain font-mono text-xs"
                            placeholder="/root"
                        />
                    </form>
                </div>

                {/* Action Buttons */}
                <div className="flex items-center gap-2">
                    <button 
                        onClick={() => setShowNewFolderModal(true)}
                        className="p-1.5 rounded-lg bg-bgMain border border-borderDark hover:text-textMain text-textFaint transition-colors"
                        title="New Folder"
                    >
                        <FolderPlus size={14} />
                    </button>
                    <button 
                        onClick={() => loadFiles(currentPath)}
                        className="p-1.5 rounded-lg bg-bgMain border border-borderDark hover:text-textMain text-textFaint transition-colors"
                        title="Refresh Directory"
                    >
                        <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
                    </button>
                </div>
            </div>

            {/* Breadcrumbs Row */}
            <div className="h-8 px-6 bg-bgSidebar border-b border-borderDark flex items-center text-xs font-mono text-textFaint gap-1 overflow-x-auto no-scrollbar shrink-0">
                <button onClick={() => handleNavigate('/')} className="hover:text-textMain transition-colors">root</button>
                {breadcrumbs.map((seg, idx) => {
                    const segPath = '/' + breadcrumbs.slice(0, idx + 1).join('/');
                    return (
                        <React.Fragment key={idx}>
                            <ChevronRight size={12} className="opacity-50" />
                            <button onClick={() => handleNavigate(segPath)} className="hover:text-textMain transition-colors">
                                {seg}
                            </button>
                        </React.Fragment>
                    );
                })}
            </div>

            {/* File List Grid / Table */}
            <div className="flex-1 overflow-y-auto p-6">
                {error && (
                    <div className="p-4 mb-4 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center gap-2">
                        <AlertCircle size={15} />
                        <span>SFTP error: {error}</span>
                    </div>
                )}

                {loading ? (
                    <div className="h-full flex flex-col items-center justify-center text-textMuted text-xs gap-2">
                        <RefreshCw size={24} className="animate-spin text-textFaint" />
                        <span>Loading remote directory...</span>
                    </div>
                ) : files.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-textMuted text-xs">
                        <Folder size={36} className="text-textFaint mb-2 opacity-50" />
                        <span>This directory is empty.</span>
                        <span className="text-textFaint text-[11px] mt-1">Drag and drop files here to upload via SFTP.</span>
                    </div>
                ) : (
                    <div className="bg-bgCard border border-borderDark rounded-xl overflow-hidden shadow-card">
                        <div className="grid grid-cols-12 px-4 py-2 text-[11px] font-semibold text-textFaint border-b border-borderDark uppercase tracking-wider">
                            <span className="col-span-6">Name</span>
                            <span className="col-span-2">Size</span>
                            <span className="col-span-2">Permissions</span>
                            <span className="col-span-2 text-right">Modified</span>
                        </div>

                        <div className="divide-y divide-borderDark/40 font-mono text-xs">
                            {files.map((file, idx) => (
                                <div 
                                    key={idx}
                                    onClick={() => handleFileClick(file)}
                                    className="grid grid-cols-12 px-4 py-2.5 items-center hover:bg-bgPanel/60 cursor-pointer transition-colors group"
                                >
                                    <div className="col-span-6 flex items-center gap-2.5 min-w-0">
                                        {getFileIcon(file)}
                                        <span className="truncate text-textMain group-hover:text-white font-sans text-xs">{file.name}</span>
                                    </div>
                                    <span className="col-span-2 text-textFaint text-[11px]">
                                        {file.isDir ? '-' : formatSize(file.size)}
                                    </span>
                                    <span className="col-span-2 text-textFaint text-[11px]">{file.mode}</span>
                                    <div className="col-span-2 flex items-center justify-end gap-2 text-textFaint text-[11px]">
                                        <span className="group-hover:hidden">{file.modTime?.split(' ')[0]}</span>
                                        <div className="hidden group-hover:flex items-center gap-1">
                                            {!file.isDir && (
                                                <button 
                                                    onClick={(e) => { e.stopPropagation(); handlePreview(file); }}
                                                    className="p-1 hover:text-textMain rounded hover:bg-bgMain"
                                                    title="Preview / Edit"
                                                >
                                                    <Eye size={13} />
                                                </button>
                                            )}
                                            <button 
                                                onClick={(e) => handleDelete(e, file)}
                                                className="p-1 hover:text-rose-400 rounded hover:bg-bgMain"
                                                title="Delete"
                                            >
                                                <Trash2 size={13} />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>

            {/* File Preview Modal */}
            {previewFile && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-6">
                    <div className="w-full max-w-3xl h-[80vh] bg-bgCard border border-borderDark rounded-xl shadow-2xl flex flex-col overflow-hidden">
                        <div className="h-11 px-4 border-b border-borderDark flex items-center justify-between bg-bgPanel shrink-0">
                            <div className="flex items-center gap-2 min-w-0">
                                <FileCode size={15} className="text-textMuted" />
                                <span className="text-xs font-semibold text-textMain truncate">{previewFile.path}</span>
                            </div>
                            <button onClick={() => setPreviewFile(null)} className="text-textFaint hover:text-textMain">
                                <X size={15} />
                            </button>
                        </div>
                        <div className="flex-1 p-4 bg-bgMain overflow-y-auto">
                            {previewLoading ? (
                                <div className="h-full flex items-center justify-center text-xs text-textFaint">Loading file content...</div>
                            ) : (
                                <pre className="font-mono text-xs text-textMain leading-relaxed select-text whitespace-pre-wrap">
                                    {previewFile.content}
                                </pre>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {/* New Folder Modal */}
            {showNewFolderModal && (
                <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4">
                    <div className="w-80 bg-bgCard border border-borderDark rounded-xl shadow-2xl overflow-hidden">
                        <div className="h-10 px-4 border-b border-borderDark flex items-center justify-between bg-bgPanel">
                            <span className="text-xs font-semibold text-textMain">Create New Directory</span>
                            <button onClick={() => setShowNewFolderModal(false)} className="text-textFaint hover:text-textMain"><X size={14} /></button>
                        </div>
                        <form onSubmit={handleCreateFolder} className="p-4 space-y-3 text-xs">
                            <input 
                                type="text"
                                placeholder="Folder name (e.g. backup)"
                                value={newFolderName}
                                onChange={(e) => setNewFolderName(e.target.value)}
                                autoFocus
                                required
                                className="w-full bg-bgMain border border-borderDark rounded-md px-3 py-2 text-textMain outline-none focus:border-borderActive font-mono"
                            />
                            <div className="flex items-center justify-end gap-2 pt-2">
                                <button type="button" onClick={() => setShowNewFolderModal(false)} className="px-3 py-1.5 hover:bg-bgHover text-textMuted rounded">Cancel</button>
                                <button type="submit" className="px-3.5 py-1.5 bg-textMain text-bgMain font-semibold rounded hover:opacity-90">Create</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};
