package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type Person struct {
	Name string
	Age  int    `json:"age"`
	City string `json:"city"`
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
		p := Person{
			Name: "John\n",
			Age:  37,
			City: "SF",
		}
		d, err := json.Marshal(&p)
		d = append(d, "\n"...)
		//fmt.Printf("Person in compact JSON: %s\n", string(d))
		fmt.Println(msg)
		if err != nil {
			log.Println("-> Message error:", err)
		}
		sendMsg(conn, d)
	}
}

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
