package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

//Connect to the database
func dbConnect(dbUser, dbPass, dbName string) {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", dbUser, dbPass, dbName))
	checkErr(err)

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}

//Get the files of a user
func sendUserFilesInfo(username string, conn net.Conn) {
	rows, err := db.Query("SELECT Filepath,LastModified FROM Files WHERE Owner=?", username)
	if err != nil {
		fmt.Println("Error during sending the files info to the user")
		return
	}
	fileInfo := filemapEntry{}
	for rows.Next() {
		err = rows.Scan(&fileInfo.Filename, &fileInfo.ModTime)
		if err != nil {
			fmt.Println(err)
			return
		}
		//Send the info to desktopClient
		d, err := json.Marshal(&fileInfo)
		if err != nil {
			fmt.Println(err)
			return
		}
		response, err := createMsg("DataServer", "FileMapEntry", string(d))
		sendMsg(conn, response)
		getMsg(conn)
	}
	response, _ := createMsg("DataServer", "FilesComplete", "")
	sendMsg(conn, response)
}

//Create a new user's file
func createUserFile(conn net.Conn, request netMsg) {
	username := request.From
	var fileInfo filemapEntry
	err := json.Unmarshal([]byte(request.Data), &fileInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	dir, file := filepath.Split(fileInfo.Filename) //Get the parents
	if dir != "" {
		newDir := filepath.Join(baseDir, username, dir)
		//Create the parent directories
		err := os.MkdirAll(newDir, 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
		newFilePath := filepath.Join(newDir, file)
		fmt.Println(newFilePath)
		getFile(conn, newFilePath)
		stmt, err := db.Prepare("Insert Files SET Filepath=?,Owner=?,LastModified=?")
		checkErr(err)
		_, err = stmt.Exec(fileInfo.Filename, username, fileInfo.ModTime)
		checkErr(err)
	} else {
		newFilePath := filepath.Join(baseDir, username, file)
		fmt.Println(newFilePath)
		getFile(conn, newFilePath)
		stmt, err := db.Prepare("Insert Files SET Filepath=?,Owner=?,LastModified=?")
		checkErr(err)
		_, err = stmt.Exec(fileInfo.Filename, username, fileInfo.ModTime)
		checkErr(err)
	}
	//Send a message that file created
	response, err := createMsg("DataServer", "OK", "")
	sendMsg(conn, response)
}

func updateUserFile(conn net.Conn, request netMsg) {
	username := request.From
	var fileInfo filemapEntry
	err := json.Unmarshal([]byte(request.Data), &fileInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	path := filepath.Join(baseDir, username, fileInfo.Filename)
	fmt.Println("Updating ", path)
	getFile(conn, path)
	stmt, err := db.Prepare("UPDATE Files set LastModified=? WHERE Filepath=? AND Owner=?")
	checkErr(err)
	_, err = stmt.Exec(fileInfo.ModTime, fileInfo.Filename, username)
	checkErr(err)
	response, err := createMsg("DataServer", "OK", "")
	sendMsg(conn, response)
}

func deleteUserFile(conn net.Conn, request netMsg) {
	username := request.From
	filename := request.Data
	path := filepath.Join(baseDir, username, filename)
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("os.Remove failed because file doesn't exist\n")
		} else {
			fmt.Printf("os.Remove failed with '%s'\n", err)
		}
		return
	}
	dir, _ := filepath.Split(path)
	os.Remove(dir) //Delete the directory of the file if is empty

	stmt, err := db.Prepare("DELETE FROM Files WHERE filepath=? AND Owner=?")
	checkErr(err)

	_, err = stmt.Exec(filename, username)
	checkErr(err)
	response, err := createMsg("DataServer", "OK", "")
	sendMsg(conn, response)
}

func getFile(conn net.Conn, path string) {
	f, err := os.Create(path)
	check(err)
	defer f.Close()
	response, err := createMsg("DataServer", "SendFile", "")
	sendMsg(conn, response)
	for {
		msg, err := getMsg(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		if msg.Rtype == "Finished" {
			break
		}
		fileLine := msg.Data
		_, err = f.WriteString(fileLine)
		_, err = f.WriteString("\n")
		//Send a message that we got the line to send the next one
		response, _ := createMsg("DataServer", "OK", "")
		sendMsg(conn, response)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
