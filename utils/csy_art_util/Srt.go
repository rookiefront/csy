package csy_art_util

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type SrtSubtitle struct {
	Index     int
	StartTime string
	EndTime   string
	Content   string
}

func ParseSrtFile(filename string) ([]SrtSubtitle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var subtitles []SrtSubtitle
	var currentIndex int
	var currentTime string
	var currentContent string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "00:") {
			re := regexp.MustCompile(`(\d{2}:\d{2}:\d{2},\d{3}) --> (\d{2}:\d{2}:\d{2},\d{3})`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				currentTime = matches[1]
			} else {
				return nil, fmt.Errorf("无法解析时间戳: %s", line)
			}
		} else if line != "" {
			currentContent += line + "\n"
		} else if currentContent != "" {
			subtitle := SrtSubtitle{
				Index:     currentIndex,
				StartTime: currentTime,
				EndTime:   currentTime, // 这里可能需要根据实际情况调整
				Content:   regexp.MustCompile(`^\d{0,5}\n`).ReplaceAllLiteralString(strings.TrimSpace(currentContent), ""),
			}
			subtitles = append(subtitles, subtitle)

			// 重置变量
			currentIndex++
			currentContent = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return subtitles, nil
}
