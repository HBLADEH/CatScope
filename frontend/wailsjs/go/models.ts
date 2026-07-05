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
	export class InstallOptions {
	    allowDowngrade: boolean;
	    grantPermissions: boolean;
	    allowTestOnly: boolean;

	    static createFrom(source: any = {}) {
	        return new InstallOptions(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allowDowngrade = source["allowDowngrade"];
	        this.grantPermissions = source["grantPermissions"];
	        this.allowTestOnly = source["allowTestOnly"];
	    }
	}
	export class InstallResult {
	    success: boolean;
	    apkPath: string;
	    durationMillis: number;
	    output: string;
	    error?: string;
	    analysisResults?: logcat.AnalysisResult[];

	    static createFrom(source: any = {}) {
	        return new InstallResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.apkPath = source["apkPath"];
	        this.durationMillis = source["durationMillis"];
	        this.output = source["output"];
	        this.error = source["error"];
	        this.analysisResults = this.convertValues(source["analysisResults"], logcat.AnalysisResult);
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
	export class LaunchResult {
	    success: boolean;
	    packageName: string;
	    durationMillis: number;
	    output: string;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new LaunchResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.packageName = source["packageName"];
	        this.durationMillis = source["durationMillis"];
	        this.output = source["output"];
	        this.error = source["error"];
	    }
	}

}

export namespace ai {

	export class AIContextOptions {
	    includeDeviceInfo: boolean;
	    includePackageInfo: boolean;
	    includeAnalysisSummary: boolean;
	    includeRelatedLogs: boolean;
	    includeBeforeContextLines: number;
	    includeAfterContextLines: number;
	    includeRawText: boolean;
	    includeSuggestions: boolean;
	    language: string;
	    packageFilter?: string;
	    levelFilter?: string[];
	    searchKeyword?: string;

	    static createFrom(source: any = {}) {
	        return new AIContextOptions(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.includeDeviceInfo = source["includeDeviceInfo"];
	        this.includePackageInfo = source["includePackageInfo"];
	        this.includeAnalysisSummary = source["includeAnalysisSummary"];
	        this.includeRelatedLogs = source["includeRelatedLogs"];
	        this.includeBeforeContextLines = source["includeBeforeContextLines"];
	        this.includeAfterContextLines = source["includeAfterContextLines"];
	        this.includeRawText = source["includeRawText"];
	        this.includeSuggestions = source["includeSuggestions"];
	        this.language = source["language"];
	        this.packageFilter = source["packageFilter"];
	        this.levelFilter = source["levelFilter"];
	        this.searchKeyword = source["searchKeyword"];
	    }
	}

}

export namespace build {

	export class APKInfo {
	    apkPath: string;
	    fileName: string;
	    modifiedTime: string;
	    size: number;

	    static createFrom(source: any = {}) {
	        return new APKInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apkPath = source["apkPath"];
	        this.fileName = source["fileName"];
	        this.modifiedTime = source["modifiedTime"];
	        this.size = source["size"];
	    }
	}
	export class BuildResult {
	    success: boolean;
	    projectPath: string;
	    task: string;
	    durationMillis: number;
	    output: string;
	    error?: string;
	    apk?: APKInfo;

	    static createFrom(source: any = {}) {
	        return new BuildResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.projectPath = source["projectPath"];
	        this.task = source["task"];
	        this.durationMillis = source["durationMillis"];
	        this.output = source["output"];
	        this.error = source["error"];
	        this.apk = this.convertValues(source["apk"], APKInfo);
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

export namespace logcat {

	export class AnalysisResult {
	    id: string;
	    type: string;
	    severity: string;
	    title: string;
	    summary: string;
	    packageName?: string;
	    pid?: number;
	    tid?: number;
	    timestamp?: string;
	    primaryTag?: string;
	    primaryMessage?: string;
	    exceptionType?: string;
	    threadName?: string;
	    signal?: string;
	    libraryName?: string;
	    reason?: string;
	    keyFrames?: string[];
	    relatedEntryIds?: number[];
	    rawText?: string;
	    suggestions?: string[];

	    static createFrom(source: any = {}) {
	        return new AnalysisResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.packageName = source["packageName"];
	        this.pid = source["pid"];
	        this.tid = source["tid"];
	        this.timestamp = source["timestamp"];
	        this.primaryTag = source["primaryTag"];
	        this.primaryMessage = source["primaryMessage"];
	        this.exceptionType = source["exceptionType"];
	        this.threadName = source["threadName"];
	        this.signal = source["signal"];
	        this.libraryName = source["libraryName"];
	        this.reason = source["reason"];
	        this.keyFrames = source["keyFrames"];
	        this.relatedEntryIds = source["relatedEntryIds"];
	        this.rawText = source["rawText"];
	        this.suggestions = source["suggestions"];
	    }
	}
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
	    source: string;
	    offlineFilePath?: string;
	    offlineFileName?: string;
	    offlineParseFailedCount?: number;
	    sessionFilePath?: string;
	    sessionName?: string;

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
	        this.source = source["source"];
	        this.offlineFilePath = source["offlineFilePath"];
	        this.offlineFileName = source["offlineFileName"];
	        this.offlineParseFailedCount = source["offlineParseFailedCount"];
	        this.sessionFilePath = source["sessionFilePath"];
	        this.sessionName = source["sessionName"];
	    }
	}
	export class OfflineLogFileResult {
	    filePath: string;
	    fileName: string;
	    entries: LogEntry[];
	    count: number;
	    parseFailedCount: number;
	    analysisResults?: AnalysisResult[];

	    static createFrom(source: any = {}) {
	        return new OfflineLogFileResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.fileName = source["fileName"];
	        this.entries = this.convertValues(source["entries"], LogEntry);
	        this.count = source["count"];
	        this.parseFailedCount = source["parseFailedCount"];
	        this.analysisResults = this.convertValues(source["analysisResults"], AnalysisResult);
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

export namespace main {

	export class BuildInstallLaunchResult {
	    build: build.BuildResult;
	    install: adb.InstallResult;
	    launch: adb.LaunchResult;
	    packageName: string;
	    apk?: build.APKInfo;
	    analysisResults?: logcat.AnalysisResult[];

	    static createFrom(source: any = {}) {
	        return new BuildInstallLaunchResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.build = this.convertValues(source["build"], build.BuildResult);
	        this.install = this.convertValues(source["install"], adb.InstallResult);
	        this.launch = this.convertValues(source["launch"], adb.LaunchResult);
	        this.packageName = source["packageName"];
	        this.apk = this.convertValues(source["apk"], build.APKInfo);
	        this.analysisResults = this.convertValues(source["analysisResults"], logcat.AnalysisResult);
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
	export class SessionOpenResult {
	    session: storage.Session;
	    summary: storage.SessionSummary;
	    entries: logcat.LogEntry[];
	    analysisResults: logcat.AnalysisResult[];

	    static createFrom(source: any = {}) {
	        return new SessionOpenResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session = this.convertValues(source["session"], storage.Session);
	        this.summary = this.convertValues(source["summary"], storage.SessionSummary);
	        this.entries = this.convertValues(source["entries"], logcat.LogEntry);
	        this.analysisResults = this.convertValues(source["analysisResults"], logcat.AnalysisResult);
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
	export class SessionSaveOptions {
	    name: string;
	    filters: storage.SessionFilters;
	    aiContextOptions: ai.AIContextOptions;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new SessionSaveOptions(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.filters = this.convertValues(source["filters"], storage.SessionFilters);
	        this.aiContextOptions = this.convertValues(source["aiContextOptions"], ai.AIContextOptions);
	        this.notes = source["notes"];
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

export namespace storage {

	export class SessionFilters {
	    level: string[];
	    packageName: string;
	    keyword: string;
	    regexEnabled: boolean;
	    tags: string[];
	    excludeKeyword: string;

	    static createFrom(source: any = {}) {
	        return new SessionFilters(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.packageName = source["packageName"];
	        this.keyword = source["keyword"];
	        this.regexEnabled = source["regexEnabled"];
	        this.tags = source["tags"];
	        this.excludeKeyword = source["excludeKeyword"];
	    }
	}
	export class Session {
	    version: number;
	    sessionId: string;
	    name: string;
	    createdAt: string;
	    updatedAt: string;
	    sourceMode: string;
	    sourceName: string;
	    sourcePath: string;
	    workspaceId: string;
	    workspaceName: string;
	    projectPath: string;
	    packageName: string;
	    selectedDevice: string;
	    knownPids: number[];
	    filters: SessionFilters;
	    aiContextOptions: ai.AIContextOptions;
	    logEntries: logcat.LogEntry[];
	    analysisResults: logcat.AnalysisResult[];
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.sessionId = source["sessionId"];
	        this.name = source["name"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.sourceMode = source["sourceMode"];
	        this.sourceName = source["sourceName"];
	        this.sourcePath = source["sourcePath"];
	        this.workspaceId = source["workspaceId"];
	        this.workspaceName = source["workspaceName"];
	        this.projectPath = source["projectPath"];
	        this.packageName = source["packageName"];
	        this.selectedDevice = source["selectedDevice"];
	        this.knownPids = source["knownPids"];
	        this.filters = this.convertValues(source["filters"], SessionFilters);
	        this.aiContextOptions = this.convertValues(source["aiContextOptions"], ai.AIContextOptions);
	        this.logEntries = this.convertValues(source["logEntries"], logcat.LogEntry);
	        this.analysisResults = this.convertValues(source["analysisResults"], logcat.AnalysisResult);
	        this.notes = source["notes"];
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

	export class SessionSummary {
	    sessionId: string;
	    name: string;
	    filePath: string;
	    createdAt: string;
	    updatedAt: string;
	    sourceMode: string;
	    sourceName: string;
	    workspaceName: string;
	    packageName: string;
	    logCount: number;
	    analysisCount: number;

	    static createFrom(source: any = {}) {
	        return new SessionSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.name = source["name"];
	        this.filePath = source["filePath"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.sourceMode = source["sourceMode"];
	        this.sourceName = source["sourceName"];
	        this.workspaceName = source["workspaceName"];
	        this.packageName = source["packageName"];
	        this.logCount = source["logCount"];
	        this.analysisCount = source["analysisCount"];
	    }
	}

}

export namespace workspace {

	export class FilterPreset {
	    id: string;
	    name: string;
	    level: string[];
	    packageName: string;
	    keyword: string;
	    regexEnabled: boolean;
	    tags: string[];
	    excludeKeyword: string;
	    builtIn?: boolean;

	    static createFrom(source: any = {}) {
	        return new FilterPreset(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.level = source["level"];
	        this.packageName = source["packageName"];
	        this.keyword = source["keyword"];
	        this.regexEnabled = source["regexEnabled"];
	        this.tags = source["tags"];
	        this.excludeKeyword = source["excludeKeyword"];
	        this.builtIn = source["builtIn"];
	    }
	}
	export class WorkspaceConfig {
	    id: string;
	    workspaceName: string;
	    projectPath: string;
	    packageName: string;
	    lastApkPath: string;
	    defaultBuildTask: string;
	    installOptions: adb.InstallOptions;
	    selectedDeviceSerial: string;
	    selectedLogLevel: string[];
	    searchKeyword: string;
	    selectedPackageMode: string;
	    maxLogLines: number;
	    autoStartLogcat: boolean;
	    autoClearOnLaunch: boolean;
	    aiContextOptions: ai.AIContextOptions;

	    static createFrom(source: any = {}) {
	        return new WorkspaceConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceName = source["workspaceName"];
	        this.projectPath = source["projectPath"];
	        this.packageName = source["packageName"];
	        this.lastApkPath = source["lastApkPath"];
	        this.defaultBuildTask = source["defaultBuildTask"];
	        this.installOptions = this.convertValues(source["installOptions"], adb.InstallOptions);
	        this.selectedDeviceSerial = source["selectedDeviceSerial"];
	        this.selectedLogLevel = source["selectedLogLevel"];
	        this.searchKeyword = source["searchKeyword"];
	        this.selectedPackageMode = source["selectedPackageMode"];
	        this.maxLogLines = source["maxLogLines"];
	        this.autoStartLogcat = source["autoStartLogcat"];
	        this.autoClearOnLaunch = source["autoClearOnLaunch"];
	        this.aiContextOptions = this.convertValues(source["aiContextOptions"], ai.AIContextOptions);
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
	export class AppConfig {
	    activeWorkspaceId: string;
	    adbPath?: string;
	    workspaces: WorkspaceConfig[];
	    filterPresets: FilterPreset[];

	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeWorkspaceId = source["activeWorkspaceId"];
	        this.adbPath = source["adbPath"];
	        this.workspaces = this.convertValues(source["workspaces"], WorkspaceConfig);
	        this.filterPresets = this.convertValues(source["filterPresets"], FilterPreset);
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

	export class ProjectConfig {
	    projectPath: string;
	    packageName: string;
	    lastApkPath: string;
	    defaultBuildTask: string;
	    installOptions: adb.InstallOptions;

	    static createFrom(source: any = {}) {
	        return new ProjectConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.packageName = source["packageName"];
	        this.lastApkPath = source["lastApkPath"];
	        this.defaultBuildTask = source["defaultBuildTask"];
	        this.installOptions = this.convertValues(source["installOptions"], adb.InstallOptions);
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
