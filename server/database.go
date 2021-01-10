package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
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
	}
	response, _ := createMsg("DataServer", "FilesComplete", "")
	sendMsg(conn, response)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
