package api

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/sha3"
)

// HashResult 哈希计算结果
type HashResult struct {
	MD5    string `json:"md5"`
	SHA1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
}

// ComputeHash 计算内容的 MD5/SHA1/SHA256 哈希值
func (h *Handler) ComputeHash(content string) *HashResult {
	data := []byte(content)

	md5Hash := md5.Sum(data)
	sha1Hash := sha1.Sum(data)
	sha256Hash := sha256.Sum256(data)

	return &HashResult{
		MD5:    hex.EncodeToString(md5Hash[:]),
		SHA1:   hex.EncodeToString(sha1Hash[:]),
		SHA256: hex.EncodeToString(sha256Hash[:]),
	}
}

// ComputeFileHash 计算文件的哈希值
func (h *Handler) ComputeFileHash(filePath string) (*HashResult, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// HashAlgoResult 单算法哈希结果
type HashAlgoResult struct {
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
	Error     string `json:"error,omitempty"`
}

// ComputeHashWithAlgo 按指定算法计算哈希
// 支持：md4 / md5 / sha1 / sha256 / sha3_256 / keccak_256
func (h *Handler) ComputeHashWithAlgo(content string, algorithm string) *HashAlgoResult {
	data := []byte(content)
	result := &HashAlgoResult{Algorithm: algorithm}

	switch algorithm {
	case "md4":
		hh := md4.New()
		hh.Write(data)
		result.Hash = hex.EncodeToString(hh.Sum(nil))
	case "md5":
		hh := md5.Sum(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "sha1":
		hh := sha1.Sum(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "sha256":
		hh := sha256.Sum256(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "sha3_256":
		hh := sha3.Sum256(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "keccak_256":
		hh := sha3.NewLegacyKeccak256()
		hh.Write(data)
		result.Hash = hex.EncodeToString(hh.Sum(nil))
	default:
		result.Error = fmt.Sprintf("不支持的算法: %s", algorithm)
	}

	return result
}

// ComputeFileHashWithAlgo 按指定算法计算文件哈希
func (h *Handler) ComputeFileHashWithAlgo(filePath string, algorithm string) *HashAlgoResult {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &HashAlgoResult{Algorithm: algorithm, Error: err.Error()}
	}
	result := &HashAlgoResult{Algorithm: algorithm}

	switch algorithm {
	case "md4":
		hh := md4.New()
		hh.Write(data)
		result.Hash = hex.EncodeToString(hh.Sum(nil))
	case "md5":
		hh := md5.Sum(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "sha1":
		hh := sha1.Sum(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "sha256":
		hh := sha256.Sum256(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "sha3_256":
		hh := sha3.Sum256(data)
		result.Hash = hex.EncodeToString(hh[:])
	case "keccak_256":
		hh := sha3.NewLegacyKeccak256()
		hh.Write(data)
		result.Hash = hex.EncodeToString(hh.Sum(nil))
	default:
		result.Error = fmt.Sprintf("不支持的算法: %s", algorithm)
	}

	return result
}
