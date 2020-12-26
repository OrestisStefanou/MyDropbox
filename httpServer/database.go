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

type dataServerInfo struct {
	serverID      int
	maxCapacity   int
	ipAddr        string
	httpPort      string
	listeningPort string
	available     int
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

//Get the info of an available dataServer to register the user to
func getAvailableDataServer() dataServerInfo {
	found := false
	var numberOfUsers int
	rows, err := db.Query("SELECT * FROM DataServers")
	checkErr(err)

	serverInfo := dataServerInfo{}
	for rows.Next() {
		if found {
			break
		}
		err = rows.Scan(&serverInfo.serverID, &serverInfo.maxCapacity, &serverInfo.ipAddr, &serverInfo.httpPort, &serverInfo.listeningPort, &serverInfo.available)
		checkErr(err)
		rows2, err := db.Query("SELECT COUNT(Users.Username) FROM Users,DataServers WHERE Users.DataServerId = ? AND DataServers.ServerId = Users.DataServerId AND DataServers.Available = True", serverInfo.serverID)
		checkErr(err)
		for rows2.Next() {
			err = rows2.Scan(&numberOfUsers)
			if numberOfUsers < serverInfo.maxCapacity && numberOfUsers > 0 {
				found = true
				break
			}
		}
	}
	//Check if by adding this user the number of users will be equal to maxcapacity
	if numberOfUsers+1 >= serverInfo.maxCapacity {
		//If true update the availability of the server to False
		// update
		stmt, err := db.Prepare("UPDATE DataServers set Available=False where ServerId=?")
		checkErr(err)

		_, err = stmt.Exec(serverInfo.serverID)
		checkErr(err)
	}
	return serverInfo
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
