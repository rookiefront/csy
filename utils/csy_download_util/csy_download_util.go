package csy_download_util

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelError
	LevelNone
)

var (
	// 全局日志器，默认输出到 stdout，包含时间戳和短文件名行号
	defaultLogger = log.New(os.Stdout, "[DOWNLOAD] ", log.LstdFlags|log.Lshortfile)
	currentLevel  = LevelInfo
)

// SetLogger 替换自定义 logger
func SetLogger(logger *log.Logger) {
	defaultLogger = logger
}

// SetLogLevel 设置日志级别
func SetLogLevel(level LogLevel) {
	currentLevel = level
}

// SetLogOutput 设置日志输出位置（例如文件）
func SetLogOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

func debugf(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		defaultLogger.Output(2, "[DEBUG] "+fmt.Sprintf(format, v...))
	}
}

func infof(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		defaultLogger.Output(2, "[INFO] "+fmt.Sprintf(format, v...))
	}
}

func errorf(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		defaultLogger.Output(2, "[ERROR] "+fmt.Sprintf(format, v...))
	}
}

// ProgressStore 进度存储接口，外部实体实现 Get/Set 方法
type ProgressStore interface {
	Get() (int64, error)   // 获取已下载字节数
	Set(downloaded int64) error // 保存已下载字节数
}

// FileProgressStore 基于文件的进度存储（默认实现，兼容原有行为）
type FileProgressStore struct {
	progressFile string
}

// NewFileProgressStore 创建文件进度存储器
func NewFileProgressStore(progressFile string) *FileProgressStore {
	return &FileProgressStore{progressFile: progressFile}
}

// Get 从文件读取已下载字节数
func (f *FileProgressStore) Get() (int64, error) {
	data, err := os.ReadFile(f.progressFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // 首次下载，没有进度文件
		}
		return 0, err
	}
	var downloaded int64
	_, err = fmt.Sscanf(string(data), "%d", &downloaded)
	if err != nil {
		// 文件损坏，从0开始
		return 0, nil
	}
	return downloaded, nil
}

// Set 将已下载字节数写入文件
func (f *FileProgressStore) Set(downloaded int64) error {
	if err := os.MkdirAll(filepath.Dir(f.progressFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(f.progressFile, []byte(fmt.Sprintf("%d", downloaded)), 0644)
}

// BlockDownloader 按固定块大小下载，支持断点续传和自动重连
type BlockDownloader struct {
	url          string        // 下载 URL
	destFile     string        // 目标文件路径
	progressStore ProgressStore // 进度存储器（接口）
	blockSize    int64         // 每次请求的块大小（字节）
	taskName     string        // 任务名称，用于日志区分
	retryCount   int           // 每个块请求失败时的重试次数
	retryDelay   time.Duration // 重试间隔

	client     *http.Client
	fileSize   int64
	downloaded int64
}

// NewBlockDownloader 创建下载器，需要传入进度存储器（实体）
// 如果希望继续使用文件存储，可以传入 NewFileProgressStore(progressFile)
func NewBlockDownloader(url, destFile string, blockSize int64, store ProgressStore) *BlockDownloader {
	return &BlockDownloader{
		url:           url,
		destFile:      destFile,
		progressStore: store,
		blockSize:     blockSize,
		taskName:      filepath.Base(destFile),
		retryCount:    3,
		retryDelay:    2 * time.Second,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SetTaskName 设置任务名称（用于日志输出）
func (d *BlockDownloader) SetTaskName(name string) {
	d.taskName = name
}

// SetRetryPolicy 设置重试策略
func (d *BlockDownloader) SetRetryPolicy(retryCount int, delay time.Duration) {
	d.retryCount = retryCount
	d.retryDelay = delay
}

// getFileSize 获取文件总大小（支持重试）
func (d *BlockDownloader) getFileSize() error {
	var resp *http.Response
	var err error
	for attempt := 1; attempt <= d.retryCount; attempt++ {
		resp, err = d.client.Head(d.url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		errorf("[%s] HEAD 请求失败 (尝试 %d/%d): %v", d.taskName, attempt, d.retryCount, err)
		if attempt < d.retryCount {
			time.Sleep(d.retryDelay)
		}
	}
	if err != nil {
		return fmt.Errorf("HEAD 请求最终失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HEAD 请求状态码错误: %s", resp.Status)
	}
	d.fileSize = resp.ContentLength
	if d.fileSize <= 0 {
		return fmt.Errorf("无效的文件大小: %d", d.fileSize)
	}
	return nil
}

// loadProgress 通过存储器加载已下载字节数
func (d *BlockDownloader) loadProgress() error {
	downloaded, err := d.progressStore.Get()
	if err != nil {
		return err
	}
	d.downloaded = downloaded
	return nil
}

// saveProgress 通过存储器保存已下载字节数
func (d *BlockDownloader) saveProgress() error {
	return d.progressStore.Set(d.downloaded)
}

// downloadBlock 下载一个块（支持重试）
func (d *BlockDownloader) downloadBlock(ctx context.Context, start, end int64) error {
	rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
	var lastErr error

	for attempt := 1; attempt <= d.retryCount; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", d.url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Range", rangeHeader)

		resp, err := d.client.Do(req)
		if err != nil {
			lastErr = err
			errorf("[%s] 块请求失败 (尝试 %d/%d): %v", d.taskName, attempt, d.retryCount, err)
			if attempt < d.retryCount {
				time.Sleep(d.retryDelay)
			}
			continue
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status: %s", resp.Status)
			if attempt < d.retryCount {
				time.Sleep(d.retryDelay)
			}
			continue
		}

		// 打开目标文件
		dest, err := os.OpenFile(d.destFile, os.O_RDWR, 0644)
		if err != nil {
			resp.Body.Close()
			return err
		}

		// 写入数据
		writeOffset := start
		buf := make([]byte, 32*1024)
		remaining := end - start + 1
		var writeErr error
		for remaining > 0 {
			nr, er := resp.Body.Read(buf)
			if nr > 0 {
				nw, ew := dest.WriteAt(buf[:nr], writeOffset)
				if ew != nil {
					writeErr = ew
					break
				}
				if nw != nr {
					writeErr = fmt.Errorf("short write")
					break
				}
				writeOffset += int64(nw)
				remaining -= int64(nw)
				d.downloaded += int64(nw)
			}
			if er != nil {
				if er == io.EOF {
					break
				}
				writeErr = er
				break
			}
		}
		dest.Close()
		resp.Body.Close()

		if writeErr == nil && remaining == 0 {
			// 块下载成功
			return nil
		}
		if writeErr != nil {
			lastErr = writeErr
		} else {
			lastErr = fmt.Errorf("incomplete block, remaining %d", remaining)
		}
		errorf("[%s] 块写入失败 (尝试 %d/%d): %v", d.taskName, attempt, d.retryCount, lastErr)
		if attempt < d.retryCount {
			time.Sleep(d.retryDelay)
		}
	}
	return fmt.Errorf("块下载最终失败: %w", lastErr)
}

// Download 开始下载（支持断点续传和自动重连）
func (d *BlockDownloader) Download(ctx context.Context) error {
	// 1. 获取文件大小（带重试）
	if err := d.getFileSize(); err != nil {
		return err
	}
	infof("[%s] 文件总大小: %d bytes", d.taskName, d.fileSize)

	// 2. 加载已下载进度（通过实体）
	if err := d.loadProgress(); err != nil {
		return fmt.Errorf("加载进度失败: %w", err)
	}
	infof("[%s] 已下载: %d bytes, 续传起点: %d", d.taskName, d.downloaded, d.downloaded)

	if d.downloaded >= d.fileSize {
		infof("[%s] 文件已完整下载", d.taskName)
		return nil
	}

	// 3. 创建目标文件并预分配空间
	dest, err := os.OpenFile(d.destFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	if err := dest.Truncate(d.fileSize); err != nil {
		dest.Close()
		return err
	}
	dest.Close()

	// 4. 循环下载每个块
	current := d.downloaded
	for current < d.fileSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := current
		end := current + d.blockSize - 1
		if end >= d.fileSize {
			end = d.fileSize - 1
		}

		infof("[%s] 开始下载块 [%d - %d]", d.taskName, start, end)
		if err := d.downloadBlock(ctx, start, end); err != nil {
			errorf("[%s] 块下载失败: %v", d.taskName, err)
			return err
		}

		current = end + 1
		d.downloaded = current

		// 保存进度（通过实体）
		if err := d.saveProgress(); err != nil {
			errorf("[%s] 保存进度失败: %v", d.taskName, err)
		}

		percent := float64(d.downloaded) / float64(d.fileSize) * 100
		fmt.Printf("[%s] 下载进度: %.2f%% (%d/%d bytes)\n", d.taskName, percent, d.downloaded, d.fileSize)
	}

	infof("[%s] 下载完成", d.taskName)
	return nil
}

// UrlToProgressFile 根据 URL 生成进度文件路径（使用 MD5 确保唯一性）
func UrlToProgressFile(url, cacheDir string) string {
	hash := md5.Sum([]byte(url))
	hashStr := hex.EncodeToString(hash[:])
	return filepath.Join(cacheDir, hashStr+".progress")
}

// SafeFileName 将 URL 转换为安全的文件名（用于展示）
func SafeFileName(url string) string {
	name := filepath.Base(url)
	if name == "" || name == "." || name == "/" {
		name = "download"
	}
	// 移除查询参数
	if idx := strings.Index(name, "?"); idx != -1 {
		name = name[:idx]
	}
	return name
}
// GetFileSize 获取文件总大小（下载前调用）
func (d *BlockDownloader) GetFileSize() (int64, error) {
	if d.fileSize == 0 {
		if err := d.getFileSize(); err != nil {
			return 0, err
		}
	}
	return d.fileSize, nil
}