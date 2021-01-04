package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var m = map[string]time.Time{}

func visit(p string, info os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !info.IsDir() {
		fileModifiedTime, prs := m[p]
		if prs == false {
			fmt.Println(p, " is a new file")
			stats, _ := os.Stat(p)
			m[p] = stats.ModTime()
		} else {
			stats, _ := os.Stat(p)
			if stats.ModTime() != fileModifiedTime {
				fmt.Println("FILE ", p, " MODIFIED")
				m[p] = stats.ModTime()
			}
		}

	}
	return nil
}

func monitorFiles(dropBoxDir string) {
	filepath.Walk(dropBoxDir, visit)
	time.Sleep(5 * time.Second)
}

// ReadLines reads all lines from a file
func ReadLines(filePath string) ([]string, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	res := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		res = append(res, line)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
