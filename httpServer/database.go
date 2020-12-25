package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type user struct {
	username string
	email    string
	password string
}

var db *sql.DB

func dbConnect(dbUser, dbPass, dbName string) {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", dbUser, dbPass, dbName))
	checkErr(err)

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}

//Insert a new user in the database
func createUser(userInfo user, dataServerID int) {
	stmt, err := db.Prepare("INSERT Users SET Username=?,Email=?,Password=?,DataServerId=?")
	checkErr(err)

	_, err = stmt.Exec(userInfo.username, userInfo.email, userInfo.password, dataServerID)
	checkErr(err)
}

//Get a user from the database.If a user with this username doesn't exist
//the user.Username will be an empty string
func getUser(username string) user {
	rows, err := db.Query("SELECT * FROM Users WHERE username=?", username)
	checkErr(err)

	userInfo := user{}
	userInfo.username = ""

	for rows.Next() {
		err = rows.Scan(&userInfo.username, &userInfo.email, &userInfo.password)
		checkErr(err)
	}
	return userInfo
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
