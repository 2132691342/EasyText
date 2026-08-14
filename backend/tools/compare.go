package tools

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"easy-text/backend/utils"
)

// DirCompareEntry 表示目录对比中单个相对路径条目的差异分类
type DirCompareEntry struct {
	RelPath    string `json:"relPath"`
	LeftOnly   bool   `json:"leftOnly"`
	RightOnly  bool   `json:"rightOnly"`
	Identical  bool   `json:"identical"`
	Different  bool   `json:"different"`
	LeftSize   int64  `json:"leftSize"`
	RightSize  int64  `json:"rightSize"`
	LeftMtime  string `json:"leftMtime"`
	RightMtime string `json:"rightMtime"`
}

// DirCompareResult 目录对比结果
type DirCompareResult struct {
	LeftBase  string            `json:"leftBase"`
	RightBase string            `json:"rightBase"`
	Entries   []DirCompareEntry `json:"entries"`
}

// BinCompareResult 二进制对比结果
type BinCompareResult struct {
	Equal           bool   `json:"equal"`
	FirstDiffOffset int64  `json:"firstDiffOffset"`
	Reason          string `json:"reason"`
	HexWindow       string `json:"hexWindow"`
}

// CompareService 目录与二进制对比服务
type CompareService struct{}

// NewCompareService 创建对比服务实例
func NewCompareService() *CompareService { return &CompareService{} }

// CompareDirectories 递归对比两个目录，按相对路径并集分类每个条目
func (s *CompareService) CompareDirectories(leftDir, rightDir string) (*DirCompareResult, error) {
	if info, err := os.Stat(leftDir); err != nil || !info.IsDir() {
		return nil, utils.NewAppError(6001, "左侧目录不存在", leftDir)
	}
	if info, err := os.Stat(rightDir); err != nil || !info.IsDir() {
		return nil, utils.NewAppError(6001, "右侧目录不存在", rightDir)
	}

	leftFiles, err := collectFiles(leftDir)
	if err != nil {
		return nil, utils.WrapError(6001, "遍历左侧目录失败", err)
	}
	rightFiles, err := collectFiles(rightDir)
	if err != nil {
		return nil, utils.WrapError(6001, "遍历右侧目录失败", err)
	}

	// 并集键
	keys := make(map[string]struct{})
	for k := range leftFiles {
		keys[k] = struct{}{}
	}
	for k := range rightFiles {
		keys[k] = struct{}{}
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	entries := make([]DirCompareEntry, 0, len(sorted))
	for _, rel := range sorted {
		l, hasL := leftFiles[rel]
		r, hasR := rightFiles[rel]
		e := DirCompareEntry{RelPath: rel}
		if hasL {
			e.LeftSize = l.size
			e.LeftMtime = l.mtime
		}
		if hasR {
			e.RightSize = r.size
			e.RightMtime = r.mtime
		}
		switch {
		case hasL && !hasR:
			e.LeftOnly = true
		case !hasL && hasR:
			e.RightOnly = true
		case l.size == r.size && l.mtime == r.mtime:
			e.Identical = true
		default:
			e.Different = true
		}
		entries = append(entries, e)
	}
	return &DirCompareResult{LeftBase: leftDir, RightBase: rightDir, Entries: entries}, nil
}

type fileMeta struct {
	size  int64
	mtime string
}

func collectFiles(base string) (map[string]fileMeta, error) {
	out := make(map[string]fileMeta)
	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		out[rel] = fileMeta{size: info.Size(), mtime: info.ModTime().Format(time.RFC3339)}
		return nil
	})
	return out, err
}

// BinaryCompare 逐字节对比两个文件，返回首个差异偏移与周围 hex 窗口
func (s *CompareService) BinaryCompare(leftPath, rightPath string) (*BinCompareResult, error) {
	lStat, err := os.Stat(leftPath)
	if err != nil {
		return nil, utils.WrapError(6002, "左侧文件不存在", err)
	}
	rStat, err := os.Stat(rightPath)
	if err != nil {
		return nil, utils.WrapError(6002, "右侧文件不存在", err)
	}
	if lStat.Size() > 100*1024*1024 || rStat.Size() > 100*1024*1024 {
		return nil, utils.NewAppError(1003, "文件过大，无法进行二进制对比", "")
	}

	lf, err := os.Open(leftPath)
	if err != nil {
		return nil, utils.WrapError(6002, "打开左侧文件失败", err)
	}
	defer lf.Close()
	rf, err := os.Open(rightPath)
	if err != nil {
		return nil, utils.WrapError(6002, "打开右侧文件失败", err)
	}
	defer rf.Close()

	const chunk = 64 * 1024
	lBuf := make([]byte, chunk)
	rBuf := make([]byte, chunk)
	var offset int64
	for {
		nL, errL := io.ReadFull(lf, lBuf)
		nR, errR := io.ReadFull(rf, rBuf)
		minN := min(nL, nR)
		if diff := bytes.Compare(lBuf[:minN], rBuf[:minN]); diff != 0 {
			for i := 0; i < minN; i++ {
				if lBuf[i] != rBuf[i] {
					absOff := offset + int64(i)
					return &BinCompareResult{
						Equal:           false,
						FirstDiffOffset: absOff,
						Reason:          "内容不一致",
						HexWindow:       hexWindow(lf, rf, absOff),
					}, nil
				}
			}
		}
		// 一方提前结束
		if (errL == io.EOF || errL == io.ErrUnexpectedEOF) && nL < nR {
			return &BinCompareResult{Equal: false, FirstDiffOffset: offset + int64(nL), Reason: "左侧文件较短"}, nil
		}
		if (errR == io.EOF || errR == io.ErrUnexpectedEOF) && nR < nL {
			return &BinCompareResult{Equal: false, FirstDiffOffset: offset + int64(nR), Reason: "右侧文件较短"}, nil
		}
		if (errL == io.EOF || errL == io.ErrUnexpectedEOF) && (errR == io.EOF || errR == io.ErrUnexpectedEOF) {
			break
		}
		offset += int64(minN)
	}
	return &BinCompareResult{Equal: true, FirstDiffOffset: -1, Reason: "完全相同"}, nil
}

// hexWindow 在差异偏移前后各取 16 字节生成可读 hex 窗口
func hexWindow(lf, rf *os.File, offset int64) string {
	start := offset - 16
	if start < 0 {
		start = 0
	}
	read := func(f *os.File) []byte {
		buf := make([]byte, 32)
		_, err := f.ReadAt(buf, start)
		if err != nil && err != io.EOF {
			return nil
		}
		return buf
	}
	l := read(lf)
	r := read(rf)
	return formatHexPair(l, r, start)
}

func formatHexPair(l, r []byte, start int64) string {
	const hexd = "0123456789abcdef"
	out := "偏移        左侧                              右侧\n"
	for i := 0; i < len(l) && i < 32; i += 16 {
		row := []byte{}
		off := start + int64(i)
		row = append(row, byte('0'))
		for s := 28; s >= 0; s -= 4 {
			row = append(row, hexd[(off>>uint(s))&0xf])
		}
		row = append(row, ' ')
		for j := 0; j < 16 && i+j < len(l); j++ {
			b := l[i+j]
			row = append(row, hexd[b>>4], hexd[b&0xf], ' ')
		}
		row = append(row, ' ')
		for j := 0; j < 16 && i+j < len(r); j++ {
			b := r[i+j]
			row = append(row, hexd[b>>4], hexd[b&0xf], ' ')
		}
		row = append(row, '\n')
		out += string(row)
	}
	return out
}
