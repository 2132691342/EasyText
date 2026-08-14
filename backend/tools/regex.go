package tools

import (
	"regexp"
	"strings"

	"easy-text/backend/utils"
)

// RegexGroup 正则匹配的分组信息
type RegexGroup struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// RegexMatchDetail 单次正则匹配的详细信息
type RegexMatchDetail struct {
	Index  int          `json:"index"`
	Value  string       `json:"value"`
	Start  int          `json:"start"`
	End    int          `json:"end"`
	Groups []RegexGroup `json:"groups"`
}

// RegexTestResult 正则测试结果
type RegexTestResult struct {
	Pattern string             `json:"pattern"`
	Flags   string             `json:"flags"` // "g" | "gi" | "gm" | "gim" | ""
	Input   string             `json:"input"`
	Matches []RegexMatchDetail `json:"matches"`
	Count   int                `json:"count"`
	Error   string             `json:"error,omitempty"`
	Valid   bool               `json:"valid"`
}

// RegexTool 正则测试工具
type RegexTool struct{}

// NewRegexTool 创建正则测试工具
func NewRegexTool() *RegexTool {
	return &RegexTool{}
}

// TestRegex 测试正则表达式，返回所有匹配及分组信息
func (r *RegexTool) TestRegex(pattern string, flags string, input string) *RegexTestResult {
	result := &RegexTestResult{
		Pattern: pattern,
		Flags:   flags,
		Input:   input,
		Matches: []RegexMatchDetail{},
		Valid:   true,
	}

	// 构建 Go regexp 选项
	isGlobal := strings.Contains(flags, "g") || flags == ""
	isCaseInsensitive := strings.Contains(flags, "i")
	isMultiline := strings.Contains(flags, "m")

	// Go regexp 不支持全局标志，通过多次匹配模拟
	flagValue := ""
	if isCaseInsensitive {
		flagValue = "(?i)"
	}
	if isMultiline {
		flagValue += "(?m)"
	}

	re, err := regexp.Compile(flagValue + pattern)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		return result
	}

	groupNames := re.SubexpNames()

	allMatches := re.FindAllStringSubmatchIndex(input, -1)
	for matchIdx, matchIdxPair := range allMatches {
		// matchIdxPair[0], matchIdxPair[1] 是整体匹配的起止位置
		detail := RegexMatchDetail{
			Index:  matchIdx,
			Value:  input[matchIdxPair[0]:matchIdxPair[1]],
			Start:  matchIdxPair[0],
			End:    matchIdxPair[1],
			Groups: []RegexGroup{},
		}

		// matchIdxPair[i*2], matchIdxPair[i*2+1] 是第 i 组的起止位置
		// i=0 是整体匹配，i>=1 是分组
		for i := 1; i < len(matchIdxPair)/2; i++ {
			grpStart := matchIdxPair[i*2]
			grpEnd := matchIdxPair[i*2+1]
			if grpStart < 0 || grpEnd < 0 {
				// 可选分组未匹配
				continue
			}
			groupName := ""
			if i < len(groupNames) {
				groupName = groupNames[i]
			}
			detail.Groups = append(detail.Groups, RegexGroup{
				Index: i,
				Name:  groupName,
				Value: input[grpStart:grpEnd],
				Start: grpStart,
				End:   grpEnd,
			})
		}

		result.Matches = append(result.Matches, detail)

		if !isGlobal {
			break // 非全局模式只匹配第一个
		}
	}

	result.Count = len(result.Matches)

	// 检查是否有语法错误
	if _, err := regexp.Compile(pattern); err != nil {
		result.Valid = false
		result.Error = err.Error()
	}

	return result
}

// ValidateRegex 仅验证正则表达式语法是否合法
func (r *RegexTool) ValidateRegex(pattern string) error {
	_, err := regexp.Compile(pattern)
	if err != nil {
		return utils.WrapError(5060, "正则表达式语法错误", err)
	}
	return nil
}

// EscapeRegex 转义正则表达式特殊字符，返回字面量模式
func (r *RegexTool) EscapeRegex(input string) string {
	return regexp.QuoteMeta(input)
}
