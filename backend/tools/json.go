package tools

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"easy-text/backend/utils"
)

// JSONError represents a JSON parsing error with position
type JSONError struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Offset  int    `json:"offset"`
}

// JSONResult represents the result of JSON operations
type JSONResult struct {
	Content string     `json:"content,omitempty"`
	Success bool       `json:"success"`
	Error   *JSONError `json:"error,omitempty"`
}

// JSONPathResult represents the result of JSON path query
type JSONPathResult struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

// JSONTool provides JSON manipulation functions
type JSONTool struct{}

// NewJSONTool creates a new JSONTool
func NewJSONTool() *JSONTool {
	return &JSONTool{}
}

// Format formats JSON content with specified indentation
func (jt *JSONTool) Format(content string, indentSize int) *JSONResult {
	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return &JSONResult{
			Success: false,
			Error:   jt.parseError(err, content),
		}
	}

	// Determine indentation
	indent := strings.Repeat(" ", indentSize)
	if indentSize == 0 {
		indent = "\t"
	}

	// Format JSON
	result, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return &JSONResult{
			Success: false,
			Error:   &JSONError{Message: err.Error()},
		}
	}

	return &JSONResult{
		Content: string(result),
		Success: true,
	}
}

// Minify minifies JSON content
func (jt *JSONTool) Minify(content string) *JSONResult {
	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return &JSONResult{
			Success: false,
			Error:   jt.parseError(err, content),
		}
	}

	// Minify JSON
	result, err := json.Marshal(data)
	if err != nil {
		return &JSONResult{
			Success: false,
			Error:   &JSONError{Message: err.Error()},
		}
	}

	return &JSONResult{
		Content: string(result),
		Success: true,
	}
}

// Validate validates JSON content and returns errors
func (jt *JSONTool) Validate(content string) *JSONResult {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return &JSONResult{
			Success: false,
			Error:   jt.parseError(err, content),
		}
	}

	return &JSONResult{
		Success: true,
	}
}

// parseError converts a JSON error to a detailed error
func (jt *JSONTool) parseError(err error, content string) *JSONError {
	// Try to extract position from error message
	jsonErr := &JSONError{
		Message: err.Error(),
		Line:    1,
		Column:  1,
		Offset:  0,
	}

	// Parse error message for position
	// Example: "invalid character 'x' looking for beginning of value"
	// or: "invalid character ',' after object key:value pair at offset 123"

	// Try to find offset in error message
	re := regexp.MustCompile(`offset\s+(\d+)`)
	matches := re.FindStringSubmatch(err.Error())
	if len(matches) > 1 {
		if offset, e := strconv.Atoi(matches[1]); e == nil {
			jsonErr.Offset = offset
			jsonErr.Line, jsonErr.Column = jt.offsetToLineCol(content, offset)
		}
	}

	return jsonErr
}

// offsetToLineCol converts an offset to line and column numbers
func (jt *JSONTool) offsetToLineCol(content string, offset int) (line, column int) {
	line = 1
	column = 1
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return
}

// GetPath gets the JSON path at a given position
func (jt *JSONTool) GetPath(content string, offset int) (string, error) {
	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", err
	}

	// Build path
	path := jt.buildPath(data, offset, "")
	return path, nil
}

// buildPath recursively builds the JSON path
func (jt *JSONTool) buildPath(data interface{}, targetOffset int, currentPath string) string {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPath := currentPath + "." + key
			if newPath[0] == '.' {
				newPath = newPath[1:]
			}
			result := jt.buildPath(val, targetOffset, newPath)
			if result != "" {
				return result
			}
		}
	case []interface{}:
		for i, val := range v {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			result := jt.buildPath(val, targetOffset, newPath)
			if result != "" {
				return result
			}
		}
	}
	return currentPath
}

// JSONToYAML converts JSON to YAML
func (jt *JSONTool) JSONToYAML(content string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", utils.ErrInvalidJSON
	}

	return jt.toYAML(data, 0), nil
}

// toYAML converts data to YAML format
func (jt *JSONTool) toYAML(data interface{}, indent int) string {
	prefix := strings.Repeat("  ", indent)

	switch v := data.(type) {
	case map[string]interface{}:
		var result strings.Builder
		for key, val := range v {
			result.WriteString(prefix + key + ":")
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				result.WriteString("\n" + jt.toYAML(val, indent+1))
			default:
				result.WriteString(" " + jt.toYAML(val, indent+1) + "\n")
			}
		}
		return result.String()
	case []interface{}:
		var result strings.Builder
		for _, item := range v {
			result.WriteString(prefix + "-")
			switch item.(type) {
			case map[string]interface{}, []interface{}:
				result.WriteString("\n" + jt.toYAML(item, indent+1))
			default:
				result.WriteString(" " + jt.toYAML(item, indent+1) + "\n")
			}
		}
		return result.String()
	case string:
		if strings.ContainsAny(v, ":\n\"") {
			return fmt.Sprintf("\"%s\"", strings.ReplaceAll(v, "\"", "\\\""))
		}
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case int, int64:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ExtractKeys extracts all keys from JSON content
func (jt *JSONTool) ExtractKeys(content string) ([]string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, utils.ErrInvalidJSON
	}

	keys := make(map[string]bool)
	jt.extractKeysRecursive(data, "", keys)

	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}
	return result, nil
}

// extractKeysRecursive extracts keys recursively
func (jt *JSONTool) extractKeysRecursive(data interface{}, prefix string, keys map[string]bool) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			keys[fullKey] = true
			jt.extractKeysRecursive(val, fullKey, keys)
		}
	case []interface{}:
		for _, item := range v {
			jt.extractKeysRecursive(item, prefix, keys)
		}
	}
}

// Flatten flattens nested JSON
func (jt *JSONTool) Flatten(content string, separator string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", utils.ErrInvalidJSON
	}

	flattened := make(map[string]interface{})
	jt.flattenRecursive(data, "", separator, flattened)

	result, err := json.MarshalIndent(flattened, "", "  ")
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// flattenRecursive flattens JSON recursively
func (jt *JSONTool) flattenRecursive(data interface{}, prefix string, separator string, result map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newKey := key
			if prefix != "" {
				newKey = prefix + separator + key
			}
			jt.flattenRecursive(val, newKey, separator, result)
		}
	default:
		result[prefix] = v
	}
}

// EscapeString escapes a JSON string
func (jt *JSONTool) EscapeString(s string) string {
	result, _ := json.Marshal(s)
	return string(result)
}

// UnescapeString unescapes a JSON string
func (jt *JSONTool) UnescapeString(s string) (string, error) {
	var result string
	err := json.Unmarshal([]byte(s), &result)
	return result, err
}

// 🆕 V2.0.0 JSON 工具扩展

// QueryPath 执行 JSONPath 查询
func (jt *JSONTool) QueryPath(content string, path string) ([]JSONPathResult, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, utils.ErrInvalidJSON
	}

	results := make([]JSONPathResult, 0)
	jt.queryPathRecursive(data, "$", path, &results)
	return results, nil
}

// queryPathRecursive 递归查询 JSONPath
func (jt *JSONTool) queryPathRecursive(data interface{}, currentPath string, targetPath string, results *[]JSONPathResult) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPath := currentPath + "." + key
			// 检查是否匹配目标路径
			if jt.matchPath(newPath, targetPath) {
				*results = append(*results, JSONPathResult{
					Path:  newPath,
					Value: val,
					Type:  jt.valueType(val),
				})
			}
			jt.queryPathRecursive(val, newPath, targetPath, results)
		}
	case []interface{}:
		for i, val := range v {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			if jt.matchPath(newPath, targetPath) {
				*results = append(*results, JSONPathResult{
					Path:  newPath,
					Value: val,
					Type:  jt.valueType(val),
				})
			}
			jt.queryPathRecursive(val, newPath, targetPath, results)
		}
	}
}

// matchPath 简单路径匹配（支持通配符 *）
func (jt *JSONTool) matchPath(actual, target string) bool {
	// 简化实现：支持 $.store.book[*].author 模式
	// 将通配符 [*] 替换为正则
	pattern := strings.ReplaceAll(target, "$", "\\$")
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "[*]", "\\[\\d+\\]")
	pattern = strings.ReplaceAll(pattern, "*", "[^.\\[\\]]+")
	pattern = "^" + pattern + "$"

	matched, _ := regexp.MatchString(pattern, actual)
	return matched
}

// valueType 返回值的类型字符串
func (jt *JSONTool) valueType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// GenerateStruct 将 JSON 转换为目标语言的结构体定义
func (jt *JSONTool) GenerateStruct(content string, lang string, rootName string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", utils.ErrInvalidJSON
	}

	if rootName == "" {
		rootName = "Root"
	}

	switch lang {
	case "go":
		return jt.generateGoStruct(data, rootName), nil
	case "typescript", "ts":
		return jt.generateTSStruct(data, rootName), nil
	case "java":
		return jt.generateJavaStruct(data, rootName), nil
	case "rust":
		return jt.generateRustStruct(data, rootName), nil
	case "python":
		return jt.generatePythonStruct(data, rootName), nil
	default:
		return jt.generateGoStruct(data, rootName), nil
	}
}

// generateGoStruct 生成 Go 结构体
func (jt *JSONTool) generateGoStruct(data interface{}, name string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))

	jt.writeGoFields(data, &sb, "")

	sb.WriteString("}\n")
	return sb.String()
}

// writeGoFields 递归写入 Go 字段
func (jt *JSONTool) writeGoFields(data interface{}, sb *strings.Builder, indent string) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	for key, val := range m {
		fieldName := jt.toPascalCase(key)
		goType, isStruct := jt.inferGoType(key, val)

		if isStruct {
			// 嵌套结构体
			nestedName := jt.toPascalCase(key)
			sb.WriteString(fmt.Sprintf("%s\t%s %s `json:\"%s\"`\n", indent, fieldName, nestedName, key))
			// 递归生成嵌套结构体
			if nestedMap, ok := val.(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("\n%s// %s represents the nested structure\n", indent, nestedName))
				sb.WriteString(fmt.Sprintf("%stype %s struct {\n", indent, nestedName))
				jt.writeGoFields(nestedMap, sb, indent+"\t")
				sb.WriteString(fmt.Sprintf("%s}\n\n", indent))
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s\t%s %s `json:\"%s\"`\n", indent, fieldName, goType, key))
		}
	}
}

// inferGoType 推断字段的 Go 类型
func (jt *JSONTool) inferGoType(key string, val interface{}) (string, bool) {
	switch v := val.(type) {
	case map[string]interface{}:
		return jt.toPascalCase(key), true
	case []interface{}:
		if len(v) > 0 {
			if _, ok := v[0].(map[string]interface{}); ok {
				return "[]" + jt.toPascalCase(key), true
			}
			elemType, _ := jt.inferGoType("", v[0])
			return "[]" + elemType, false
		}
		return "[]interface{}", false
	case string:
		return "string", false
	case float64:
		if val.(float64) == float64(int64(val.(float64))) {
			return "int64", false
		}
		return "float64", false
	case bool:
		return "bool", false
	case nil:
		return "interface{}", false
	default:
		return "interface{}", false
	}
}

// generateTSStruct 生成 TypeScript 接口
func (jt *JSONTool) generateTSStruct(data interface{}, name string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", name))

	jt.writeTSFields(data, &sb, "  ")

	sb.WriteString("}\n")
	return sb.String()
}

// writeTSFields 递归写入 TypeScript 字段
func (jt *JSONTool) writeTSFields(data interface{}, sb *strings.Builder, indent string) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	for key, val := range m {
		tsType, isNested := jt.inferTSType(key, val)

		if isNested {
			nestedName := jt.toPascalCase(key)
			sb.WriteString(fmt.Sprintf("%s%s: %s;\n", indent, key, nestedName))
			if nestedMap, ok := val.(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("\n%sinterface %s {\n", indent, nestedName))
				jt.writeTSFields(nestedMap, sb, indent+"  ")
				sb.WriteString(fmt.Sprintf("%s}\n\n", indent))
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s%s: %s;\n", indent, key, tsType))
		}
	}
}

// inferTSType 推断 TypeScript 类型
func (jt *JSONTool) inferTSType(key string, val interface{}) (string, bool) {
	switch v := val.(type) {
	case map[string]interface{}:
		return jt.toPascalCase(key), true
	case []interface{}:
		if len(v) > 0 {
			if _, ok := v[0].(map[string]interface{}); ok {
				return jt.toPascalCase(key) + "[]", true
			}
			elemType, _ := jt.inferTSType("", v[0])
			return elemType + "[]", false
		}
		return "any[]", false
	case string:
		return "string", false
	case float64:
		return "number", false
	case bool:
		return "boolean", false
	case nil:
		return "any", false
	default:
		return "any", false
	}
}

// generateJavaStruct 生成 Java 类
func (jt *JSONTool) generateJavaStruct(data interface{}, name string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("public class %s {\n", name))

	jt.writeJavaFields(data, &sb, "    ")

	sb.WriteString("\n")
	jt.writeJavaFields(data, &sb, "    ") // 生成 getter/setter（简化处理）
	sb.WriteString("}\n")
	return sb.String()
}

// writeJavaFields 写入 Java 字段
func (jt *JSONTool) writeJavaFields(data interface{}, sb *strings.Builder, indent string) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	for key, val := range m {
		javaType := jt.inferJavaType(key, val)
		fieldName := jt.toCamelCase(key)
		sb.WriteString(fmt.Sprintf("%sprivate %s %s; // %s\n", indent, javaType, fieldName, key))
	}
}

// inferJavaType 推断 Java 类型
func (jt *JSONTool) inferJavaType(key string, val interface{}) string {
	switch v := val.(type) {
	case map[string]interface{}:
		return "Map<String, Object>"
	case []interface{}:
		if len(v) > 0 {
			if _, ok := v[0].(map[string]interface{}); ok {
				return "List<Map<String, Object>>"
			}
			elemType := jt.inferJavaType("", v[0])
			return "List<" + elemType + ">"
		}
		return "List<Object>"
	case string:
		return "String"
	case float64:
		return "Double"
	case bool:
		return "Boolean"
	case nil:
		return "Object"
	default:
		return "Object"
	}
}

// generateRustStruct 生成 Rust 结构体
func (jt *JSONTool) generateRustStruct(data interface{}, name string) string {
	var sb strings.Builder
	sb.WriteString("#[derive(Debug, serde::Serialize, serde::Deserialize)]\n")
	sb.WriteString(fmt.Sprintf("pub struct %s {\n", name))

	jt.writeRustFields(data, &sb, "    ")

	sb.WriteString("}\n")
	return sb.String()
}

// writeRustFields 写入 Rust 字段
func (jt *JSONTool) writeRustFields(data interface{}, sb *strings.Builder, indent string) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	for key, val := range m {
		rustType := jt.inferRustType(key, val)
		fieldName := jt.toSnakeCase(key)
		sb.WriteString(fmt.Sprintf("%s#[serde(rename = \"%s\")]\n", indent, key))
		sb.WriteString(fmt.Sprintf("%spub %s: %s,\n", indent, fieldName, rustType))
	}
}

// inferRustType 推断 Rust 类型
func (jt *JSONTool) inferRustType(key string, val interface{}) string {
	switch v := val.(type) {
	case map[string]interface{}:
		return "serde_json::Value"
	case []interface{}:
		if len(v) > 0 {
			if _, ok := v[0].(map[string]interface{}); ok {
				return "Vec<serde_json::Value>"
			}
			elemType := jt.inferRustType("", v[0])
			return "Vec<" + elemType + ">"
		}
		return "Vec<serde_json::Value>"
	case string:
		return "String"
	case float64:
		return "f64"
	case bool:
		return "bool"
	case nil:
		return "Option<serde_json::Value>"
	default:
		return "serde_json::Value"
	}
}

// generatePythonStruct 生成 Python dataclass
func (jt *JSONTool) generatePythonStruct(data interface{}, name string) string {
	var sb strings.Builder
	sb.WriteString("from dataclasses import dataclass\n")
	sb.WriteString("from typing import Any, List, Optional\n\n")
	sb.WriteString("@dataclass\n")
	sb.WriteString(fmt.Sprintf("class %s:\n", name))

	jt.writePythonFields(data, &sb, "    ")

	sb.WriteString("\n    @staticmethod\n")
	sb.WriteString("    def from_dict(data: dict) -> '" + name + "':\n")
	sb.WriteString("        return " + name + "(\n")
	m, ok := data.(map[string]interface{})
	if ok {
		for key := range m {
			fieldName := jt.toSnakeCase(key)
			sb.WriteString(fmt.Sprintf("            %s=data.get(\"%s\"),\n", fieldName, key))
		}
	}
	sb.WriteString("        )\n")
	return sb.String()
}

// writePythonFields 写入 Python 字段
func (jt *JSONTool) writePythonFields(data interface{}, sb *strings.Builder, indent string) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	for key, val := range m {
		pyType := jt.inferPythonType(key, val)
		fieldName := jt.toSnakeCase(key)
		sb.WriteString(fmt.Sprintf("%s%s: %s\n", indent, fieldName, pyType))
	}
}

// inferPythonType 推断 Python 类型
func (jt *JSONTool) inferPythonType(key string, val interface{}) string {
	switch v := val.(type) {
	case map[string]interface{}:
		return "dict"
	case []interface{}:
		if len(v) > 0 {
			if _, ok := v[0].(map[string]interface{}); ok {
				return "List[dict]"
			}
			elemType := jt.inferPythonType("", v[0])
			return "List[" + elemType + "]"
		}
		return "List[Any]"
	case string:
		return "str"
	case float64:
		return "float"
	case bool:
		return "bool"
	case nil:
		return "Optional[Any]"
	default:
		return "Any"
	}
}

// toPascalCase 将 snake_case/camelCase/kebab-case 转换为 PascalCase
func (jt *JSONTool) toPascalCase(s string) string {
	// 分割规则
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})

	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return result
}

// toCamelCase 转换为 camelCase
func (jt *JSONTool) toCamelCase(s string) string {
	pascal := jt.toPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(pascal[0:1]) + pascal[1:]
}

// toSnakeCase 转换为 snake_case
func (jt *JSONTool) toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32) // to lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// JSONDiffEntry 结构化 Diff 条目
type JSONDiffEntry struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // "added" | "removed" | "modified" | "unchanged"
	OldValue interface{} `json:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue,omitempty"`
}

// JSONDiffSummary 结构化 Diff 统计
type JSONDiffSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
}

// JSONDiffResult 结构化 Diff 结果
type JSONDiffResult struct {
	Entries []JSONDiffEntry `json:"entries"`
	Summary JSONDiffSummary `json:"summary"`
}

// StructuredDiff 对两个 JSON 进行字段级结构化对比
func (jt *JSONTool) StructuredDiff(left string, right string) (*JSONDiffResult, error) {
	var leftData, rightData interface{}
	if err := json.Unmarshal([]byte(left), &leftData); err != nil {
		return nil, utils.ErrInvalidJSON
	}
	if err := json.Unmarshal([]byte(right), &rightData); err != nil {
		return nil, utils.ErrInvalidJSON
	}

	result := &JSONDiffResult{
		Entries: make([]JSONDiffEntry, 0),
	}
	jt.diffRecursive(leftData, rightData, "$", result)
	return result, nil
}

// diffRecursive 递归对比 JSON
func (jt *JSONTool) diffRecursive(left, right interface{}, path string, result *JSONDiffResult) {
	leftMap, leftOk := left.(map[string]interface{})
	rightMap, rightOk := right.(map[string]interface{})

	if !leftOk || !rightOk {
		// 标量比较
		leftJSON, _ := json.Marshal(left)
		rightJSON, _ := json.Marshal(right)
		if string(leftJSON) != string(rightJSON) {
			result.Entries = append(result.Entries, JSONDiffEntry{
				Path:     path,
				Type:     "modified",
				OldValue: left,
				NewValue: right,
			})
			result.Summary.Modified++
		}
		return
	}

	// 检查左有右无的 key
	allKeys := make(map[string]bool)
	for k := range leftMap {
		allKeys[k] = true
	}
	for k := range rightMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		childPath := path + "." + key
		if path == "$" {
			childPath = "$." + key
		}

		leftVal, lExists := leftMap[key]
		rightVal, rExists := rightMap[key]

		if lExists && !rExists {
			result.Entries = append(result.Entries, JSONDiffEntry{
				Path:     childPath,
				Type:     "removed",
				OldValue: leftVal,
			})
			result.Summary.Removed++
		} else if !lExists && rExists {
			result.Entries = append(result.Entries, JSONDiffEntry{
				Path:     childPath,
				Type:     "added",
				NewValue: rightVal,
			})
			result.Summary.Added++
		} else {
			jt.diffRecursive(leftVal, rightVal, childPath, result)
		}
	}
}
