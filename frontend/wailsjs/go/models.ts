export namespace diagnostics {
	
	export class DiagnosticsResult {
	    target: string;
	    port: number;
	    success: boolean;
	    latencyMs: number;
	    ips?: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.port = source["port"];
	        this.success = source["success"];
	        this.latencyMs = source["latencyMs"];
	        this.ips = source["ips"];
	        this.message = source["message"];
	    }
	}

}

export namespace discovery {
	
	export class DetectedTool {
	    name: string;
	    installed: boolean;
	    path: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new DetectedTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.installed = source["installed"];
	        this.path = source["path"];
	        this.version = source["version"];
	    }
	}
	export class DockerContainerInfo {
	    id: string;
	    image: string;
	    command: string;
	    createdAt: string;
	    state: string;
	    status: string;
	    names: string;
	    ports: string;
	
	    static createFrom(source: any = {}) {
	        return new DockerContainerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.image = source["image"];
	        this.command = source["command"];
	        this.createdAt = source["createdAt"];
	        this.state = source["state"];
	        this.status = source["status"];
	        this.names = source["names"];
	        this.ports = source["ports"];
	    }
	}

}

export namespace gitops {
	
	export class SyncResult {
	    success: boolean;
	    message: string;
	    // Go type: time
	    lastSyncedAt: any;
	    commitHash?: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.lastSyncedAt = this.convertValues(source["lastSyncedAt"], null);
	        this.commitHash = source["commitHash"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace importers {
	
	export class ImportResult {
	    importedCount: number;
	    hosts: models.Host[];
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.importedCount = source["importedCount"];
	        this.hosts = this.convertValues(source["hosts"], models.Host);
	        this.errors = source["errors"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace models {
	
	export class Connection {
	    id: string;
	    hostId: string;
	    name: string;
	    type: string;
	    credentialId?: string;
	    port: number;
	    target?: string;
	    params?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Connection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hostId = source["hostId"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.credentialId = source["credentialId"];
	        this.port = source["port"];
	        this.target = source["target"];
	        this.params = source["params"];
	    }
	}
	export class DiscoveredDevice {
	    ip: string;
	    hostname?: string;
	    openPorts: number[];
	    services: string[];
	    latencyMs: number;
	    vendor?: string;
	    matchedProto?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.hostname = source["hostname"];
	        this.openPorts = source["openPorts"];
	        this.services = source["services"];
	        this.latencyMs = source["latencyMs"];
	        this.vendor = source["vendor"];
	        this.matchedProto = source["matchedProto"];
	    }
	}
	export class GitOpsConfig {
	    repoUrl: string;
	    branch: string;
	    authType: string;
	    sshKeyPath?: string;
	    accessToken?: string;
	    autoSync: boolean;
	    encryptSecret: boolean;
	    encryptionKey?: string;
	    // Go type: time
	    lastSynced: any;
	
	    static createFrom(source: any = {}) {
	        return new GitOpsConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repoUrl = source["repoUrl"];
	        this.branch = source["branch"];
	        this.authType = source["authType"];
	        this.sshKeyPath = source["sshKeyPath"];
	        this.accessToken = source["accessToken"];
	        this.autoSync = source["autoSync"];
	        this.encryptSecret = source["encryptSecret"];
	        this.encryptionKey = source["encryptionKey"];
	        this.lastSynced = this.convertValues(source["lastSynced"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PortForwardRule {
	    id: string;
	    hostId: string;
	    name: string;
	    type: string;
	    bindAddress: string;
	    bindPort: number;
	    targetAddress?: string;
	    targetPort?: number;
	    autoStart: boolean;
	    active: boolean;
	    rxBytes: number;
	    txBytes: number;
	    activeConns: number;
	    errorMessage?: string;
	
	    static createFrom(source: any = {}) {
	        return new PortForwardRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hostId = source["hostId"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.bindAddress = source["bindAddress"];
	        this.bindPort = source["bindPort"];
	        this.targetAddress = source["targetAddress"];
	        this.targetPort = source["targetPort"];
	        this.autoStart = source["autoStart"];
	        this.active = source["active"];
	        this.rxBytes = source["rxBytes"];
	        this.txBytes = source["txBytes"];
	        this.activeConns = source["activeConns"];
	        this.errorMessage = source["errorMessage"];
	    }
	}
	export class RemoteService {
	    id: string;
	    hostId: string;
	    name: string;
	    type: string;
	    remoteHost: string;
	    remotePort: number;
	    localPort?: number;
	    autoTunnel: boolean;
	    path?: string;
	    icon?: string;
	
	    static createFrom(source: any = {}) {
	        return new RemoteService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hostId = source["hostId"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.remoteHost = source["remoteHost"];
	        this.remotePort = source["remotePort"];
	        this.localPort = source["localPort"];
	        this.autoTunnel = source["autoTunnel"];
	        this.path = source["path"];
	        this.icon = source["icon"];
	    }
	}
	export class JumpHostHop {
	    hopIndex: number;
	    hostId?: string;
	    hostname: string;
	    port: number;
	    username: string;
	    authMethod: string;
	    password?: string;
	    privateKeyPath?: string;
	    privateKeyData?: string;
	    keyPassphrase?: string;
	    hardwareKeySlot?: string;
	
	    static createFrom(source: any = {}) {
	        return new JumpHostHop(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hopIndex = source["hopIndex"];
	        this.hostId = source["hostId"];
	        this.hostname = source["hostname"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.authMethod = source["authMethod"];
	        this.password = source["password"];
	        this.privateKeyPath = source["privateKeyPath"];
	        this.privateKeyData = source["privateKeyData"];
	        this.keyPassphrase = source["keyPassphrase"];
	        this.hardwareKeySlot = source["hardwareKeySlot"];
	    }
	}
	export class Host {
	    id: string;
	    name: string;
	    hostname: string;
	    port: number;
	    protocol: string;
	    username: string;
	    credentialId?: string;
	    authMethod: string;
	    password?: string;
	    privateKeyPath?: string;
	    privateKeyData?: string;
	    keyPassphrase?: string;
	    certPath?: string;
	    jumpChain?: JumpHostHop[];
	    environment: string;
	    folder: string;
	    tags: string[];
	    color?: string;
	    x11Forwarding?: boolean;
	    connections?: Connection[];
	    services?: RemoteService[];
	    forwardings?: PortForwardRule[];
	    autoCommands?: string[];
	    snippetIds?: string[];
	    dockerEndpoint?: string;
	    k8sNamespace?: string;
	    health: string;
	    latencyMs: number;
	    // Go type: time
	    lastSeen?: any;
	    notes?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Host(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.hostname = source["hostname"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.username = source["username"];
	        this.credentialId = source["credentialId"];
	        this.authMethod = source["authMethod"];
	        this.password = source["password"];
	        this.privateKeyPath = source["privateKeyPath"];
	        this.privateKeyData = source["privateKeyData"];
	        this.keyPassphrase = source["keyPassphrase"];
	        this.certPath = source["certPath"];
	        this.jumpChain = this.convertValues(source["jumpChain"], JumpHostHop);
	        this.environment = source["environment"];
	        this.folder = source["folder"];
	        this.tags = source["tags"];
	        this.color = source["color"];
	        this.x11Forwarding = source["x11Forwarding"];
	        this.connections = this.convertValues(source["connections"], Connection);
	        this.services = this.convertValues(source["services"], RemoteService);
	        this.forwardings = this.convertValues(source["forwardings"], PortForwardRule);
	        this.autoCommands = source["autoCommands"];
	        this.snippetIds = source["snippetIds"];
	        this.dockerEndpoint = source["dockerEndpoint"];
	        this.k8sNamespace = source["k8sNamespace"];
	        this.health = source["health"];
	        this.latencyMs = source["latencyMs"];
	        this.lastSeen = this.convertValues(source["lastSeen"], null);
	        this.notes = source["notes"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ResourceCapabilities {
	    canConnect: boolean;
	    canOpenTerminal: boolean;
	    canBrowseFiles: boolean;
	    canOpenLogs: boolean;
	    canStart: boolean;
	    canStop: boolean;
	    canRestart: boolean;
	    canInspect: boolean;
	    canCreateTunnel: boolean;
	    canOpenService: boolean;
	    canDelete: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ResourceCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.canConnect = source["canConnect"];
	        this.canOpenTerminal = source["canOpenTerminal"];
	        this.canBrowseFiles = source["canBrowseFiles"];
	        this.canOpenLogs = source["canOpenLogs"];
	        this.canStart = source["canStart"];
	        this.canStop = source["canStop"];
	        this.canRestart = source["canRestart"];
	        this.canInspect = source["canInspect"];
	        this.canCreateTunnel = source["canCreateTunnel"];
	        this.canOpenService = source["canOpenService"];
	        this.canDelete = source["canDelete"];
	    }
	}
	export class InfrastructureNode {
	    id: string;
	    parentId?: string;
	    nodeType: string;
	    providerId?: string;
	    resourceId?: string;
	    hostId?: string;
	    connectionId?: string;
	    serviceId?: string;
	    name: string;
	    alias?: string;
	    status: string;
	    icon?: string;
	    capabilities: ResourceCapabilities;
	    children?: InfrastructureNode[];
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new InfrastructureNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.parentId = source["parentId"];
	        this.nodeType = source["nodeType"];
	        this.providerId = source["providerId"];
	        this.resourceId = source["resourceId"];
	        this.hostId = source["hostId"];
	        this.connectionId = source["connectionId"];
	        this.serviceId = source["serviceId"];
	        this.name = source["name"];
	        this.alias = source["alias"];
	        this.status = source["status"];
	        this.icon = source["icon"];
	        this.capabilities = this.convertValues(source["capabilities"], ResourceCapabilities);
	        this.children = this.convertValues(source["children"], InfrastructureNode);
	        this.metadata = source["metadata"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class KnownHostRecord {
	    hostname: string;
	    port: number;
	    keyType: string;
	    fingerprint: string;
	    hostKeyRaw: string;
	    // Go type: time
	    firstSeen: any;
	    // Go type: time
	    lastSeen: any;
	    trusted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new KnownHostRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.port = source["port"];
	        this.keyType = source["keyType"];
	        this.fingerprint = source["fingerprint"];
	        this.hostKeyRaw = source["hostKeyRaw"];
	        this.firstSeen = this.convertValues(source["firstSeen"], null);
	        this.lastSeen = this.convertValues(source["lastSeen"], null);
	        this.trusted = source["trusted"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Resource {
	    id: string;
	    providerId: string;
	    type: string;
	    name: string;
	    parentId?: string;
	    folder?: string;
	    status: string;
	    capabilities: ResourceCapabilities;
	    connections?: Connection[];
	    services?: RemoteService[];
	    tags?: string[];
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Resource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.providerId = source["providerId"];
	        this.type = source["type"];
	        this.name = source["name"];
	        this.parentId = source["parentId"];
	        this.folder = source["folder"];
	        this.status = source["status"];
	        this.capabilities = this.convertValues(source["capabilities"], ResourceCapabilities);
	        this.connections = this.convertValues(source["connections"], Connection);
	        this.services = this.convertValues(source["services"], RemoteService);
	        this.tags = source["tags"];
	        this.metadata = source["metadata"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProviderDiscoveryResult {
	    providerId: string;
	    name: string;
	    status: string;
	    resources: Resource[];
	    error?: string;
	    // Go type: time
	    lastSync: any;
	
	    static createFrom(source: any = {}) {
	        return new ProviderDiscoveryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerId = source["providerId"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.resources = this.convertValues(source["resources"], Resource);
	        this.error = source["error"];
	        this.lastSync = this.convertValues(source["lastSync"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class Session {
	    id: string;
	    hostId: string;
	    connectionId?: string;
	    title: string;
	    state: string;
	    cols: number;
	    rows: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    lastActiveAt: any;
	    errorMessage?: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hostId = source["hostId"];
	        this.connectionId = source["connectionId"];
	        this.title = source["title"];
	        this.state = source["state"];
	        this.cols = source["cols"];
	        this.rows = source["rows"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.lastActiveAt = this.convertValues(source["lastActiveAt"], null);
	        this.errorMessage = source["errorMessage"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Snippet {
	    id: string;
	    title: string;
	    description: string;
	    command: string;
	    tags: string[];
	    variables?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Snippet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.command = source["command"];
	        this.tags = source["tags"];
	        this.variables = source["variables"];
	    }
	}

}

export namespace plugins {
	
	export class PluginCommand {
	    id: string;
	    title: string;
	    category?: string;
	
	    static createFrom(source: any = {}) {
	        return new PluginCommand(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.category = source["category"];
	    }
	}
	export class PluginView {
	    id: string;
	    name: string;
	    icon?: string;
	
	    static createFrom(source: any = {}) {
	        return new PluginView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	    }
	}
	export class PluginContribution {
	    commands?: PluginCommand[];
	    views?: PluginView[];
	
	    static createFrom(source: any = {}) {
	        return new PluginContribution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commands = this.convertValues(source["commands"], PluginCommand);
	        this.views = this.convertValues(source["views"], PluginView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PluginManifest {
	    id: string;
	    name: string;
	    displayName: string;
	    version: string;
	    publisher: string;
	    description: string;
	    icon?: string;
	    enabled: boolean;
	    // Go type: time
	    installedAt: any;
	    permissions: string[];
	    contributes: PluginContribution;
	
	    static createFrom(source: any = {}) {
	        return new PluginManifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.version = source["version"];
	        this.publisher = source["publisher"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.enabled = source["enabled"];
	        this.installedAt = this.convertValues(source["installedAt"], null);
	        this.permissions = source["permissions"];
	        this.contributes = this.convertValues(source["contributes"], PluginContribution);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace services {
	
	export class ActiveServiceStatus {
	    service: models.RemoteService;
	    localUrl: string;
	    tunnelId: string;
	    isRunning: boolean;
	    healthy: boolean;
	    statusMsg: string;
	
	    static createFrom(source: any = {}) {
	        return new ActiveServiceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = this.convertValues(source["service"], models.RemoteService);
	        this.localUrl = source["localUrl"];
	        this.tunnelId = source["tunnelId"];
	        this.isRunning = source["isRunning"];
	        this.healthy = source["healthy"];
	        this.statusMsg = source["statusMsg"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace ssh {
	
	export class RemoteFileInfo {
	    name: string;
	    path: string;
	    size: number;
	    mode: string;
	    isDir: boolean;
	    modTime: string;
	
	    static createFrom(source: any = {}) {
	        return new RemoteFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.mode = source["mode"];
	        this.isDir = source["isDir"];
	        this.modTime = source["modTime"];
	    }
	}

}

