package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

const baseDir = "/home/orestis/MyDropboxClients"

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Please specify an address.")
	}
	addr, err := net.ResolveTCPAddr("tcp", os.Args[1])
	if err != nil {
		log.Fatalln("Invalid address:", os.Args[1], err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalln("Listener:", os.Args[1], err)
	}
	for {
		time.Sleep(time.Millisecond * 100)
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatalln("<- Accept:", os.Args[1], err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	time.Sleep(time.Second / 2)
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
		case "createUser":
			createUser(conn, request)
		}

	}
}

func createUser(conn net.Conn, r netMsg) {
	//Create a folder for the user
	userDir := filepath.Join(baseDir, r.Data) //Data only contains the username in this case
	err := os.Mkdir(userDir, 0755)
	if err != nil {
		fmt.Println("Something went wrong creating the directory")
		response, err := createMsg("DataServer", "ERROR", "")
		if err != nil {
			fmt.Println("Problem at creating the message")
			return
		}
		sendMsg(conn, response)
		return
	}
	//Respond to the http server that directory created
	response, err := createMsg("DataServer", "OK", "")
	if err != nil {
		fmt.Println("Problem at creating the message")
		return
	}
	sendMsg(conn, response)
}
