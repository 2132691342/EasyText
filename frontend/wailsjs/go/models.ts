export namespace api {
	
	export class RenameItem {
	    oldPath: string;
	    newName: string;
	    newPath: string;
	    status: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new RenameItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldPath = source["oldPath"];
	        this.newName = source["newName"];
	        this.newPath = source["newPath"];
	        this.status = source["status"];
	        this.error = source["error"];
	    }
	}
	export class BatchRenameResult {
	    items: RenameItem[];
	    success: number;
	    failed: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new BatchRenameResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], RenameItem);
	        this.success = source["success"];
	        this.failed = source["failed"];
	        this.total = source["total"];
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
	export class HashAlgoResult {
	    algorithm: string;
	    hash: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new HashAlgoResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.algorithm = source["algorithm"];
	        this.hash = source["hash"];
	        this.error = source["error"];
	    }
	}
	export class HashResult {
	    md5: string;
	    sha1: string;
	    sha256: string;
	
	    static createFrom(source: any = {}) {
	        return new HashResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.md5 = source["md5"];
	        this.sha1 = source["sha1"];
	        this.sha256 = source["sha256"];
	    }
	}
	
	export class SessionFile {
	    path: string;
	    encoding: string;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.encoding = source["encoding"];
	        this.language = source["language"];
	    }
	}
	export class Session {
	    files: SessionFile[];
	    activeId: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = this.convertValues(source["files"], SessionFile);
	        this.activeId = source["activeId"];
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

export namespace config {
	
	export class ShortcutDef {
	    id: string;
	    name: string;
	    category: string;
	    defaultKey: string;
	    currentKey: string;
	
	    static createFrom(source: any = {}) {
	        return new ShortcutDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.defaultKey = source["defaultKey"];
	        this.currentKey = source["currentKey"];
	    }
	}
	export class UIConfig {
	    language: string;
	    showFileTree: boolean;
	    showStatusBar: boolean;
	    showToolBar: boolean;
	    showFileListView: boolean;
	    showWebAddr: boolean;
	    fileTreeWidth: number;
	    zoomLevel: number;
	    toolbarIconSize: number;
	    favorites: string[];
	    statusBarItems: Record<string, boolean>;
	    toolbarItems: Record<string, boolean>;
	    recentFilesLimit: number;
	    lastFolder: string;
	    closeToTray: boolean;
	    restoreSession?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.showFileTree = source["showFileTree"];
	        this.showStatusBar = source["showStatusBar"];
	        this.showToolBar = source["showToolBar"];
	        this.showFileListView = source["showFileListView"];
	        this.showWebAddr = source["showWebAddr"];
	        this.fileTreeWidth = source["fileTreeWidth"];
	        this.zoomLevel = source["zoomLevel"];
	        this.toolbarIconSize = source["toolbarIconSize"];
	        this.favorites = source["favorites"];
	        this.statusBarItems = source["statusBarItems"];
	        this.toolbarItems = source["toolbarItems"];
	        this.recentFilesLimit = source["recentFilesLimit"];
	        this.lastFolder = source["lastFolder"];
	        this.closeToTray = source["closeToTray"];
	        this.restoreSession = source["restoreSession"];
	    }
	}
	export class FileConfig {
	    defaultEncoding: string;
	    autoDetectEncoding: boolean;
	    defaultLineEnding: string;
	    ignorePatterns: string[];
	
	    static createFrom(source: any = {}) {
	        return new FileConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultEncoding = source["defaultEncoding"];
	        this.autoDetectEncoding = source["autoDetectEncoding"];
	        this.defaultLineEnding = source["defaultLineEnding"];
	        this.ignorePatterns = source["ignorePatterns"];
	    }
	}
	export class ThemeConfig {
	    currentTheme: string;
	    autoTheme: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ThemeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentTheme = source["currentTheme"];
	        this.autoTheme = source["autoTheme"];
	    }
	}
	export class ColumnModeConfig {
	    numberStart: number;
	    numberStep: number;
	    numberBase: number;
	    dateFormat: string;
	    caseConversion: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnModeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.numberStart = source["numberStart"];
	        this.numberStep = source["numberStep"];
	        this.numberBase = source["numberBase"];
	        this.dateFormat = source["dateFormat"];
	        this.caseConversion = source["caseConversion"];
	    }
	}
	export class EditorConfig {
	    fontSize: number;
	    fontFamily: string;
	    tabSize: number;
	    insertSpaces: boolean;
	    wordWrap: boolean;
	    lineNumbers: boolean;
	    autoSave: boolean;
	    autoSaveInterval: number;
	    autoSaveMode: string;
	    highlightLine: boolean;
	    bracketPairColor: boolean;
	    minimap: boolean;
	    showIndentGuide: boolean;
	    showWhitespace: boolean;
	    showEol: boolean;
	    foldEnable: boolean;
	    columnMode: boolean;
	    columnModeConfig: ColumnModeConfig;
	
	    static createFrom(source: any = {}) {
	        return new EditorConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fontSize = source["fontSize"];
	        this.fontFamily = source["fontFamily"];
	        this.tabSize = source["tabSize"];
	        this.insertSpaces = source["insertSpaces"];
	        this.wordWrap = source["wordWrap"];
	        this.lineNumbers = source["lineNumbers"];
	        this.autoSave = source["autoSave"];
	        this.autoSaveInterval = source["autoSaveInterval"];
	        this.autoSaveMode = source["autoSaveMode"];
	        this.highlightLine = source["highlightLine"];
	        this.bracketPairColor = source["bracketPairColor"];
	        this.minimap = source["minimap"];
	        this.showIndentGuide = source["showIndentGuide"];
	        this.showWhitespace = source["showWhitespace"];
	        this.showEol = source["showEol"];
	        this.foldEnable = source["foldEnable"];
	        this.columnMode = source["columnMode"];
	        this.columnModeConfig = this.convertValues(source["columnModeConfig"], ColumnModeConfig);
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
	    version: number;
	    editor: EditorConfig;
	    theme: ThemeConfig;
	    file: FileConfig;
	    ui: UIConfig;
	    shortcuts?: ShortcutDef[];
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.editor = this.convertValues(source["editor"], EditorConfig);
	        this.theme = this.convertValues(source["theme"], ThemeConfig);
	        this.file = this.convertValues(source["file"], FileConfig);
	        this.ui = this.convertValues(source["ui"], UIConfig);
	        this.shortcuts = this.convertValues(source["shortcuts"], ShortcutDef);
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

export namespace file {
	
	export class ChunkResult {
	    data: number[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new ChunkResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.total = source["total"];
	    }
	}
	export class FileInfo {
	    path: string;
	    name: string;
	    ext: string;
	    size: number;
	    modified: string;
	    encoding: string;
	    lineEnding: string;
	    lineCount: number;
	    isReadOnly: boolean;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.ext = source["ext"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	        this.encoding = source["encoding"];
	        this.lineEnding = source["lineEnding"];
	        this.lineCount = source["lineCount"];
	        this.isReadOnly = source["isReadOnly"];
	        this.isDir = source["isDir"];
	    }
	}
	export class TreeNode {
	    path: string;
	    name: string;
	    isDir: boolean;
	    children?: TreeNode[];
	    expanded: boolean;
	    ext?: string;
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.isDir = source["isDir"];
	        this.children = this.convertValues(source["children"], TreeNode);
	        this.expanded = source["expanded"];
	        this.ext = source["ext"];
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
	export class FileTree {
	    root?: TreeNode;
	    basePath: string;
	
	    static createFrom(source: any = {}) {
	        return new FileTree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = this.convertValues(source["root"], TreeNode);
	        this.basePath = source["basePath"];
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
	export class ReadResult {
	    content: string;
	    info: FileInfo;
	    detectedEncoding: string;
	
	    static createFrom(source: any = {}) {
	        return new ReadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.info = this.convertValues(source["info"], FileInfo);
	        this.detectedEncoding = source["detectedEncoding"];
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
	
	export class WriteResult {
	    path: string;
	    size: number;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WriteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.success = source["success"];
	    }
	}

}

export namespace tools {
	
	export class BinCompareResult {
	    equal: boolean;
	    firstDiffOffset: number;
	    reason: string;
	    hexWindow: string;
	
	    static createFrom(source: any = {}) {
	        return new BinCompareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.equal = source["equal"];
	        this.firstDiffOffset = source["firstDiffOffset"];
	        this.reason = source["reason"];
	        this.hexWindow = source["hexWindow"];
	    }
	}
	export class BookmarkEntry {
	    id: number;
	    filePath: string;
	    lineNumber: number;
	    note: string;
	    tag: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BookmarkEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filePath = source["filePath"];
	        this.lineNumber = source["lineNumber"];
	        this.note = source["note"];
	        this.tag = source["tag"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class DiffLine {
	    type: string;
	    content: string;
	    oldLine: number;
	    newLine: number;
	    diffChars?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	        this.oldLine = source["oldLine"];
	        this.newLine = source["newLine"];
	        this.diffChars = source["diffChars"];
	    }
	}
	export class DiffBlock {
	    oldStart: number;
	    oldCount: number;
	    newStart: number;
	    newCount: number;
	    lines: DiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new DiffBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldStart = source["oldStart"];
	        this.oldCount = source["oldCount"];
	        this.newStart = source["newStart"];
	        this.newCount = source["newCount"];
	        this.lines = this.convertValues(source["lines"], DiffLine);
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
	export class DiffItem {
	    type: string;
	    content: string;
	    oldLine?: number;
	    newLine?: number;
	    position: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	        this.oldLine = source["oldLine"];
	        this.newLine = source["newLine"];
	        this.position = source["position"];
	    }
	}
	
	export class DiffResult {
	    diffs: DiffItem[];
	    added: number;
	    removed: number;
	    modified: number;
	    unchanged: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.diffs = this.convertValues(source["diffs"], DiffItem);
	        this.added = source["added"];
	        this.removed = source["removed"];
	        this.modified = source["modified"];
	        this.unchanged = source["unchanged"];
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
	export class DirCompareEntry {
	    relPath: string;
	    leftOnly: boolean;
	    rightOnly: boolean;
	    identical: boolean;
	    different: boolean;
	    leftSize: number;
	    rightSize: number;
	    leftMtime: string;
	    rightMtime: string;
	
	    static createFrom(source: any = {}) {
	        return new DirCompareEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relPath = source["relPath"];
	        this.leftOnly = source["leftOnly"];
	        this.rightOnly = source["rightOnly"];
	        this.identical = source["identical"];
	        this.different = source["different"];
	        this.leftSize = source["leftSize"];
	        this.rightSize = source["rightSize"];
	        this.leftMtime = source["leftMtime"];
	        this.rightMtime = source["rightMtime"];
	    }
	}
	export class DirCompareResult {
	    leftBase: string;
	    rightBase: string;
	    entries: DirCompareEntry[];
	
	    static createFrom(source: any = {}) {
	        return new DirCompareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.leftBase = source["leftBase"];
	        this.rightBase = source["rightBase"];
	        this.entries = this.convertValues(source["entries"], DirCompareEntry);
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
	export class DraftEntry {
	    id: number;
	    filePath: string;
	    content: string;
	    encoding: string;
	    lineEnding: string;
	    savedAt: string;
	    fileModtime: number;
	
	    static createFrom(source: any = {}) {
	        return new DraftEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filePath = source["filePath"];
	        this.content = source["content"];
	        this.encoding = source["encoding"];
	        this.lineEnding = source["lineEnding"];
	        this.savedAt = source["savedAt"];
	        this.fileModtime = source["fileModtime"];
	    }
	}
	export class EncodingInfo {
	    name: string;
	    displayName: string;
	
	    static createFrom(source: any = {}) {
	        return new EncodingInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	    }
	}
	export class FindMatch {
	    index: number;
	    file: string;
	    line: number;
	    column: number;
	    content: string;
	    matchText: string;
	    matchLength: number;
	
	    static createFrom(source: any = {}) {
	        return new FindMatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.file = source["file"];
	        this.line = source["line"];
	        this.column = source["column"];
	        this.content = source["content"];
	        this.matchText = source["matchText"];
	        this.matchLength = source["matchLength"];
	    }
	}
	export class FindInFileResult {
	    file: string;
	    matches: FindMatch[];
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new FindInFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.matches = this.convertValues(source["matches"], FindMatch);
	        this.count = source["count"];
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
	
	export class FindOptions {
	    search: string;
	    replace: string;
	    caseSensitive: boolean;
	    wholeWord: boolean;
	    useRegex: boolean;
	    includeSubdir: boolean;
	    filePattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FindOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.search = source["search"];
	        this.replace = source["replace"];
	        this.caseSensitive = source["caseSensitive"];
	        this.wholeWord = source["wholeWord"];
	        this.useRegex = source["useRegex"];
	        this.includeSubdir = source["includeSubdir"];
	        this.filePattern = source["filePattern"];
	    }
	}
	export class JSONDiffEntry {
	    path: string;
	    type: string;
	    oldValue?: any;
	    newValue?: any;
	
	    static createFrom(source: any = {}) {
	        return new JSONDiffEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.type = source["type"];
	        this.oldValue = source["oldValue"];
	        this.newValue = source["newValue"];
	    }
	}
	export class JSONDiffSummary {
	    added: number;
	    removed: number;
	    modified: number;
	
	    static createFrom(source: any = {}) {
	        return new JSONDiffSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.added = source["added"];
	        this.removed = source["removed"];
	        this.modified = source["modified"];
	    }
	}
	export class JSONDiffResult {
	    entries: JSONDiffEntry[];
	    summary: JSONDiffSummary;
	
	    static createFrom(source: any = {}) {
	        return new JSONDiffResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entries = this.convertValues(source["entries"], JSONDiffEntry);
	        this.summary = this.convertValues(source["summary"], JSONDiffSummary);
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
	
	export class JSONError {
	    message: string;
	    line: number;
	    column: number;
	    offset: number;
	
	    static createFrom(source: any = {}) {
	        return new JSONError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.line = source["line"];
	        this.column = source["column"];
	        this.offset = source["offset"];
	    }
	}
	export class JSONPathResult {
	    path: string;
	    value: any;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new JSONPathResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.value = source["value"];
	        this.type = source["type"];
	    }
	}
	export class JSONResult {
	    content?: string;
	    success: boolean;
	    error?: JSONError;
	
	    static createFrom(source: any = {}) {
	        return new JSONResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.success = source["success"];
	        this.error = this.convertValues(source["error"], JSONError);
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
	export class MacroStep {
	    type: string;
	    text?: string;
	    from?: number;
	    to?: number;
	    anchor?: number;
	    head?: number;
	    timestamp: number;
	    command?: string;
	
	    static createFrom(source: any = {}) {
	        return new MacroStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.text = source["text"];
	        this.from = source["from"];
	        this.to = source["to"];
	        this.anchor = source["anchor"];
	        this.head = source["head"];
	        this.timestamp = source["timestamp"];
	        this.command = source["command"];
	    }
	}
	export class Macro {
	    id: string;
	    name: string;
	    steps: MacroStep[];
	    createdAt: number;
	    modifiedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Macro(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.steps = this.convertValues(source["steps"], MacroStep);
	        this.createdAt = source["createdAt"];
	        this.modifiedAt = source["modifiedAt"];
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
	
	export class RecentEntryResult {
	    path: string;
	    isFolder: boolean;
	    name: string;
	    accessedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new RecentEntryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.isFolder = source["isFolder"];
	        this.name = source["name"];
	        this.accessedAt = source["accessedAt"];
	    }
	}
	export class RegexGroup {
	    index: number;
	    name: string;
	    value: string;
	    start: number;
	    end: number;
	
	    static createFrom(source: any = {}) {
	        return new RegexGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.value = source["value"];
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}
	export class RegexMatchDetail {
	    index: number;
	    value: string;
	    start: number;
	    end: number;
	    groups: RegexGroup[];
	
	    static createFrom(source: any = {}) {
	        return new RegexMatchDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.value = source["value"];
	        this.start = source["start"];
	        this.end = source["end"];
	        this.groups = this.convertValues(source["groups"], RegexGroup);
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
	export class RegexTestResult {
	    pattern: string;
	    flags: string;
	    input: string;
	    matches: RegexMatchDetail[];
	    count: number;
	    error?: string;
	    valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RegexTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pattern = source["pattern"];
	        this.flags = source["flags"];
	        this.input = source["input"];
	        this.matches = this.convertValues(source["matches"], RegexMatchDetail);
	        this.count = source["count"];
	        this.error = source["error"];
	        this.valid = source["valid"];
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
	export class ScriptContext {
	    filePath: string;
	    content: string;
	    selection: string;
	    cursorLine: number;
	    cursorCol: number;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.content = source["content"];
	        this.selection = source["selection"];
	        this.cursorLine = source["cursorLine"];
	        this.cursorCol = source["cursorCol"];
	        this.language = source["language"];
	    }
	}
	export class ScriptInfo {
	    id: string;
	    name: string;
	    description: string;
	    language: string;
	    code: string;
	    enabled: boolean;
	    menuGroup: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.enabled = source["enabled"];
	        this.menuGroup = source["menuGroup"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class ScriptResult {
	    success: boolean;
	    output: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.output = source["output"];
	        this.error = source["error"];
	    }
	}
	export class SnippetEntry {
	    id: number;
	    name: string;
	    prefix: string;
	    body: string;
	    description: string;
	    language: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SnippetEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.prefix = source["prefix"];
	        this.body = source["body"];
	        this.description = source["description"];
	        this.language = source["language"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

