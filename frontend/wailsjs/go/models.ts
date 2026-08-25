export namespace models {
	
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
	    color: string;
	    x11Forwarding: boolean;
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
	

}

