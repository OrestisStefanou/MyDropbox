package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	initdFile    = "/etc/init.d/mydropboxClient"
	varDir       = "/var/mydropboxClient/"
	pidFile      = "mydropboxClient.pid"
	outFile      = "mydropboxClient.log"
	errFile      = "mydropboxClient.err"
	initdContent = `#!/bin/sh

### BEGIN INIT INFO
# Provides:          mydropboxClient
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Custom daemon
# Description:       Enable service provided by daemon.
### END INIT INFO

"%s" $1
`
)

// ErrSudo is an error that suggest to execute the command as super user
var ErrSudo error

var (
	bin string
	cmd string
)

func init() {
	p, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	bin = p
	if len(os.Args) != 1 {
		cmd = os.Args[1]
	}
	ErrSudo = fmt.Errorf("try `sudo %s %s`", bin, cmd)
}

func main() {
	var err error
	switch cmd {
	case "run":
		err = runApp()
	case "install":
		err = installApp()
	case "uninstall":
		err = uninstallApp()
	case "status":
		err = statusApp()
	case "start":
		err = startApp()
	case "stop":
		err = stopApp()
	default:
		helpApp()
	}
	if err != nil {
		fmt.Println(cmd, "error:", err)
	}
}

func installApp() error {
	const perm = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if err := os.MkdirAll(varDir, 0755); err != nil {
		if !os.IsPermission(err) {
			return err
		}
		return ErrSudo
	}
	//Get the name of the current user to create the path of myDropbox dir
	wd, err := os.Getwd()
	mydropboxDir := filepath.Join(string(wd), "myDropbox")
	//Create the dir
	if err := os.MkdirAll(mydropboxDir, 0755); err != nil {
		if !os.IsPermission(err) {
			return err
		}
		return ErrSudo
	}
	//Write the path of myDropbox dir to the conf file
	f, err := os.OpenFile("client.conf", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("os.Open failed with '%s'\n", err)
	}
	_, err = f.Write([]byte(mydropboxDir))
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	//Copy the configuration file to varDir
	//Open the original conf file
	original, err := os.Open("client.conf")
	if err != nil {
		log.Fatal(err)
	}
	//Create the new file
	newFilePath := filepath.Join(varDir, "client.conf")
	newFile, err := os.Create(newFilePath)
	if err != nil {
		log.Fatal(err)
	}
	//Copy the file
	_, err = io.Copy(newFile, original)
	if err != nil {
		log.Fatal(err)
	}
	newFile.Close()
	original.Close()

	_, err = os.Stat(initdFile)
	if err == nil {
		return errors.New("Already installed")
	}
	f, err = os.OpenFile(initdFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		if !os.IsPermission(err) {
			return err
		}
		return ErrSudo
	}
	defer f.Close()
	if _, err = fmt.Fprintf(f, initdContent, bin); err != nil {
		return err
	}
	fmt.Println("MydropboxClient", bin, "installed")
	return nil
}

func uninstallApp() error {
	_, err := os.Stat(initdFile)
	if err != nil && os.IsNotExist(err) {
		return errors.New("not installed")
	}
	if err = os.Remove(initdFile); err != nil {
		if err != nil {
			if !os.IsPermission(err) {
				return err
			}
			return ErrSudo
		}
	}
	fmt.Println("MydropboxClient", bin, "removed")
	return err
}

func statusApp() (err error) {
	var pid int
	defer func() {
		if pid == 0 {
			fmt.Println("status: not active")
			return
		}
		fmt.Println("status: active - pid", pid)
	}()
	pid, err = getPid()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if err = p.Signal(syscall.Signal(0)); err != nil {
		fmt.Println(pid, "not found - removing PID file...")
		os.Remove(filepath.Join(varDir, pidFile))
		pid = 0
	}
	return nil
}

func writePid(pid int) (err error) {
	f, err := os.OpenFile(filepath.Join(varDir, pidFile), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = fmt.Fprintf(f, "%d", pid); err != nil {
		return err
	}
	return nil
}

func getPid() (pid int, err error) {
	b, err := ioutil.ReadFile(filepath.Join(varDir, pidFile))
	if err != nil {
		return 0, err
	}
	if pid, err = strconv.Atoi(string(b)); err != nil {
		return 0, fmt.Errorf("Invalid PID value: %s", string(b))
	}
	return pid, nil
}

func startApp() (err error) {
	const perm = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	cmd := exec.Command(bin, "run")
	cmd.Stdout, err = os.OpenFile(filepath.Join(varDir, outFile), perm, 0644)
	if err != nil {
		return err
	}
	cmd.Stderr, err = os.OpenFile(filepath.Join(varDir, errFile), perm, 0644)
	if err != nil {
		return err
	}
	cmd.Dir = "/"
	if err = cmd.Start(); err != nil {
		return err
	}
	if err := writePid(cmd.Process.Pid); err != nil {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Cannot kill process", cmd.Process.Pid, err)
		}
		return err
	}
	fmt.Println("Started with PID", cmd.Process.Pid)
	return nil
}

func stopApp() (err error) {
	pid, err := getPid()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if err = p.Signal(os.Kill); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(varDir, pidFile)); err != nil {
		return err
	}
	fmt.Println("Stopped PID", pid)
	return nil
}

func runApp() error {
	//3.Get the file info from dataServer to initialize fileMap or load it from a file localy
	//4.Check for updates and send them to the dataServer
	fmt.Println("RUNNING")
	//Read the username and myDropbox dir from conf file
	confFilePath := filepath.Join(varDir, "client.conf")
	lines, err := ReadLines(confFilePath)
	if err != nil {
		log.Fatal(err)
	}
	myUsername := lines[0]
	mydropboxDir := lines[1]
	fmt.Println(mydropboxDir)
	var dataServerInfo string
	//Run the loop until we connect to the server(In case of no internet try until there is internet connection)
	for {
		//Connect to router server to get the info of the dataServer
		addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:4000")
		if err != nil {
			fmt.Println("PROBLEM WITH THE IP AND PORT PROVIDED")
			os.Exit(1)
		}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			//fmt.Println(err)
			continue
		}
		//Send a request to get the ip and port of dataServer with our remote directory
		req, err := createMsg("DesktopClient", "getDataServerInfo", myUsername)
		sendMsg(conn, req)
		response, err := getMsg(conn)
		if err != nil {
			//fmt.Println(err)
			continue
		}
		fmt.Println(response)
		dataServerInfo = response.Data
		conn.Close()
		break
	}
	//Connect to dataServer
	dataServerAddr, err := net.ResolveTCPAddr("tcp", dataServerInfo)
	if err != nil {
		log.Fatal(err)
	}
	dataServerConn, err := net.DialTCP("tcp", nil, dataServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	//Test message
	type mapEntry struct {
		Filename string
		ModTime  string
	}
	testData := mapEntry{
		Filename: "hello.go",
		ModTime:  "09-01-2021",
	}
	d, err := json.Marshal(&testData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(d))
	request, _ := createMsg("DesktopClient", "Testing", string(d))
	sendMsg(dataServerConn, request)
	///////

	for {
		monitorFiles("/home/orestis/Downloads") //Change this to myDropboxDir
		checkDeletedFiles()
	}
	//return nil
}

func helpApp() error {
	fmt.Println("usage:", bin, "install|uninstall|status|start|stop")
	return nil
}
