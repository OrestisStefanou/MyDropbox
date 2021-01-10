package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

var filesMap = map[string]string{} //key is the path of the file,value is modified time in string format

func initializeFilesMap(conn net.Conn, myUsername string) {
	request, err := createMsg("DesktopClient", "FilesMapInit", myUsername)
	if err != nil {
		return
	}
	sendMsg(conn, request)
	for {
		response, err := getMsg(conn)
		if response.Rtype == "FilesComplete" {
			fmt.Println("Got all the files info")
			break
		}
		var entry filemapEntry
		err = json.Unmarshal([]byte(response.Data), &entry)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(entry)
	}
}

func visit(p string, info os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !info.IsDir() {
		fileModifiedTime, prs := filesMap[p]
		if prs == false {
			fmt.Println(p, " is a new file")
			stats, _ := os.Stat(p)
			filesMap[p] = stats.ModTime().Format("2006-02-01 15:04:05.000 MST")
		} else {
			stats, _ := os.Stat(p)
			if stats.ModTime().Format("2006-02-01 15:04:05.000 MST") != fileModifiedTime {
				fmt.Println("FILE ", p, " MODIFIED")
				filesMap[p] = stats.ModTime().Format("2006-02-01 15:04:05.000 MST")
				//Send updates to dataServer
			}
		}

	}
	return nil
}

func monitorFiles(dropBoxDir string) {
	filepath.Walk(dropBoxDir, visit)
	time.Sleep(5 * time.Second)
}

func checkDeletedFiles() {
	for filePath := range filesMap {
		_, err := os.Stat(filePath)
		if err != nil { //File deleted
			fmt.Println(filePath, "deleted")
			delete(filesMap, filePath)
			//Send updates to dataServer
		}
	}
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
