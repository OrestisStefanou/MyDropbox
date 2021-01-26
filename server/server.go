package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const baseDir = "/home/orestis/MyDropboxClients"

type fileServerEntry struct {
	procInfo *exec.Cmd
	procPort string
}

var fileServers = make(map[string]fileServerEntry)

var mu sync.RWMutex //Read write mutex to protect fileServers map from race conditions

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
	//Connect to the database
	dbConnect("orestis", "Ore$tis1997", "myDropbox")
	go handleFileServers()
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
		//fmt.Printf("netMsg struct parsed from JSON: %#v\n", request)
		switch request.Rtype {
		case "createUser":
			createUser(conn, request)
		case "UserLogin":
			userLogin(conn, request)
		case "FilesMapInit":
			sendFileInfo(conn, request)
		case "NewFile":
			//Create a new file and update the database
			createUserFile(conn, request)
		case "UpdateFile":
			updateUserFile(conn, request)
		case "FileDeleted":
			deleteUserFile(conn, request)
		case "SendUserFiles":
			sendUserFiles(conn, request)
		default:
			fmt.Println(request)
			response, _ := createMsg("DataServer", "Response", "Testing")
			sendMsg(conn, response)
		}

	}
}

//Send the listening port of dataServer for the user
func userLogin(conn net.Conn, r netMsg) {
	username := r.Data
	mu.RLock()
	fileServerInfo, prs := fileServers[username]
	if !prs {
		fmt.Println("User's data server listening port is not in the map!")
		response, err := createMsg("DataServer", "Error", "")
		if err != nil {
			fmt.Println("Problem at creating the message")
			return
		}
		sendMsg(conn, response)
	}
	mu.RUnlock()
	fileServerPort := fileServerInfo.procPort
	//Respond to the http server the dataServer Port
	response, err := createMsg("DataServer", "OK", fileServerPort)
	if err != nil {
		fmt.Println("Problem at creating the message")
		return
	}
	sendMsg(conn, response)
}

//Send the files info of the user to desktop client
func sendFileInfo(conn net.Conn, r netMsg) {
	username := r.Data
	fmt.Println(username)
	sendUserFilesInfo(username, conn)
}

func createUser(conn net.Conn, r netMsg) {
	//Create a folder for the user
	username := r.Data
	userDir := filepath.Join(baseDir, username) //Data only contains the username in this case
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
	err = createZip(userDir, username)
	if err != nil {
		response, err := createMsg("DataServer", "ERROR", "")
		if err != nil {
			fmt.Println("Problem at creating the message")
			return
		}
		sendMsg(conn, response)
		return
	}
	//Get the port of fileServer for this user
	var fileServerPort string
	var fileServerInfo fileServerEntry
	var prs bool
	for { //Run this loop until the goroutine that handles the fileServers starts the fileServer
		mu.RLock()
		fileServerInfo, prs = fileServers[username]
		if prs {
			mu.RUnlock()
			break
		}
		mu.RUnlock()
		time.Sleep(1 * time.Second)
	}
	fileServerPort = fileServerInfo.procPort
	fmt.Println("File server port is ", fileServerPort)
	//Respond to the http server that directory created
	response, err := createMsg("DataServer", "OK", fileServerPort)
	if err != nil {
		fmt.Println("Problem at creating the message")
		return
	}
	sendMsg(conn, response)
}

//Create the zip file that contains the neccessary files to install the app
func createZip(userDir, username string) error {
	//Create the conf file
	f, err := os.Create("client.conf")
	if err != nil {
		return err
	}
	_, err = f.WriteString(username + "\n")
	if err != nil {
		return err
	}
	f.Sync()
	f.Close()

	//Create the zip file
	filesToZip := []string{"client.conf", "daemon"}
	output := filepath.Join(userDir, "myDropboxApp")
	if err := ZipFiles(output, filesToZip); err != nil {
		return err
	}
	//Now that we create the zip file we remove the conf file
	os.Remove("client.conf")
	return nil
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func handleFileServers() {
	port := 3000
	for {
		time.Sleep(1 * time.Second)
		c, err := ioutil.ReadDir(baseDir)
		if err != nil {
			log.Println("Error during reading the base directory")
			return
		}
		mu.Lock()
		for _, entry := range c {
			_, prs := fileServers[entry.Name()]
			if prs { //File server is already running
				continue
			} else {
				//Start the file server for this directory
				fileServerPath := filepath.Join(baseDir, entry.Name())
				fmt.Println(fileServerPath, port)
				cmd := exec.Command("./fileServer", fileServerPath, fmt.Sprint(port))
				cmd.Start()
				serverEntry := fileServerEntry{cmd, fmt.Sprint(port)}
				fileServers[entry.Name()] = serverEntry
				port++
			}
		}
		mu.Unlock()
	}
}
