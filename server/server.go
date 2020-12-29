package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const baseDir = "/home/orestis/MyDropboxClients"

//Struct of message to communicate
type netMsg struct {
	From  string
	Rtype string
	Data  string
}

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
