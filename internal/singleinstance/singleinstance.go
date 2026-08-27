// Package singleinstance 提供 Windows 单实例检测与进程间通信。
//
// 设计动机：
//   - 当用户通过文件关联打开文件时，默认会启动一个新的进程实例，
//     导致同一应用被加载多次，浪费资源且用户体验差。
//   - 通过命名 Mutex 检测是否已有实例运行，若有则通过 IPC 将文件路径
//     传递给已有实例后退出新实例。
//
// 实现方案：
//   - 单实例检测：使用 Windows 命名 CreateMutex API
//   - 进程间通信：使用 TCP socket 监听本地端口传递文件路径
//   - 已有实例收到消息后通过 channel 通知主窗口打开文件
package singleinstance

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

const (
	// mutexName 是用于单实例检测的命名 Mutex 名称
	mutexName = "Global\\EasyText_SingleInstance_Mutex"
	// ipcPort 是用于进程间通信的本地端口
	ipcPort = 19876
	// ipcToken 是用于验证 IPC 消息的令牌
	ipcToken = "EasyText-IPC-v1"
)

// SingleInstance 管理单实例检测和 IPC
type SingleInstance struct {
	mutex     windows.Handle
	ownMutex  bool
	fileCh    chan string
	listener  net.Listener
	stopCh    chan struct{}
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex
}

// New 创建单实例管理器
// 如果已有实例运行，返回 nil 和 ErrAlreadyRunning
func New() (*SingleInstance, error) {
	// 尝试创建命名 mutex
	mutex, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(mutexName))

	// 检查是否因为 mutex 已存在而失败
	if err == windows.ERROR_ALREADY_EXISTS {
		// 已有实例在运行，关闭句柄并返回错误
		if mutex != 0 {
			windows.CloseHandle(mutex)
		}
		return nil, ErrAlreadyRunning
	}

	// 其他错误
	if err != nil {
		return nil, fmt.Errorf("创建单实例 Mutex 失败: %w", err)
	}

	// 成功创建 mutex，说明没有其他实例在运行
	return &SingleInstance{
		mutex:    mutex,
		ownMutex: true,
		fileCh:   make(chan string, 10),
		stopCh:   make(chan struct{}),
	}, nil
}

// ErrAlreadyRunning 表示已有实例在运行
var ErrAlreadyRunning = fmt.Errorf("EasyText 已在运行")

// FileChannel 返回文件路径接收 channel
func (si *SingleInstance) FileChannel() <-chan string {
	return si.fileCh
}

// IsRunning 检查是否已有实例在运行（静态方法）
func IsRunning() bool {
	_, err := windows.OpenMutex(windows.MUTEX_ALL_ACCESS, false, windows.StringToUTF16Ptr(mutexName))
	if err != nil {
		return false
	}
	return true
}

// SendFileToRunningInstance 将文件路径发送给已有实例
// 通过 TCP socket 连接到本地端口发送文件路径
func SendFileToRunningInstance(filePath string) error {
	// 连接到已有实例
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", ipcPort), 2*time.Second)
	if err != nil {
		return fmt.Errorf("连接已有实例失败: %w", err)
	}
	defer conn.Close()

	// 设置超时
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// 发送消息格式: token\nfilePath\n
	msg := fmt.Sprintf("%s\n%s\n", ipcToken, filePath)
	_, err = conn.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("发送文件路径失败: %w", err)
	}

	return nil
}

// StartListener 启动 IPC 监听器
// 在本地端口监听，接收来自其他实例的文件路径
func (si *SingleInstance) StartListener() error {
	si.mu.Lock()
	defer si.mu.Unlock()

	if si.isRunning {
		return nil
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", ipcPort))
	if err != nil {
		return fmt.Errorf("启动 IPC 监听器失败: %w", err)
	}

	si.listener = listener
	si.isRunning = true

	si.wg.Add(1)
	go si.acceptLoop()

	return nil
}

// Stop 停止监听并释放资源
func (si *SingleInstance) Stop() {
	close(si.stopCh)

	if si.listener != nil {
		si.listener.Close()
	}

	si.wg.Wait()

	if si.ownMutex && si.mutex != 0 {
		windows.CloseHandle(si.mutex)
		si.mutex = 0
		si.ownMutex = false
	}
}

// acceptLoop 接受连接循环
func (si *SingleInstance) acceptLoop() {
	defer si.wg.Done()

	for {
		select {
		case <-si.stopCh:
			return
		default:
		}

		// 设置接受超时，以便定期检查 stopCh
		si.listener.(*net.TCPListener).SetDeadline(time.Now().Add(500 * time.Millisecond))

		conn, err := si.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue // 超时，继续循环检查 stopCh
			}
			// 其他错误（如监听器已关闭），退出循环
			return
		}

		// 处理连接
		si.wg.Add(1)
		go func() {
			defer si.wg.Done()
			si.handleConnection(conn)
		}()
	}
}

// handleConnection 处理单个连接
func (si *SingleInstance) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 读取消息
	reader := bufio.NewReader(conn)
	token, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	// 验证 token
	if token != ipcToken+"\n" {
		return
	}

	// 读取文件路径
	filePath, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	// 移除换行符
	if len(filePath) > 0 && filePath[len(filePath)-1] == '\n' {
		filePath = filePath[:len(filePath)-1]
	}

	if filePath != "" {
		select {
		case si.fileCh <- filePath:
		case <-time.After(time.Second):
			// channel 满了，丢弃
		}
	}
}

// Release 释放 mutex（用于测试）
func (si *SingleInstance) Release() {
	if si.ownMutex && si.mutex != 0 {
		windows.CloseHandle(si.mutex)
		si.mutex = 0
		si.ownMutex = false
	}
}
