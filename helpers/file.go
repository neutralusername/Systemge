package helpers

import (
	"fmt"
	"os"
	"strings"
)

func OpenFileAppend(filePath string) *os.File {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file : %s", err))
	}
	return file
}

func OpenFileTruncate(filePath string) *os.File {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file : %s", err))
	}
	return file
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func GetFileContent(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Sprintf("error opening file : %s", err))
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		panic(fmt.Sprintf("error getting file stat : %s", err))
	}
	size := stat.Size()
	data := make([]byte, size)
	_, err = file.Read(data)
	if err != nil {
		panic(fmt.Sprintf("error reading file : %s", err))
	}
	return string(data)
}

func GetFileName(filePath string) string {
	return filePath[strings.LastIndex(filePath, "/")+1:]
}

func SplitLines(content string) []string {
	if strings.Contains(content, "\r\n") {
		return strings.Split(content, "\r\n")
	}
	return strings.Split(content, "\n")
}

func ReplaceLine(fileContent string, lineIndex int, newLine string) string {
	lines := SplitLines(fileContent)
	lines[lineIndex] = newLine
	return strings.Join(lines, "\n")
}

func CreateDirectory(path string) {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		panic(fmt.Sprintf("error creating directory : %s", err))
	}
}

func CreateFile(filePath string, content string) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(fmt.Sprintf("error creating file : %s", err))
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		panic(fmt.Sprintf("error writing to file : %s", err))
	}
}
