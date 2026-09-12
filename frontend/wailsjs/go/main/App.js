// @ts-check
// WARNING: 本文件由 scripts/gen-bindings.cjs 生成（或 wails CLI 生成），请勿手改。
export function GetBookmarks(filePath) {
  return window['go']['main']['App']['GetBookmarks'](filePath);
}

export function GetAllBookmarks() {
  return window['go']['main']['App']['GetAllBookmarks']();
}

export function AddBookmark(filePath, lineNumber, note, tag) {
  return window['go']['main']['App']['AddBookmark'](filePath, lineNumber, note, tag);
}

export function RemoveBookmark(id) {
  return window['go']['main']['App']['RemoveBookmark'](id);
}

export function UpdateBookmarkNote(id, note) {
  return window['go']['main']['App']['UpdateBookmarkNote'](id, note);
}

export function UpdateBookmarkTag(id, tag) {
  return window['go']['main']['App']['UpdateBookmarkTag'](id, tag);
}

export function CompareDirectories(leftDir, rightDir) {
  return window['go']['main']['App']['CompareDirectories'](leftDir, rightDir);
}

export function BinaryCompare(leftPath, rightPath) {
  return window['go']['main']['App']['BinaryCompare'](leftPath, rightPath);
}

export function GetConfig() {
  return window['go']['main']['App']['GetConfig']();
}

export function UpdateConfig(newConfig) {
  return window['go']['main']['App']['UpdateConfig'](newConfig);
}

export function SetCloseToTray(enabled) {
  return window['go']['main']['App']['SetCloseToTray'](enabled);
}

export function ForceQuit() {
  return window['go']['main']['App']['ForceQuit']();
}

export function GetSetting(key) {
  return window['go']['main']['App']['GetSetting'](key);
}

export function SetSetting(key, value) {
  return window['go']['main']['App']['SetSetting'](key, value);
}

export function AddFavorite(path) {
  return window['go']['main']['App']['AddFavorite'](path);
}

export function RemoveFavorite(path) {
  return window['go']['main']['App']['RemoveFavorite'](path);
}

export function GetFavorites() {
  return window['go']['main']['App']['GetFavorites']();
}

export function Convert(content, fromFmt, toFmt) {
  return window['go']['main']['App']['Convert'](content, fromFmt, toFmt);
}

export function ShowMessageDialog(title, message, dialogType) {
  return window['go']['main']['App']['ShowMessageDialog'](title, message, dialogType);
}

export function ShowConfirmDialog(title, message) {
  return window['go']['main']['App']['ShowConfirmDialog'](title, message);
}

export function GetAppVersion() {
  return window['go']['main']['App']['GetAppVersion']();
}

export function GetSystemInfo() {
  return window['go']['main']['App']['GetSystemInfo']();
}

export function Exit() {
  return window['go']['main']['App']['Exit']();
}

export function CompareDiff(oldText, newText) {
  return window['go']['main']['App']['CompareDiff'](oldText, newText);
}

export function CompareDiffLines(oldText, newText) {
  return window['go']['main']['App']['CompareDiffLines'](oldText, newText);
}

export function GetDiffPatch(oldText, newText) {
  return window['go']['main']['App']['GetDiffPatch'](oldText, newText);
}

export function ApplyDiffPatch(text, patch) {
  return window['go']['main']['App']['ApplyDiffPatch'](text, patch);
}

export function CompareCharacters(oldStr, newStr) {
  return window['go']['main']['App']['CompareCharacters'](oldStr, newStr);
}

export function OpenDirectoryDialog() {
  return window['go']['main']['App']['OpenDirectoryDialog']();
}

export function GetDirectoryTree(path) {
  return window['go']['main']['App']['GetDirectoryTree'](path);
}

export function CreateDirectory(path) {
  return window['go']['main']['App']['CreateDirectory'](path);
}

export function ListDirectory(path) {
  return window['go']['main']['App']['ListDirectory'](path);
}

export function AutoSaveDraft(filePath, content, encoding, lineEnding) {
  return window['go']['main']['App']['AutoSaveDraft'](filePath, content, encoding, lineEnding);
}

export function GetDraft(filePath) {
  return window['go']['main']['App']['GetDraft'](filePath);
}

export function ListDrafts() {
  return window['go']['main']['App']['ListDrafts']();
}

export function DeleteDraft(filePath) {
  return window['go']['main']['App']['DeleteDraft'](filePath);
}

export function ClearAllDrafts() {
  return window['go']['main']['App']['ClearAllDrafts']();
}

export function CheckDraftConflict(filePath) {
  return window['go']['main']['App']['CheckDraftConflict'](filePath);
}

export function GetSupportedEncodings() {
  return window['go']['main']['App']['GetSupportedEncodings']();
}

export function ConvertEncoding(content, fromEncoding, toEncoding) {
  return window['go']['main']['App']['ConvertEncoding'](content, fromEncoding, toEncoding);
}

export function DetectEncoding(content) {
  return window['go']['main']['App']['DetectEncoding'](content);
}

export function ConvertToUTF8(content, fromEncoding) {
  return window['go']['main']['App']['ConvertToUTF8'](content, fromEncoding);
}

export function ConvertFromUTF8(content, toEncoding) {
  return window['go']['main']['App']['ConvertFromUTF8'](content, toEncoding);
}

export function HasBOM(content) {
  return window['go']['main']['App']['HasBOM'](content);
}

export function RemoveBOM(content) {
  return window['go']['main']['App']['RemoveBOM'](content);
}

export function AddBOM(content, encoding) {
  return window['go']['main']['App']['AddBOM'](content, encoding);
}

export function OpenFileDialog() {
  return window['go']['main']['App']['OpenFileDialog']();
}

export function SaveFileDialog(defaultFilename) {
  return window['go']['main']['App']['SaveFileDialog'](defaultFilename);
}

export function ReadFile(path) {
  return window['go']['main']['App']['ReadFile'](path);
}

export function ReadPartial(path, offset, count) {
  return window['go']['main']['App']['ReadPartial'](path, offset, count);
}

export function SaveFile(path, content, encoding) {
  return window['go']['main']['App']['SaveFile'](path, content, encoding);
}

export function SaveFileWithBackup(path, content, encoding) {
  return window['go']['main']['App']['SaveFileWithBackup'](path, content, encoding);
}

export function GetFileInfo(path) {
  return window['go']['main']['App']['GetFileInfo'](path);
}

export function DeleteFile(path) {
  return window['go']['main']['App']['DeleteFile'](path);
}

export function DeleteDirectory(path) {
  return window['go']['main']['App']['DeleteDirectory'](path);
}

export function RenameFile(oldPath, newPath) {
  return window['go']['main']['App']['RenameFile'](oldPath, newPath);
}

export function CopyFile(src, dst) {
  return window['go']['main']['App']['CopyFile'](src, dst);
}

export function IsBinaryFile(path) {
  return window['go']['main']['App']['IsBinaryFile'](path);
}

export function ReadFileBytes(path) {
  return window['go']['main']['App']['ReadFileBytes'](path);
}

export function ReadFileChunk(path, offset, size) {
  return window['go']['main']['App']['ReadFileChunk'](path, offset, size);
}

export function SaveFileBytes(path, data) {
  return window['go']['main']['App']['SaveFileBytes'](path, data);
}

export function StartFileWatch(path) {
  return window['go']['main']['App']['StartFileWatch'](path);
}

export function StopFileWatch(path) {
  return window['go']['main']['App']['StopFileWatch'](path);
}

export function FindInFile(filePath, search, options) {
  return window['go']['main']['App']['FindInFile'](filePath, search, options);
}

export function FindInDirectory(dirPath, search, options) {
  return window['go']['main']['App']['FindInDirectory'](dirPath, search, options);
}

export function ReplaceInFile(filePath, search, replace, options) {
  return window['go']['main']['App']['ReplaceInFile'](filePath, search, replace, options);
}

export function BatchReplace(dirPath, options) {
  return window['go']['main']['App']['BatchReplace'](dirPath, options);
}

export function SearchInFiles(filePaths, search, options) {
  return window['go']['main']['App']['SearchInFiles'](filePaths, search, options);
}

export function ReplaceInFiles(filePaths, search, replace, options) {
  return window['go']['main']['App']['ReplaceInFiles'](filePaths, search, replace, options);
}

export function ComputeHash(content) {
  return window['go']['main']['App']['ComputeHash'](content);
}

export function ComputeFileHash(filePath) {
  return window['go']['main']['App']['ComputeFileHash'](filePath);
}

export function ComputeHashWithAlgo(content, algorithm) {
  return window['go']['main']['App']['ComputeHashWithAlgo'](content, algorithm);
}

export function ComputeFileHashWithAlgo(filePath, algorithm) {
  return window['go']['main']['App']['ComputeFileHashWithAlgo'](filePath, algorithm);
}

export function FormatJSON(content, indentSize) {
  return window['go']['main']['App']['FormatJSON'](content, indentSize);
}

export function MinifyJSON(content) {
  return window['go']['main']['App']['MinifyJSON'](content);
}

export function ValidateJSON(content) {
  return window['go']['main']['App']['ValidateJSON'](content);
}

export function FlattenJSON(content, separator) {
  return window['go']['main']['App']['FlattenJSON'](content, separator);
}

export function ExtractJSONKeys(content) {
  return window['go']['main']['App']['ExtractJSONKeys'](content);
}

export function JsonPathQuery(jsonStr, path) {
  return window['go']['main']['App']['JsonPathQuery'](jsonStr, path);
}

export function JsonToStruct(jsonStr, lang, rootName) {
  return window['go']['main']['App']['JsonToStruct'](jsonStr, lang, rootName);
}

export function JsonStructuredDiff(leftJSON, rightJSON) {
  return window['go']['main']['App']['JsonStructuredDiff'](leftJSON, rightJSON);
}

export function StartMacroRecording() {
  return window['go']['main']['App']['StartMacroRecording']();
}

export function StopMacroRecording() {
  return window['go']['main']['App']['StopMacroRecording']();
}

export function RecordMacroStep(step) {
  return window['go']['main']['App']['RecordMacroStep'](step);
}

export function GetMacros() {
  return window['go']['main']['App']['GetMacros']();
}

export function DeleteMacro(id) {
  return window['go']['main']['App']['DeleteMacro'](id);
}

export function RenameMacro(id, newName) {
  return window['go']['main']['App']['RenameMacro'](id, newName);
}

export function IsMacroRecording() {
  return window['go']['main']['App']['IsMacroRecording']();
}

export function SaveCurrentMacro(name) {
  return window['go']['main']['App']['SaveCurrentMacro'](name);
}

export function TestRegex(pattern, flags, input) {
  return window['go']['main']['App']['TestRegex'](pattern, flags, input);
}

export function ValidateRegex(pattern) {
  return window['go']['main']['App']['ValidateRegex'](pattern);
}

export function EscapeRegex(input) {
  return window['go']['main']['App']['EscapeRegex'](input);
}

export function BatchRenamePreview(files, pattern, startIndex, step, delimiter) {
  return window['go']['main']['App']['BatchRenamePreview'](files, pattern, startIndex, step, delimiter);
}

export function BatchRenameExecute(preview) {
  return window['go']['main']['App']['BatchRenameExecute'](preview);
}

export function GetCommonNamePatterns() {
  return window['go']['main']['App']['GetCommonNamePatterns']();
}

export function ListScripts() {
  return window['go']['main']['App']['ListScripts']();
}

export function GetScript(id) {
  return window['go']['main']['App']['GetScript'](id);
}

export function SaveScript(script) {
  return window['go']['main']['App']['SaveScript'](script);
}

export function DeleteScript(id) {
  return window['go']['main']['App']['DeleteScript'](id);
}

export function ExecuteScript(id, context) {
  return window['go']['main']['App']['ExecuteScript'](id, context);
}

export function SaveSession(files, activeID) {
  return window['go']['main']['App']['SaveSession'](files, activeID);
}

export function GetSession() {
  return window['go']['main']['App']['GetSession']();
}

export function GetSnippets(language) {
  return window['go']['main']['App']['GetSnippets'](language);
}

export function CreateSnippet(entry) {
  return window['go']['main']['App']['CreateSnippet'](entry);
}

export function UpdateSnippet(entry) {
  return window['go']['main']['App']['UpdateSnippet'](entry);
}

export function DeleteSnippet(id) {
  return window['go']['main']['App']['DeleteSnippet'](id);
}

export function ImportSnippets(jsonData) {
  return window['go']['main']['App']['ImportSnippets'](jsonData);
}

export function ExportSnippets() {
  return window['go']['main']['App']['ExportSnippets']();
}

export function GetRecentFiles() {
  return window['go']['main']['App']['GetRecentFiles']();
}

export function GetRecentFolders() {
  return window['go']['main']['App']['GetRecentFolders']();
}

export function AddRecentEntry(path, isFolder) {
  return window['go']['main']['App']['AddRecentEntry'](path, isFolder);
}

export function ClearRecentFiles() {
  return window['go']['main']['App']['ClearRecentFiles']();
}

export function ClearRecentFolders() {
  return window['go']['main']['App']['ClearRecentFolders']();
}

export function RegisterFileAssoc() {
  return window['go']['main']['App']['RegisterFileAssoc']();
}

export function UnregisterFileAssoc() {
  return window['go']['main']['App']['UnregisterFileAssoc']();
}

export function IsFileAssocRegistered() {
  return window['go']['main']['App']['IsFileAssocRegistered']();
}

