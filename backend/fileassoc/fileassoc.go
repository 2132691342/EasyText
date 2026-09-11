// Package fileassoc 提供 Windows 文件关联运行时注册。
//
// 设计动机：
//   - 便携模式（无 NSIS 安装）：用户从 GitHub 下载 easy-text.exe 后，
//     双击 .log / .txt 不会被 EasyText 接管。让 EasyText 出现在「打开方式」菜单
//     至少能解决"右键 → 打开方式 → EasyText"的链路。
//   - 设置页提供"注册 / 取消"开关，用户主动控制，不静默抢默认。
//
// 写入位置（HKCU，不需要管理员权限）：
//
//	HKCU\Software\Classes\.<EXT>\OpenWithProgids\EasyText.text   = ""
//	HKCU\Software\Classes\EasyText.text                          = "EasyText 文本文件"
//	HKCU\Software\Classes\EasyText.text\DefaultIcon              = "<exe>",0
//	HKCU\Software\Classes\EasyText.text\shell\open\command       = '"<exe>" "%1"'
//
// 只动 OpenWithProgids 与 ProgID 定义本身，**绝不修改** HKCU\Software\Classes\.<EXT>
// 的默认指向，保留用户既有的默认编辑器选择。
package fileassoc

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// ProgID 标识 EasyText 注册的 ProgID，与 NSIS 安装器里的 EasyText.text 保持一致。
const progID = "EasyText.text"

// DefaultExtensions 是与 NSIS 安装器对齐的扩展名列表（NSIS 中如果调整，这里要同步）。
var DefaultExtensions = []string{
	"txt", "log", "md", "markdown",
	"json", "yaml", "yml",
	"xml", "ini",
	"csv", "tsv", "conf", "cfg", "properties", "text",
}

// exePath 返回当前进程可执行文件的绝对路径（注册表需要的）。
// 使用 os.Executable 在 Windows 下确保得到 .exe 全路径。
func exePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取 exe 路径失败：%w", err)
	}
	return p, nil
}

// Register 把一组扩展名注册到 EasyText 的 OpenWithProgids。
//
// 写入项：
//   - HKCU\Software\Classes\.<EXT>\OpenWithProgids\EasyText.text = ""
//   - HKCU\Software\Classes\EasyText.text\...                    = command / icon / description
//
// 幂等：重复调用只覆盖自己写过的值。返回注册成功的扩展名列表与遇到的错误。
func Register(exts []string) (registered []string, err error) {
	exe, err := exePath()
	if err != nil {
		return nil, err
	}
	if len(exts) == 0 {
		exts = DefaultExtensions
	}

	for _, ext := range exts {
		ext = strings.TrimPrefix(strings.TrimSpace(ext), ".")
		if ext == "" {
			continue
		}
		// 1) 写 HKCU\Software\Classes\.<ext>\OpenWithProgids\EasyText.text
		owpPath := fmt.Sprintf(`Software\Classes\.%s\OpenWithProgids`, ext)
		k, _, openErr := registry.CreateKey(registry.CURRENT_USER, owpPath, registry.SET_VALUE)
		if openErr != nil {
			return registered, fmt.Errorf("创建 %s 失败：%w", owpPath, openErr)
		}
		if err := k.SetStringValue(progID, ""); err != nil {
			_ = k.Close()
			return registered, fmt.Errorf("写 %s 失败：%w", owpPath, err)
		}
		_ = k.Close()

		// 2) 写 ProgID 定义（仅第一次写，重复注册覆盖同名值）
		progIDPath := fmt.Sprintf(`Software\Classes\%s`, progID)
		pk, _, perr := registry.CreateKey(registry.CURRENT_USER, progIDPath, registry.SET_VALUE)
		if perr != nil {
			return registered, fmt.Errorf("创建 %s 失败：%w", progIDPath, perr)
		}
		_ = pk.SetStringValue("", "EasyText 文本文件")

		// DefaultIcon
		if iconK, _, icErr := registry.CreateKey(registry.CURRENT_USER, progIDPath+`\DefaultIcon`, registry.SET_VALUE); icErr == nil {
			_ = iconK.SetStringValue("", fmt.Sprintf(`"%s",0`, exe))
			_ = iconK.Close()
		}

		// shell\open\command
		if cmdK, _, cmdErr := registry.CreateKey(registry.CURRENT_USER, progIDPath+`\shell\open\command`, registry.SET_VALUE); cmdErr == nil {
			_ = cmdK.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exe))
			_ = cmdK.Close()
		}

		_ = pk.Close()
		registered = append(registered, ext)
	}

	// 写一个内部标记，供 IsRegistered 查询用
	if flagK, _, ferr := registry.CreateKey(registry.CURRENT_USER, `Software\EasyText`, registry.SET_VALUE); ferr == nil {
		_ = flagK.SetStringValue("RuntimeAssocExts", strings.Join(registered, ","))
		_ = flagK.Close()
	}
	return registered, nil
}

// Unregister 反向清理 Register 写入的键值。失败时累计并返回。
//
// 注意：若 NSIS 安装器已在 HKCU 注册（同一 ProgID），Unregister 会一并清理，
// 这是预期行为——HKCU 是 per-user，install/portable 不应在同一用户身上共存。
func Unregister(exts []string) error {
	if len(exts) == 0 {
		exts = DefaultExtensions
	}
	var errs []string
	for _, ext := range exts {
		ext = strings.TrimPrefix(strings.TrimSpace(ext), ".")
		if ext == "" {
			continue
		}
		// 1) 删除 OpenWithProgids 下的 EasyText.text 值
		owpPath := fmt.Sprintf(`Software\Classes\.%s\OpenWithProgids`, ext)
		k, err := registry.OpenKey(registry.CURRENT_USER, owpPath, registry.SET_VALUE)
		if err == nil {
			_ = k.DeleteValue(progID)
			_ = k.Close()
		}
	}
	// 删除 ProgID 主键（仅当已无任何程序引用它才安全）。保守起见，
	// 这里只在所有 OpenWithProgids 值都被移除后才删除 ProgID 主键。
	// 因为本程序是孤立的便携应用，直接删除主键是安全的。
	if err := registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\`+progID); err != nil {
		// "not found" 不是错误
		if err != registry.ErrNotExist {
			errs = append(errs, fmt.Sprintf("删除 ProgID 失败：%v", err))
		}
	}
	// 清理运行时标记
	if err := registry.DeleteKey(registry.CURRENT_USER, `Software\EasyText`); err != nil {
		if err != registry.ErrNotExist {
			errs = append(errs, fmt.Sprintf("清理标记失败：%v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// IsRegistered 查询当前用户是否已注册（运行时标记是否存在）。
func IsRegistered() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\EasyText`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	_, _, getErr := k.GetStringValue("RuntimeAssocExts")
	_ = k.Close()
	return getErr == nil
}
