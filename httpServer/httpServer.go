package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
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
		myUser := user{}
		myUser.username = r.Form.Get("entered_username")
		myUser.email = r.Form.Get("entered_email")
		myUser.password = r.Form.Get("entered_pass")
		//t, err := template.ParseFiles("welcomeresponse.html")
		//check(err)
		//t.Execute(w, myUser)
		fmt.Println(myUser)
	}
}
