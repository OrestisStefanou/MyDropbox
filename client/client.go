package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

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
	createConn(addr)
}

func createConn(addr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatalln("-> Connection:", err)
	}
	log.Println("-> Connection to", addr)
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("# ")
		msg, err := r.ReadBytes('\n')
		if err != nil {
			log.Fatal("ReadBytes error")
		}
		fmt.Println(msg)
		req, err := createMsg("httpServer", "createUser", "Orestis")
		sendMsg(conn, req)
		response, err := getMsg(conn)
		fmt.Println("Response is:", response.Rtype)
	}
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
