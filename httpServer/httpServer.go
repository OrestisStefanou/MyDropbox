package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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
		//r.ParseForm()
		//m/yUser := user{}
		//myUser.Name = r.Form.Get("entered_name")
		//myUser.nationality = r.Form.Get("entered_nationality")
		//t, err := template.ParseFiles("welcomeresponse.html")
		//check(err)
		//t.Execute(w, myUser)
		//fmt.Println(myUser)
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
		//t, err := template.ParseFiles("welcomeresponse.html")
		//check(err)
		//t.Execute(w, myUser)
		fmt.Println(userInfo)
		//1.Check if a user with this username already exists
		result, err := getUser(userInfo.username)
		if err != nil { //User with this username does not exist
			//Get available data server info
			fmt.Println(err)
			serverInfo := getAvailableDataServer()
			fmt.Println(serverInfo)
			//Send a request to create a folder for the user
			//If response is all good create a new user in the database

		} else {
			//Send error html page
			t, err := template.ParseFiles("errorPage.html")
			check(err)
			t.Execute(w, nil)
			fmt.Println(result)
		}
	}
}
