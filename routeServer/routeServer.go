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

var tempFiles chan string
var tempFilesDir string

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Please specify temporary files directory")
	}
	tempFilesDir = os.Args[1]
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:4000")
	if err != nil {
		log.Fatalln("Invalid address:", os.Args[1], err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalln("Listener:", os.Args[1], err)
	}
	//Create a log file for current day
	currentDate := time.Now().Format("2006-02-1")
	logfile, err := os.OpenFile(filepath.Join("./Logfiles", currentDate), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("os.OpenFile() failed with '%s\n", err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	//Connect to the database
	dbConnect("orestis", "Ore$tis1997", "myDropbox")
	//Initialize the channel
	tempFiles = make(chan string, 10)
	//Start a go routine to handle temporary Files
	go handleTempFiles()
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
	log.Println(conn.RemoteAddr().String(), ":", username, ":Requested to get file:", filename)
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
	defer dataServerConn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fileDir := filepath.Join(tempFilesDir, username)
	err = os.MkdirAll(fileDir, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}
	recieveFile(dataServerConn, fileDir, username, filename)
	//Send the file of the path to http server
	_, file := filepath.Split(filename)
	tempFilePath := filepath.Join(fileDir, file)
	response, _ := createMsg("RouteServer", "Filepath", tempFilePath)
	sendMsg(conn, response)
	//Send the filepath to tempFiles channel so the go routine can handle it
	tempFiles <- tempFilePath
}

func sendDataServerInfo(conn net.Conn, username string) {
	userInfo, err := getUser(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println(conn.RemoteAddr().String(), ":", username, ":Requested to get DataServer information")
	serverInfo, err := getDataServer(userInfo.dataServerID)
	if err != nil {
		fmt.Println(err)
		return
	}
	networkInfo := fmt.Sprintf("%s:%s", serverInfo.ipAddr, serverInfo.listeningPort)
	response, err := createMsg("RouteServer", "DataServerInfo", networkInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	sendMsg(conn, response)
}

//Check if a file in tempDir exists for more than 5 minutes
//If true delete it
func handleTempFiles() {
	//Map with key the filepath and value the time of creation
	tempFilesMap := make(map[string]time.Time)
	for {
		select {
		case file := <-tempFiles:
			//fmt.Printf("Got a new tempfile:%s from a channel\n", file)
			tempFilesMap[file] = time.Now()
		default:
			//Range the map to see if a file exists for more than 5 minutes
			for tempFile, creationTime := range tempFilesMap {
				deleteTime := creationTime.Add(5 * time.Minute)
				timeToDelete := time.Now().After(deleteTime)
				if timeToDelete {
					log.Println("Removing file:", tempFile)
					os.Remove(tempFile)
					delete(tempFilesMap, tempFile)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}
}
