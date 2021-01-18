package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

//Structs for html pages

type FileServerInfo struct {
	FileServerURL string
}

type errorMessage struct {
	Message string
}

func main() {
	dbConnect("orestis", "Ore$tis1997", "myDropbox")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/signup", signUpHandler)
	http.HandleFunc("/signIn", signInHandler)
	http.ListenAndServe(":8080", nil)
	//Start a go routine to handle desktop client connections
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		testInfo := FileServerInfo{"https://gobyexample.com/http-servers"}
		t, err := template.ParseFiles("index.html")
		check(err)
		t.Execute(w, testInfo)
	} else {

	}
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("signIn.html")
		checkErr(err)
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		userInfo := user{}
		userInfo.username = r.Form.Get("entered_username")
		userInfo.password = r.Form.Get("entered_pass")
		//Check if a user with this username already exists
		result, err := getUser(userInfo.username)
		if err != nil {
			//Send error html page
			t, err := template.ParseFiles("errorPage.html")
			check(err)
			errorMsg := errorMessage{"Wrong username given!"}
			t.Execute(w, errorMsg)
		} else {
			if userInfo.password != result.password {
				//Send error html page
				t, err := template.ParseFiles("errorPage.html")
				check(err)
				errorMsg := errorMessage{"Wrong password given!"}
				t.Execute(w, errorMsg)
			}
			serverInfo, _ := getDataServer(result.dataServerID)
			//Send a request to dataServer to get listening port of fileServer for the user
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", serverInfo.ipAddr, serverInfo.listeningPort))
			if err != nil {
				log.Fatalln("Invalid address:", serverInfo.ipAddr, err)
			}
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				log.Fatalln("-> Connection:", err)
			}
			request, err := createMsg("httpServer", "UserLogin", userInfo.username)
			sendMsg(conn, request)
			response, err := getMsg(conn)
			fmt.Println("Response is:", response.Rtype)
			if response.Rtype == "Error" {
				//Send error html page
				t, err := template.ParseFiles("errorPage.html")
				check(err)
				errorMsg := errorMessage{"Something went wrong!"}
				t.Execute(w, errorMsg)
			}
			serverURL := fmt.Sprintf("http://%s:%s/", serverInfo.ipAddr, response.Data) //Create fileServer address
			fileServer := FileServerInfo{serverURL}
			t, err := template.ParseFiles("welcome.html")
			check(err)
			t.Execute(w, fileServer)
		}
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
			errorMsg := errorMessage{"User with this username already exists!"}
			t.Execute(w, errorMsg)
			fmt.Println(result)
		}
	}
}
