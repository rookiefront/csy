package file

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type FileHandle struct {
}

func NewFile() *FileHandle {
	return &FileHandle{}
}

func (file *FileHandle) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func (file *FileHandle) ReadFileContent(filePath string) string {
	readFile, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(readFile)
}
func (file *FileHandle) FileExistsCreateDir(filename string) error {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		fmt.Println(filepath.Dir(filename), filename)
		err := os.MkdirAll(filepath.Dir(filename), 0644)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// IsFile is_file()
func (file *FileHandle) IsFile(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// IsDir is_dir()
func (file *FileHandle) IsDir(filename string) (bool, error) {
	fd, err := os.Stat(filename)
	if err != nil {
		return false, err
	}
	fm := fd.Mode()
	return fm.IsDir(), nil
}

// ISIMG
func (file *FileHandle) IsImg(filename string) (bool, error) {
	// Open File
	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return false, err
	}

	contentType := http.DetectContentType(buffer)
	log.Println(contentType)
	if !strings.HasPrefix(contentType, "im") {
		return false, nil
	}
	return true, nil
}

// Copy 文件
func (file *FileHandle) CopyFile(sourceFile, destinationFile string, markDir ...bool) error {
	if len(markDir) != 0 {
		os.MkdirAll(filepath.Dir(destinationFile), 0777)
	}
	src, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return nil
}

func (file *FileHandle) SplitFile(originalFile, targetFolder string, chunkSize uint64) (err error) {
	var splitFiles []string
	// 打开原始文件
	original, err := os.Open(originalFile)
	if err != nil {
		return fmt.Errorf("无法打开原始文件: %s", err)
	}
	defer original.Close()

	// 获取原始文件的信息
	_, err = original.Stat()
	if err != nil {
		return fmt.Errorf("无法获取原始文件信息: %s", err)
	}

	// 创建目标文件夹
	os.RemoveAll(targetFolder)
	err = os.MkdirAll(targetFolder, 0755)
	if err != nil {
		return fmt.Errorf("无法创建目标文件夹: %s", err)
	}

	// 缓冲区大小
	bufferSize := int(chunkSize)
	buffer := make([]byte, bufferSize)

	// 当前分片大小
	currentChunkSize := uint64(0)

	// 当前分片编号
	currentChunkNumber := 1

	// 获取原始文件的后缀名
	originalExt := filepath.Ext(originalFile)
	// 去掉后缀名的原始文件名
	originalName := strings.TrimSuffix(filepath.Base(originalFile), originalExt)

	// 创建新的分片文件
	chunkFileName := fmt.Sprintf("%s/%s_chunk%d%s", targetFolder, originalName, currentChunkNumber, originalExt)
	chunkFile, err := os.Create(chunkFileName)
	if err != nil {
		return fmt.Errorf("无法创建分片文件: %s", err)
	}
	defer chunkFile.Close()

	// 逐个字节进行分割
	for {
		// 从原始文件读取字节到缓冲区
		bytesRead, err := original.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("无法读取原始文件: %s", err)
		}

		// 如果没有更多字节可读，则结束循环
		if bytesRead == 0 {
			break
		}

		// 写入缓冲区的字节到当前分片文件
		_, err = chunkFile.Write(buffer[:bytesRead])
		if err != nil {
			return fmt.Errorf("无法写入分片文件: %s", err)
		}

		// 更新当前分片大小
		currentChunkSize += uint64(bytesRead)

		// 如果当前分片大小超过指定大小，则创建新的分片文件
		if currentChunkSize >= chunkSize {
			// 关闭当前分片文件
			err = chunkFile.Close()
			if err != nil {
				return fmt.Errorf("无法关闭分片文件: %s", err)
			}
			splitFiles = append(splitFiles, filepath.Base(chunkFileName))
			fmt.Printf("已创建分片文件: %s\n", chunkFileName)

			// 递增分片编号
			currentChunkNumber++

			// 重置当前分片大小
			currentChunkSize = 0

			// 创建新的分片文件
			chunkFileName = fmt.Sprintf("%s/%s_chunk%d%s", targetFolder, originalName, currentChunkNumber, originalExt)
			chunkFile, err = os.Create(chunkFileName)
			if err != nil {
				return fmt.Errorf("无法创建分片文件: %s", err)
			}
			defer chunkFile.Close()
		}
	}

	// 关闭最后一个分片文件
	err = chunkFile.Close()
	if err != nil {
		return fmt.Errorf("无法关闭分片文件: %s", err)
	}

	fmt.Printf("已创建分片文件: %s\n", chunkFileName)
	splitFiles = append(splitFiles, filepath.Base(chunkFileName))
	marshal, err := json.Marshal(splitFiles)
	if err == nil {
		os.WriteFile(filepath.Dir(chunkFileName)+"/files.json", marshal, 0755)
	}
	return nil
}

// 合并分割之后的文件
func (file *FileHandle) MergeFiles(sourceFolder, targetFile string) error {
	// 创建目标文件
	target, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("无法创建目标文件: %s", err)
	}
	defer target.Close()

	// 获取源文件夹中的所有分片文件
	chunkFiles, err := filepath.Glob(fmt.Sprintf("%s/*", sourceFolder))
	if err != nil {
		return fmt.Errorf("无法读取分片文件: %s", err)
	}

	// 按照分片文件名排序
	_sortChunkFiles(chunkFiles)

	// 逐个分片文件进行合并
	for _, chunkFile := range chunkFiles {
		// 打开分片文件
		chunk, err := os.Open(chunkFile)
		if err != nil {
			return fmt.Errorf("无法打开分片文件: %s", err)
		}
		defer chunk.Close()

		// 从分片文件复制内容到目标文件
		_, err = io.Copy(target, chunk)
		if err != nil {
			return fmt.Errorf("无法复制分片文件内容: %s", err)
		}

		fmt.Printf("已合并分片文件: %s\n", chunkFile)
	}

	fmt.Printf("已还原文件: %s\n", targetFile)

	return nil
}

// 排序分片文件，按照分片编号排序
func _sortChunkFiles(chunkFiles []string) {
	sort.Slice(chunkFiles, func(i, j int) bool {
		chunkFile1 := chunkFiles[i]
		chunkFile2 := chunkFiles[j]

		// 提取分片编号
		chunkNumber1 := _getChunkNumber(chunkFile1)
		chunkNumber2 := _getChunkNumber(chunkFile2)

		return chunkNumber1 < chunkNumber2
	})
}

// 提取分片编号
func _getChunkNumber(chunkFile string) int {
	fileName := filepath.Base(chunkFile)
	fileExt := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)

	//var chunkNumber int
	//fileNameWithoutExt = "mysql-5.7.41-winx64_chunk18"
	//_, err := fmt.Sscanf(fileNameWithoutExt, "chunk%d", &chunkNumber)
	chunkNumber, err := _extractChunkNumber(fileNameWithoutExt)
	if err != nil {
		return 0
	}

	return chunkNumber
}

func _extractChunkNumber(fileNameWithoutExt string) (int, error) {
	r := regexp.MustCompile(`chunk(\d+)`)
	matches := r.FindStringSubmatch(fileNameWithoutExt)
	if len(matches) < 2 {
		return 0, fmt.Errorf("未找到 chunkNumber")
	}

	chunkNumber := matches[1]
	var chunk int
	_, err := fmt.Sscanf(chunkNumber, "%d", &chunk)
	if err != nil {
		return 0, err
	}

	return chunk, nil
}

// 解压 zip 文件
func (file *FileHandle) UnZip(zipFilePath, targetDir string) error {
	// 打开 ZIP 文件
	zipFile, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("无法打开 ZIP 文件: %s", err)
	}
	defer zipFile.Close()

	// 创建目标文件夹
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("无法创建目标文件夹: %s", err)
	}

	// 遍历 ZIP 文件中的文件和文件夹
	for _, file := range zipFile.File {
		// 构建解压后的文件路径
		filePath := filepath.Join(targetDir, file.Name)

		if file.FileInfo().IsDir() {
			// 创建文件夹
			err = os.MkdirAll(filePath, file.Mode())
			if err != nil {
				return fmt.Errorf("无法创建文件夹: %s", err)
			}
			continue
		}

		// 创建解压后的文件
		outputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("无法创建文件: %s", err)
		}
		defer outputFile.Close()

		// 打开 ZIP 文件中的文件
		zipFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("无法打开 ZIP 文件中的文件: %s", err)
		}
		defer zipFile.Close()

		// 将 ZIP 文件中的内容复制到解压后的文件
		_, err = io.Copy(outputFile, zipFile)
		if err != nil {
			return fmt.Errorf("无法解压文件: %s", err)
		}
	}

	return nil
}

// 下载文件

func (file *FileHandle) DownloadFile(url string, outputPath string, progressCallBack func(progress int, downloadedSize, fileSize int64, downloadSpeed float64)) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// 创建输出文件
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// 创建缓冲区
	buffer := make([]byte, 1024)

	// 获取文件大小
	fileSize := response.ContentLength
	downloadedSize := int64(0)

	// 创建进度条
	progress := 0
	// 开始时间
	startTime := time.Now()

	// 定时器，每秒更新一次下载速度
	ticker := time.NewTicker(time.Second)
	defer func() {
		// 结束时间
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		downloadSpeed := float64(downloadedSize) / duration.Seconds()
		progressCallBack(progress, downloadedSize, fileSize, downloadSpeed)
		ticker.Stop()
	}()

	go func() {
		for range ticker.C {
			// 计算已下载的字节数
			currentSize := downloadedSize

			// 计算下载速度
			duration := time.Since(startTime)
			downloadSpeed := float64(currentSize) / duration.Seconds()
			progressCallBack(progress, downloadedSize, fileSize, downloadSpeed)

			//fmt.Printf("下载速度: %.2f bytes/second\n", downloadSpeed)
		}
	}()

	// 读取响应体，并写入文件
	for {
		// 读取数据
		n, err := response.Body.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		// 写入文件
		_, err = outputFile.Write(buffer[:n])
		if err != nil {
			return err
		}

		// 更新下载大小
		downloadedSize += int64(n)

		// 计算下载进度
		newProgress := int((float64(downloadedSize) / float64(fileSize)) * 100)
		if newProgress != progress {
			progress = newProgress
		}

		if downloadedSize == fileSize {
			break
		}
		// 下载完成
		if err == io.EOF {
			break
		}
	}

	return nil
}
func (f2 *FileHandle) CopyDir(src string, dest string) error {
	err := checkPathNotContained(src, dest)
	if err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}

	file, err := f.Stat()
	if err != nil {
		return err
	}

	if !file.IsDir() {
		return fmt.Errorf("Source " + file.Name() + " is not a directory!")
	}

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			if err = NewFile().CopyDir(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name())); err != nil {
				return err
			}
		} else {
			if err = NewFile().CopyFile(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkPathNotContained returns an error if 'subpath' is inside 'path'
func checkPathNotContained(path string, subpath string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	absSubPath, err := filepath.Abs(subpath)
	if err != nil {
		return err
	}

	current := absSubPath
	for {
		if current == absPath {
			return fmt.Errorf("cannot copy a folder onto itself")
		}
		up := filepath.Dir(current)
		if current == up {
			break
		}
		current = up
	}
	return nil
}
