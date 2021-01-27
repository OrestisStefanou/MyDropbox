package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:4000")
	if err != nil {
		log.Fatalln("Invalid address:", os.Args[1], err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalln("Listener:", os.Args[1], err)
	}
	//Connect to the database
	dbConnect("orestis", "Ore$tis1997", "myDropbox")
	for {
		time.Sleep(time.Millisecond * 100)
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	for {
		request, err := getMsg(conn)
		if err != nil {
			if err == io.EOF {
				log.Println("<-", err)
				return
			}
			if nerr, ok := err.(net.Error); ok && !nerr.Temporary() {
				log.Println("<- Network error:", err)
				return
			}
			log.Println("<- Message error:", err)
			continue
		}
		fmt.Printf("netMsg struct parsed from JSON: %#v\n", request)
		switch request.Rtype {
		case "getDataServerInfo":
			sendDataServerInfo(conn, request.Data)
		case "GetFile":
			getFileFromDataServer(conn, request)
		}

	}
}

func getFileFromDataServer(conn net.Conn, request netMsg) {
	username := request.From
	filename := request.Data
	userInfo, err := getUser(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	serverInfo, err := getDataServer(userInfo.dataServerID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(serverInfo)
	fmt.Println(filename)
	//Connect to dataServer to receive the file
	addr := fmt.Sprintf("%s:%s", serverInfo.ipAddr, serverInfo.listeningPort)
	dataServerAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	dataServerConn, err := net.DialTCP("tcp", nil, dataServerAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	recieveFile(dataServerConn, "/home/orestis/Desktop", username, filename) //CHANGE THE PATH
	dataServerConn.Close()
	//Send the file of the path to http server
}

func sendDataServerInfo(conn net.Conn, username string) {
	userInfo, err := getUser(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	serverInfo, err := getDataServer(userInfo.dataServerID)
	if err != nil {
		fmt.Println(err)
		return
	}
	networkInfo := fmt.Sprintf("%s:%s", serverInfo.ipAddr, serverInfo.listeningPort)
	response, err := createMsg("RequestServer", "DataServerInfo", networkInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	sendMsg(conn, response)
}
