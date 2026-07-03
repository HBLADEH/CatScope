export namespace adb {
	
	export class AndroidDevice {
	    serial: string;
	    state: string;
	    model?: string;
	    brand?: string;
	    androidVersion?: string;
	    sdkVersion?: string;
	    abi?: string;
	    isEmulator?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AndroidDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serial = source["serial"];
	        this.state = source["state"];
	        this.model = source["model"];
	        this.brand = source["brand"];
	        this.androidVersion = source["androidVersion"];
	        this.sdkVersion = source["sdkVersion"];
	        this.abi = source["abi"];
	        this.isEmulator = source["isEmulator"];
	    }
	}
	export class InstalledPackage {
	    packageName: string;
	    label?: string;
	
	    static createFrom(source: any = {}) {
	        return new InstalledPackage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.packageName = source["packageName"];
	        this.label = source["label"];
	    }
	}

}

export namespace logcat {
	
	export class LogEntry {
	    id: number;
	    timestamp: string;
	    pid: number;
	    tid: number;
	    level: string;
	    tag: string;
	    message: string;
	    packageName?: string;
	    raw: string;
	    multiline?: string[];
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.pid = source["pid"];
	        this.tid = source["tid"];
	        this.level = source["level"];
	        this.tag = source["tag"];
	        this.message = source["message"];
	        this.packageName = source["packageName"];
	        this.raw = source["raw"];
	        this.multiline = source["multiline"];
	    }
	}
	export class LogBatch {
	    entries: LogEntry[];
	    count: number;
	    discardedCount: number;
	    lastID: number;
	
	    static createFrom(source: any = {}) {
	        return new LogBatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entries = this.convertValues(source["entries"], LogEntry);
	        this.count = source["count"];
	        this.discardedCount = source["discardedCount"];
	        this.lastID = source["lastID"];
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
	
	export class LogStatus {
	    running: boolean;
	    serial: string;
	    lastError?: string;
	    count: number;
	    discardedCount: number;
	    lastID: number;
	    adbPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.serial = source["serial"];
	        this.lastError = source["lastError"];
	        this.count = source["count"];
	        this.discardedCount = source["discardedCount"];
	        this.lastID = source["lastID"];
	        this.adbPath = source["adbPath"];
	    }
	}
	export class PackagePIDState {
	    packageName: string;
	    pids: number[];
	    knownPids: number[];
	    lastPid?: number;
	
	    static createFrom(source: any = {}) {
	        return new PackagePIDState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.packageName = source["packageName"];
	        this.pids = source["pids"];
	        this.knownPids = source["knownPids"];
	        this.lastPid = source["lastPid"];
	    }
	}

}

