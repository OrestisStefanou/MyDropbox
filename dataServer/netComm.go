package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

//Struct of message to communicate
type netMsg struct {
	From  string
	Rtype string
	Data  string
}

type filemapEntry struct {
	Filename string
	ModTime  string
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

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}
