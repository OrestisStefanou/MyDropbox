package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

type User struct {
	Name        string
	nationality string //unexported field.
}

type FileServerInfo struct {
	FileServerURL string
}

func main() {
	dbConnect("orestis", "Ore$tis1997", "myDropbox")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/signup", signUpHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		myUser := User{"https://gobyexample.com/http-servers", "Kipros"}
		t, err := template.ParseFiles("index.html")
		check(err)
		t.Execute(w, myUser)
	} else {

	}
}

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("signup.html")
		check(err)
		t.Execute(w, nil)
	} else {
		//Handle the signup form
		r.ParseForm()
		userInfo := user{}
		userInfo.username = r.Form.Get("entered_username")
		userInfo.email = r.Form.Get("entered_email")
		userInfo.password = r.Form.Get("entered_pass")
		//1.Check if a user with this username already exists
		result, err := getUser(userInfo.username)
		if err != nil { //User with this username does not exist
			//Get available data server info
			serverInfo := getAvailableDataServer()
			userInfo.dataServerID = serverInfo.serverID
			fmt.Println(serverInfo)
			//Send a request to create a folder for the user
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", serverInfo.ipAddr, serverInfo.listeningPort))
			if err != nil {
				log.Fatalln("Invalid address:", serverInfo.ipAddr, err)
			}
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				log.Fatalln("-> Connection:", err)
			}
			request, err := createMsg("httpServer", "createUser", userInfo.username)
			sendMsg(conn, request)
			response, err := getMsg(conn)
			fmt.Println("Response is:", response.Rtype)
			//If response is all good create a new user in the database
			if response.Rtype == "OK" {
				createUser(userInfo)
				//Send success html page
				serverURL := fmt.Sprintf("http://%s:%s/", serverInfo.ipAddr, response.Data) //Create fileServer address
				fileServer := FileServerInfo{serverURL}
				t, err := template.ParseFiles("welcome.html")
				check(err)
				fmt.Println(fileServer)
				t.Execute(w, fileServer)
			}

		} else {
			//Send error html page
			t, err := template.ParseFiles("errorPage.html")
			check(err)
			t.Execute(w, nil)
			fmt.Println(result)
		}
	}
}
