package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
)

func main() {
	dbConnect("orestis", "Ore$tis1997", "myDropbox")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/signup", signUpHandler)
	http.ListenAndServe(":8080", nil)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("index.html")
		check(err)
		t.Execute(w, nil)
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

//Struct of message to communicate
type netMsg struct {
	From  string
	Rtype string
	Data  string
}

//Creates a message in JSON format
func createMsg(from, rtype, data string) ([]byte, error) {
	r := netMsg{
		From:  from,
		Rtype: rtype,
		Data:  data,
	}
	d, err := json.Marshal(&r)
	if err != nil {
		returnErr := fmt.Errorf("Error during json encoding")
		return d, returnErr
	}
	d = append(d, "\n"...)
	return d, nil
}

//Receive a message from a socket
func getMsg(conn net.Conn) (netMsg, error) {
	jsonResponse := netMsg{}
	r := bufio.NewReader(conn)
	response, err := r.ReadString('\n')
	if err != nil {
		return jsonResponse, err
	}
	response = strings.TrimSpace(response)
	err = json.Unmarshal([]byte(response), &jsonResponse)
	if err != nil {
		return jsonResponse, err
	}
	return jsonResponse, nil
}

//Send a message through a socket
func sendMsg(conn net.Conn, msg []byte) {
	msgLen := len(msg)
	totalBytesSent, err := conn.Write(msg)
	if err != nil {
		log.Println("-> Connection:", err)
		return
	}
	for totalBytesSent < msgLen {
		bytesSent, err := conn.Write(msg[totalBytesSent:])
		if err != nil {
			log.Println("-> Connection:", err)
			return
		}
		totalBytesSent += bytesSent
	}
}
