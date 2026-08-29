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
    ArrowUp, 
    Eye, 
    X, 
    Server, 
    AlertCircle,
    ChevronRight,
    Download
} from 'lucide-react';
import { ListRemoteFiles, UploadRemoteFile, ReadRemoteFile, DeleteRemoteFile, CreateRemoteFolder } from '../wailsjs/go/main/App';

interface SidebarSFTPProps {
    hosts: any[];
    activeHostId?: string;
    onOpenFilePreview?: (file: { path: string; name: string; content: string }) => void;
}

export const SidebarSFTP: React.FC<SidebarSFTPProps> = ({ hosts = [], activeHostId, onOpenFilePreview }) => {
    // Determine initial host
    const getHostId = (h: any) => h?.id || h?.ID || '';
    const getHostName = (h: any) => h?.name || h?.Name || 'Server';

    const [selectedHostId, setSelectedHostId] = useState<string>(activeHostId || getHostId(hosts[0]));
    const [currentPath, setCurrentPath] = useState<string>('/root');
    const [pathInput, setPathInput] = useState<string>('/root');
    const [files, setFiles] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const [isDragging, setIsDragging] = useState(false);
    const [uploading, setUploading] = useState(false);

    const [showNewFolder, setShowNewFolder] = useState(false);
    const [newFolderName, setNewFolderName] = useState('');

    // Preview state
    const [previewFile, setPreviewFile] = useState<{ path: string; name: string; content: string } | null>(null);
    const [previewLoading, setPreviewLoading] = useState(false);

    // Sync selected host when active tab changes to an SSH host
    useEffect(() => {
        if (activeHostId) {
            setSelectedHostId(activeHostId);
        } else if (!selectedHostId && hosts.length > 0) {
            setSelectedHostId(getHostId(hosts[0]));
        }
    }, [activeHostId, hosts]);

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
            loadFiles('.').catch(() => loadFiles('/'));
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

    const handleFileClick = async (file: any) => {
        if (file.isDir) {
            handleNavigate(file.path);
        } else {
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
        }
    };

    const handleDelete = async (e: React.MouseEvent, file: any) => {
        e.stopPropagation();
        if (!confirm(`Delete remote ${file.isDir ? 'folder' : 'file'} "${file.name}"?`)) return;
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
            setShowNewFolder(false);
            setNewFolderName('');
            loadFiles(currentPath);
        } catch (err: any) {
            alert(`Folder creation failed: ${err}`);
        }
    };

    // Drag & drop handlers
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
        if (droppedFiles.length === 0 || !selectedHostId) return;

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
        return parseFloat((bytes / Math.pow(k, i)).toFixed(0)) + sizes[i];
    };

    const getFileIcon = (file: any) => {
        if (file.isDir) return <Folder size={14} className="text-amber-400 fill-amber-400/20 shrink-0" />;
        const ext = file.name.split('.').pop()?.toLowerCase();
        if (['js', 'ts', 'tsx', 'jsx', 'py', 'go', 'rs', 'c', 'cpp', 'html', 'css', 'json', 'yaml', 'yml', 'sh'].includes(ext)) {
            return <FileCode size={14} className="text-blue-400 shrink-0" />;
        }
        if (['txt', 'md', 'log', 'conf', 'cfg', 'env'].includes(ext)) {
            return <FileText size={14} className="text-zinc-300 shrink-0" />;
        }
        return <File size={14} className="text-zinc-400 shrink-0" />;
    };

    if (hosts.length === 0) {
        return (
            <div className="p-4 text-center text-xs text-textFaint">
                No SSH endpoints configured yet.
            </div>
        );
    }

    return (
        <div 
            className="flex-1 flex flex-col overflow-hidden relative select-none"
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
        >
            {/* Drag & drop overlay */}
            {isDragging && (
                <div className="absolute inset-0 bg-bgMain/95 backdrop-blur-sm z-50 flex flex-col items-center justify-center border-2 border-dashed border-textMain m-2 rounded-xl pointer-events-none text-center p-3">
                    <Upload size={28} className="text-textMain mb-1.5 animate-bounce" />
                    <span className="text-xs font-semibold text-textMain">Drop files to upload via SFTP</span>
                    <span className="text-[10px] text-textFaint mt-0.5">{currentPath}</span>
                </div>
            )}

            {/* Host Selector & Toolbar */}
            <div className="p-2 border-b border-borderDark/60 bg-bgSidebar space-y-1.5 shrink-0">
                <div className="flex items-center gap-1.5">
                    <Server size={13} className="text-textFaint shrink-0" />
                    <select
                        value={selectedHostId}
                        onChange={(e) => setSelectedHostId(e.target.value)}
                        className="flex-1 bg-bgMain border border-borderDark rounded px-2 py-1 text-xs text-textMain outline-none focus:border-borderActive truncate"
                    >
                        {hosts.map((h, idx) => (
                            <option key={idx} value={getHostId(h)}>
                                {getHostName(h)}
                            </option>
                        ))}
                    </select>
                </div>

                {/* Path navigation bar */}
                <div className="flex items-center gap-1 bg-bgMain border border-borderDark rounded px-1.5 py-0.5">
                    <button 
                        onClick={handleGoUp}
                        className="p-1 hover:bg-bgPanel rounded text-textFaint hover:text-textMain shrink-0"
                        title="Go up"
                    >
                        <ArrowUp size={12} />
                    </button>
                    <form 
                        onSubmit={(e) => { e.preventDefault(); handleNavigate(pathInput); }}
                        className="flex-1 min-w-0"
                    >
                        <input 
                            type="text" 
                            value={pathInput}
                            onChange={(e) => setPathInput(e.target.value)}
                            className="bg-transparent border-none outline-none w-full text-[11px] font-mono text-textMain"
                            placeholder="/root"
                        />
                    </form>
                    <button 
                        onClick={() => setShowNewFolder(!showNewFolder)}
                        className="p-1 hover:bg-bgPanel rounded text-textFaint hover:text-textMain shrink-0"
                        title="New folder"
                    >
                        <FolderPlus size={12} />
                    </button>
                    <button 
                        onClick={() => loadFiles(currentPath)}
                        className="p-1 hover:bg-bgPanel rounded text-textFaint hover:text-textMain shrink-0"
                        title="Refresh"
                    >
                        <RefreshCw size={12} className={loading ? 'animate-spin' : ''} />
                    </button>
                </div>

                {/* New folder input popup */}
                {showNewFolder && (
                    <form onSubmit={handleCreateFolder} className="flex items-center gap-1 mt-1">
                        <input 
                            type="text" 
                            placeholder="Folder name"
                            value={newFolderName}
                            onChange={(e) => setNewFolderName(e.target.value)}
                            autoFocus
                            className="flex-1 bg-bgMain border border-borderDark rounded px-2 py-1 text-xs text-textMain outline-none font-mono"
                        />
                        <button type="submit" className="px-2 py-1 bg-textMain text-bgMain font-semibold text-xs rounded">Add</button>
                        <button type="button" onClick={() => setShowNewFolder(false)} className="p-1 text-textFaint hover:text-textMain"><X size={13} /></button>
                    </form>
                )}
            </div>

            {/* File List in Sidebar */}
            <div className="flex-1 overflow-y-auto py-1">
                {error ? (
                    <div className="p-3 text-[11px] text-rose-400 bg-rose-500/10 border border-rose-500/20 m-2 rounded">
                        SFTP error: {error}
                    </div>
                ) : loading ? (
                    <div className="p-4 text-center text-xs text-textFaint flex items-center justify-center gap-2">
                        <RefreshCw size={13} className="animate-spin" />
                        <span>Loading files...</span>
                    </div>
                ) : files.length === 0 ? (
                    <div className="p-4 text-center text-[11px] text-textFaint">
                        Empty directory. Drag & drop files here to upload.
                    </div>
                ) : (
                    <div className="flex flex-col">
                        {files.map((file, idx) => (
                            <div 
                                key={idx}
                                onClick={() => handleFileClick(file)}
                                className="group px-2.5 py-1.5 flex items-center justify-between text-xs hover:bg-bgPanel cursor-pointer transition-colors"
                            >
                                <div className="flex items-center gap-2 min-w-0">
                                    {getFileIcon(file)}
                                    <span className="truncate text-textMain font-sans text-xs group-hover:text-white">
                                        {file.name}
                                    </span>
                                </div>

                                <div className="flex items-center gap-1.5 shrink-0 ml-2">
                                    <span className="text-[10px] text-textFaint font-mono">
                                        {file.isDir ? '' : formatSize(file.size)}
                                    </span>
                                    <button 
                                        onClick={(e) => handleDelete(e, file)}
                                        className="hidden group-hover:block p-0.5 text-textFaint hover:text-rose-400 rounded"
                                        title="Delete"
                                    >
                                        <Trash2 size={11} />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* File Preview Modal */}
            {previewFile && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-6" onClick={() => setPreviewFile(null)}>
                    <div className="w-full max-w-3xl h-[80vh] bg-bgCard border border-borderDark rounded-xl shadow-2xl flex flex-col overflow-hidden" onClick={(e) => e.stopPropagation()}>
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
        </div>
    );
};
