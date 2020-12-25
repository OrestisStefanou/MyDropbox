package main

import (
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
