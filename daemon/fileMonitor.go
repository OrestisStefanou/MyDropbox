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

func initializeFilesMap() {
	request, err := createMsg("DesktopClient", "FilesMapInit", myUsername)
	if err != nil {
		return
	}
	sendMsg(dataServerConn, request)
	for {
		response, err := getMsg(dataServerConn)
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
		file := filepath.Join(mydropboxDir, entry.Filename) //REPLACE THIS WITH myDropboxdir
		filesMap[file] = entry.ModTime
		msg, _ := createMsg("DesktopClient", "GotIt", "")
		sendMsg(dataServerConn, msg)
	}
}

func visit(p string, info os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !info.IsDir() {
		fileModifiedTime, prs := filesMap[p]
		if prs == false {
			rel, err := filepath.Rel(mydropboxDir, p) //Replace this with myDropboxDir
			if err != nil {
				panic(err)
			}
			fmt.Println(rel, " is a new file")
			stats, _ := os.Stat(p)
			modTime := stats.ModTime().Format("2006-02-01 15:04:05.000 MST")
			filesMap[p] = modTime
			//Send the new file to the dataServer
			fileInfo := filemapEntry{rel, modTime}
			d, err := json.Marshal(&fileInfo)
			if err != nil {
				fmt.Println(err)
				return err
			}
			request, err := createMsg(myUsername, "NewFile", string(d))
			if err != nil {
				fmt.Println(err)
				return err
			}
			sendMsg(dataServerConn, request)
			getMsg(dataServerConn)
			uploadFile(dataServerConn, p)
			getMsg(dataServerConn)
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

func uploadFile(conn net.Conn, path string) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		request, _ := createMsg(myUsername, "Line", line)
		sendMsg(conn, request)
		getMsg(conn)
	}
	request, _ := createMsg(myUsername, "Finished", "")
	sendMsg(conn, request)
	if err = scanner.Err(); err != nil {
		return
	}

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
